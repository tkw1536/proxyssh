package server

import (
	"github.com/gliderlabs/ssh"
	"github.com/tkw1536/proxyssh/utils"
)

// AuthorizeKeys returns an ssh.PublicKeysHandler that calls keyfinder() and authorizes all keys returned by the handler.
//
// logger is a logger that is called when the keyfinder returns an error.
// keyfinder is a function that is called for a provided context and should return a list of keys that are authorized.
//
// This function protects against timing attacks by always checking all keys even though a quicker implementation could be possible.
// To further improve against timing attacks, keyfinder should always return the same number of keys (for any session).
func AuthorizeKeys(logger utils.Logger, keyfinder func(ctx ssh.Context) (keys []ssh.PublicKey, err error)) ssh.PublicKeyHandler {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		keys, err := keyfinder(ctx)
		if err != nil {
			utils.FmtSSHLog(logger, ctx, "error_keyfinder %s", err.Error())
			return false
		}

		// we could do a 'return true' but we explicitly do not
		// to avoid timing attacks to find which key the user suplied
		var res bool
		for _, ak := range keys {
			if ssh.KeysEqual(ak, key) {
				res = true
			}
		}
		return res
	}
}
