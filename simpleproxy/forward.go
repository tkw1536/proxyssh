package simpleproxy

import (
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/utils"
)

// AllowForwardPorts sets a list of ports that is allowed to be forwarded
func AllowForwardPorts(logger utils.Logger, ports []utils.NetworkAddress) ssh.LocalPortForwardingCallback {
	return func(ctx ssh.Context, dhost string, dport uint32) bool {
		for _, p := range ports {
			if p.Hostname == dhost && p.Port == dport {
				utils.FmtSSHLog(logger, ctx, "grant_portforward %s:%d", dhost, dport)
				return true
			}
		}
		utils.FmtSSHLog(logger, ctx, "deny_portforward %s:%d", dhost, dport)
		return false
	}
}

// AllowReversePorts sets a list of ports that is allowed to be reverse forwarded
func AllowReversePorts(logger utils.Logger, ports []utils.NetworkAddress) ssh.ReversePortForwardingCallback {
	return func(ctx ssh.Context, bindHost string, bindPort uint32) bool {
		for _, p := range ports {
			if p.Hostname == bindHost && p.Port == bindPort {
				utils.FmtSSHLog(logger, ctx, "grant_reverse_portforward %s:%d", bindHost, bindPort)
				return true
			}
		}
		utils.FmtSSHLog(logger, ctx, "deny_reverse_portforward %s:%d", bindHost, bindPort)
		return false
	}
}

// AllowPortForwarding is like EnablePortForwarding except that it uses list of ports
func AllowPortForwarding(logger utils.Logger, server *ssh.Server, localPorts []utils.NetworkAddress, reversePorts []utils.NetworkAddress) *ssh.Server {
	return EnablePortForwarding(server, AllowForwardPorts(logger, localPorts), AllowReversePorts(logger, reversePorts))
}

// EnablePortForwarding enables port forwarding on server with the given callbacks
func EnablePortForwarding(server *ssh.Server, localCallback ssh.LocalPortForwardingCallback, reverseCallback ssh.ReversePortForwardingCallback) *ssh.Server {
	forwardHandler := &ssh.ForwardedTCPHandler{}

	// store the fowarding callbacks
	server.LocalPortForwardingCallback = localCallback
	server.ReversePortForwardingCallback = reverseCallback

	// setup server requst handlers
	if server.RequestHandlers == nil {
		server.RequestHandlers = make(map[string]ssh.RequestHandler)
		for n, h := range ssh.DefaultRequestHandlers {
			server.RequestHandlers[n] = h
		}
	}
	server.RequestHandlers["tcpip-forward"] = forwardHandler.HandleSSHRequest
	server.RequestHandlers["cancel-tcpip-forward"] = forwardHandler.HandleSSHRequest

	// setup channel handlers
	if server.ChannelHandlers == nil {
		server.ChannelHandlers = make(map[string]ssh.ChannelHandler)
		for n, h := range ssh.DefaultChannelHandlers {
			server.ChannelHandlers[n] = h
		}
	}
	server.ChannelHandlers["direct-tcpip"] = ssh.DirectTCPIPHandler

	return server
}
