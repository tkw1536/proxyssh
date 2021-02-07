package proxyssh

import (
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/internal/utils"
)

// Configuration is a configuration for an ssh.Server.
type Configuration interface {
	Apply(logger utils.Logger, server *ssh.Server) error
}

// NewServer makes a new server, applies the appropriate configurations, and then applies the options.
func NewServer(logger utils.Logger, options *Options, configurations ...Configuration) (*ssh.Server, error) {
	server := &ssh.Server{}

	for _, config := range configurations {
		if err := ApplyConfiguration(logger, server, config); err != nil {
			return nil, err
		}
	}

	// Apply the options after the configurations.
	// This ensures that options like 'unsafe' work properly.
	if err := ApplyConfiguration(logger, server, options); err != nil {
		return nil, err
	}

	return server, nil
}

// ApplyConfiguration applies a configuration to the provided ssh.Server.
//
// To apply a configuration, first the configuration.Apply() function is called.
// When Configuration implements Handler, additionally calls ApplyHandler().
//
// When a configuration is nil, ApplyConfiguration does nothing.
func ApplyConfiguration(logger utils.Logger, server *ssh.Server, configuration Configuration) error {

	if configuration == nil {
		return nil
	}

	// apply the configuration!
	if err := configuration.Apply(logger, server); err != nil {
		return err
	}

	// if it is a handler, apply it!
	if handler, isHandler := configuration.(Handler); isHandler {
		if err := ApplyHandler(logger, server, handler); err != nil {
			return err
		}
	}

	return nil
}
