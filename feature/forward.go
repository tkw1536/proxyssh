package feature

import (
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/utils"
)

// AllowForwardTo returns a ssh.LocalPortForwardingCallback that allows forwarding traffic to the provided addresses only.
//
// logger is called whenever a request from a caller is allowed or denied.
func AllowForwardTo(logger utils.Logger, addresses []NetworkAddress) ssh.LocalPortForwardingCallback {
	if len(addresses) > 0 {
		logger.Printf("allow_forward_to %v", addresses)
	}
	return func(ctx ssh.Context, dhost string, dport uint32) bool {
		return filterInternal(logger, "", ctx, addresses, NetworkAddress{Hostname: dhost, Port: NetworkPort(dport)})
	}
}

// AllowForwardFrom returns a ssh.ReversePortForwardingCallback that allows reading traffic to the provided addresses only.
//
// logger is called whenever a request from a caller is allowed or denied.
func AllowForwardFrom(logger utils.Logger, addresses []NetworkAddress) ssh.ReversePortForwardingCallback {
	if len(addresses) > 0 {
		logger.Printf("allow_forward_from %v", addresses)
	}
	return func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
		return filterInternal(logger, "_reverse", ctx, addresses, NetworkAddress{Hostname: bindHost, Port: NetworkPort(bindPort)})
	}
}

// filterInternal is the internal function used by AllowForwardPorts and AllowReversePorts
func filterInternal(logger utils.Logger, logExtra string, ctx ssh.Context, addresses []NetworkAddress, actualAddress NetworkAddress) bool {
	for _, p := range addresses {
		if p.Hostname == actualAddress.Hostname && p.Port == actualAddress.Port {
			utils.FmtSSHLog(logger, ctx, "grant%s_portforward %s", logExtra, actualAddress.String())
			return true
		}
	}
	utils.FmtSSHLog(logger, ctx, "deny%s_portforward %s", logExtra, actualAddress.String())
	return false
}

// AllowPortForwarding enables port forwarding on the provided server only to and from the given addresses.
// This function also calls EnablePortForwarding, please see appropriate documentation
//
// See also AllowForwardTo, AllowForwardFrom and EnablePortForwarding.
//
func AllowPortForwarding(logger utils.Logger, server *ssh.Server, toAddresses []NetworkAddress, fromAddresses []NetworkAddress) {
	EnablePortForwarding(server, AllowForwardTo(logger, toAddresses), AllowForwardFrom(logger, fromAddresses))
}

// EnablePortForwarding enables portforwarding with the given callbacks on the ssh Server server.
// This includes tcpip forward requests as well as direct-tcpip channels.
//
// This function overwrites any already configured LocalPortForwardingCallback and ReversePortForwardingCallback functions.
// It will furthermore remove the 'tcpip-forward' and 'cancel-tcpip-forward' request handlers along with the 'direct-tcpip' channel handler.
func EnablePortForwarding(server *ssh.Server, localCallback ssh.LocalPortForwardingCallback, reverseCallback ssh.ReversePortForwardingCallback) {
	forwardHandler := &ssh.ForwardedTCPHandler{}

	// store the fowarding callbacks
	server.LocalPortForwardingCallback = localCallback
	server.ReversePortForwardingCallback = reverseCallback

	// if we don't have any request or channel handlers, we need to setup the default ones.
	// this code is adapted from ssh.Server.ensureHandlers()
	if server.RequestHandlers == nil {
		server.RequestHandlers = make(map[string]ssh.RequestHandler)
		for n, h := range ssh.DefaultRequestHandlers {
			server.RequestHandlers[n] = h
		}
	}
	if server.ChannelHandlers == nil {
		server.ChannelHandlers = make(map[string]ssh.ChannelHandler)
		for n, h := range ssh.DefaultChannelHandlers {
			server.ChannelHandlers[n] = h
		}
	}

	// setup the channel handlers for tcip forwarding
	server.RequestHandlers["tcpip-forward"] = forwardHandler.HandleSSHRequest
	server.RequestHandlers["cancel-tcpip-forward"] = forwardHandler.HandleSSHRequest

	// allow direct-tcip handlers also
	server.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler
}
