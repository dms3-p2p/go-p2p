package p2p

// This file contains all dms3-p2p configuration options (except the defaults,
// those are in defaults.go)

import (
	"fmt"
	"net"

	config "github.com/dms3-p2p/go-p2p/config"
	bhost "github.com/dms3-p2p/go-p2p/p2p/host/basic"

	circuit "github.com/dms3-p2p/go-p2p-circuit"
	crypto "github.com/dms3-p2p/go-p2p-crypto"
	ifconnmgr "github.com/dms3-p2p/go-p2p-interface-connmgr"
	pnet "github.com/dms3-p2p/go-p2p-interface-pnet"
	metrics "github.com/dms3-p2p/go-p2p-metrics"
	pstore "github.com/dms3-p2p/go-p2p-peerstore"
	filter "github.com/dms3-p2p/go-maddr-filter"
	ma "github.com/dms3-mft/go-multiaddr"
)

// ListenAddrStrings configures dms3-p2p to listen on the given (unparsed)
// addresses.
func ListenAddrStrings(s ...string) Option {
	return func(cfg *Config) error {
		for _, addrstr := range s {
			a, err := ma.NewMultiaddr(addrstr)
			if err != nil {
				return err
			}
			cfg.ListenAddrs = append(cfg.ListenAddrs, a)
		}
		return nil
	}
}

// ListenAddrs configures dms3-p2p to listen on the given addresses.
func ListenAddrs(addrs ...ma.Multiaddr) Option {
	return func(cfg *Config) error {
		cfg.ListenAddrs = append(cfg.ListenAddrs, addrs...)
		return nil
	}
}

// Security configures dms3-p2p to use the given security transport (or transport
// constructor).
//
// Name is the protocol name.
//
// The transport can be a constructed security.Transport or a function taking
// any subset of this dms3-p2p node's:
// * Public key
// * Private key
// * Peer ID
// * Host
// * Network
// * Peerstore
func Security(name string, tpt interface{}) Option {
	stpt, err := config.SecurityConstructor(tpt)
	err = traceError(err, 1)
	return func(cfg *Config) error {
		if err != nil {
			return err
		}
		if cfg.Insecure {
			return fmt.Errorf("cannot use security transports with an insecure dms3-p2p configuration")
		}
		cfg.SecurityTransports = append(cfg.SecurityTransports, config.MsSecC{SecC: stpt, ID: name})
		return nil
	}
}

// NoSecurity is an option that completely disables all transport security.
// It's incompatible with all other transport security protocols.
var NoSecurity Option = func(cfg *Config) error {
	if len(cfg.SecurityTransports) > 0 {
		return fmt.Errorf("cannot use security transports with an insecure dms3-p2p configuration")
	}
	cfg.Insecure = true
	return nil
}

// Muxer configures dms3-p2p to use the given stream multiplexer (or stream
// multiplexer constructor).
//
// Name is the protocol name.
//
// The transport can be a constructed mux.Transport or a function taking any
// subset of this dms3-p2p node's:
// * Peer ID
// * Host
// * Network
// * Peerstore
func Muxer(name string, tpt interface{}) Option {
	mtpt, err := config.MuxerConstructor(tpt)
	err = traceError(err, 1)
	return func(cfg *Config) error {
		if err != nil {
			return err
		}
		cfg.Muxers = append(cfg.Muxers, config.MsMuxC{MuxC: mtpt, ID: name})
		return nil
	}
}

// Transport configures dms3-p2p to use the given transport (or transport
// constructor).
//
// The transport can be a constructed transport.Transport or a function taking
// any subset of this dms3-p2p node's:
// * Transport Upgrader (*tptu.Upgrader)
// * Host
// * Stream muxer (muxer.Transport)
// * Security transport (security.Transport)
// * Private network protector (pnet.Protector)
// * Peer ID
// * Private Key
// * Public Key
// * Address filter (filter.Filter)
// * Peerstore
func Transport(tpt interface{}) Option {
	tptc, err := config.TransportConstructor(tpt)
	err = traceError(err, 1)
	return func(cfg *Config) error {
		if err != nil {
			return err
		}
		cfg.Transports = append(cfg.Transports, tptc)
		return nil
	}
}

