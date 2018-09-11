package p2p

// This file contains all the default configuration options.

import (
	"crypto/rand"

	crypto "github.com/dms3-p2p/go-p2p-crypto"
	pstore "github.com/dms3-p2p/go-p2p-peerstore"
	secio "github.com/dms3-p2p/go-p2p-secio"
	tcp "github.com/dms3-p2p/go-tcp-transport"
	ws "github.com/dms3-p2p/go-ws-transport"
	multiaddr "github.com/dms3-mft/go-multiaddr"
	mplex "github.com/dms3-why/go-smux-multiplex"
	yamux "github.com/dms3-why/go-smux-yamux"
)

// DefaultSecurity is the default security option.
//
// Useful when you want to extend, but not replace, the supported transport
// security protocols.
var DefaultSecurity = Security(secio.ID, secio.New)

// DefaultMuxers configures dms3-p2p to use the stream connection multiplexers.
//
// Use this option when you want to *extend* the set of multiplexers used by
// dms3-p2p instead of replacing them.
var DefaultMuxers = ChainOptions(
	Muxer("/yamux/1.0.0", yamux.DefaultTransport),
	Muxer("/mplex/6.7.0", mplex.DefaultTransport),
)

// DefaultTransports are the default dms3-p2p transports.
//
// Use this option when you want to *extend* the set of multiplexers used by
// dms3-p2p instead of replacing them.
var DefaultTransports = ChainOptions(
	Transport(tcp.NewTCPTransport),
	Transport(ws.New),
)

// DefaultPeerstore configures dms3-p2p to use the default peerstore.
var DefaultPeerstore Option = func(cfg *Config) error {
	return cfg.Apply(Peerstore(pstore.NewPeerstore()))
}

// RandomIdentity generates a random identity (default behaviour)
var RandomIdentity = func(cfg *Config) error {
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)
	if err != nil {
		return err
	}
	return cfg.Apply(Identity(priv))
}

// DefaultListenAddrs configures dms3-p2p to use default listen address
var DefaultListenAddrs = func(cfg *Config) error {
	defaultIP4ListenAddr, err := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/0")
	if err != nil {
		return err
	}

	defaultIP6ListenAddr, err := multiaddr.NewMultiaddr("/ip6/::/tcp/0")
	if err != nil {
		return err
	}
	return cfg.Apply(ListenAddrs(
		defaultIP4ListenAddr,
		defaultIP6ListenAddr,
	))
}

// Complete list of default options and when to fallback on them.
//
// Please *DON'T* specify default options any other way. Putting this all here
// makes tracking defaults *much* easier.
var defaults = []struct {
	fallback func(cfg *Config) bool
	opt      Option
}{
	{
		fallback: func(cfg *Config) bool { return cfg.Transports == nil && cfg.ListenAddrs == nil },
		opt:      DefaultListenAddrs,
	},
	{
		fallback: func(cfg *Config) bool { return cfg.Transports == nil },
		opt:      DefaultTransports,
	},
	{
		fallback: func(cfg *Config) bool { return cfg.Muxers == nil },
		opt:      DefaultMuxers,
	},
	{
		fallback: func(cfg *Config) bool { return !cfg.Insecure && cfg.SecurityTransports == nil },
		opt:      DefaultSecurity,
	},
	{
		fallback: func(cfg *Config) bool { return cfg.PeerKey == nil },
		opt:      RandomIdentity,
	},
	{
		fallback: func(cfg *Config) bool { return cfg.Peerstore == nil },
		opt:      DefaultPeerstore,
	},
}

// Defaults configures dms3-p2p to use the default options. Can be combined with
// other options to *extend* the default options.
var Defaults Option = func(cfg *Config) error {
	for _, def := range defaults {
		if err := cfg.Apply(def.opt); err != nil {
			return err
		}
	}
	return nil
}

// FallbackDefaults applies default options to the dms3-p2p node if and only if no
// other relevent options have been applied. will be appended to the options
// passed into New.
var FallbackDefaults Option = func(cfg *Config) error {
	for _, def := range defaults {
		if !def.fallback(cfg) {
			continue
		}
		if err := cfg.Apply(def.opt); err != nil {
			return err
		}
	}
	return nil
}