// Peerstore configures dms3-p2p to use the given peerstore.
func Peerstore(ps pstore.Peerstore) Option {
	return func(cfg *Config) error {
		if cfg.Peerstore != nil {
			return fmt.Errorf("cannot specify multiple peerstore options")
		}

		cfg.Peerstore = ps
		return nil
	}
}

// PrivateNetwork configures dms3-p2p to use the given private network protector.
func PrivateNetwork(prot pnet.Protector) Option {
	return func(cfg *Config) error {
		if cfg.Protector != nil {
			return fmt.Errorf("cannot specify multiple private network options")
		}

		cfg.Protector = prot
		return nil
	}
}

// BandwidthReporter configures dms3-p2p to use the given bandwidth reporter.
func BandwidthReporter(rep metrics.Reporter) Option {
	return func(cfg *Config) error {
		if cfg.Reporter != nil {
			return fmt.Errorf("cannot specify multiple bandwidth reporter options")
		}

		cfg.Reporter = rep
		return nil
	}
}

// Identity configures dms3-p2p to use the given private key to identify itself.
func Identity(sk crypto.PrivKey) Option {
	return func(cfg *Config) error {
		if cfg.PeerKey != nil {
			return fmt.Errorf("cannot specify multiple identities")
		}

		cfg.PeerKey = sk
		return nil
	}
}

// ConnectionManager configures dms3-p2p to use the given connection manager.
func ConnectionManager(connman ifconnmgr.ConnManager) Option {
	return func(cfg *Config) error {
		if cfg.ConnManager != nil {
			return fmt.Errorf("cannot specify multiple connection managers")
		}
		cfg.ConnManager = connman
		return nil
	}
}

// AddrsFactory configures dms3-p2p to use the given address factory.
func AddrsFactory(factory config.AddrsFactory) Option {
	return func(cfg *Config) error {
		if cfg.AddrsFactory != nil {
			return fmt.Errorf("cannot specify multiple address factories")
		}
		cfg.AddrsFactory = factory
		return nil
	}
}

// EnableRelay configures dms3-p2p to enable the relay transport.
func EnableRelay(options ...circuit.RelayOpt) Option {
	return func(cfg *Config) error {
		cfg.Relay = true
		cfg.RelayOpts = options
		return nil
	}
}

// FilterAddresses configures dms3-p2p to never dial nor accept connections from
// the given addresses.
func FilterAddresses(addrs ...*net.IPNet) Option {
	return func(cfg *Config) error {
		if cfg.Filters == nil {
			cfg.Filters = filter.NewFilters()
		}
		for _, addr := range addrs {
			cfg.Filters.AddDialFilter(addr)
		}
		return nil
	}
}

// NATPortMap configures dms3-p2p to use the default NATManager. The default
// NATManager will attempt to open a port in your network's firewall using UPnP.
func NATPortMap() Option {
	return NATManager(bhost.NewNATManager)
}

// NATManager will configure dms3-p2p to use the requested NATManager. This
// function should be passed a NATManager *constructor* that takes a dms3-p2p Network.
func NATManager(nm config.NATManagerC) Option {
	return func(cfg *Config) error {
		if cfg.NATManager != nil {
			return fmt.Errorf("cannot specify multiple NATManagers")
		}
		cfg.NATManager = nm
		return nil
	}
}

// NoListenAddrs will configure dms3-p2p to not listen by default.
//
// This will both clear any configured listen addrs and prevent dms3-p2p from
// applying the default listen address option.
var NoListenAddrs = func(cfg *Config) error {
	cfg.ListenAddrs = []ma.Multiaddr{}
	return nil
}

// NoTransports will configure dms3-p2p to not enable any transports.
//
// This will both clear any configured transports (specified in prior dms3-p2p
// options) and prevent dms3-p2p from applying the default transports.
var NoTransports = func(cfg *Config) error {
	cfg.Transports = []config.TptC{}
	return nil
}
