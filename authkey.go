package proxyssh

import "github.com/gliderlabs/ssh"

// AuthorizeKeys returns an ssh.PublicKeysHandler that calls the keyfinder function to find authorized keys
func AuthorizeKeys(keyfinder func(ctx ssh.Context) (keys []ssh.PublicKey, err error)) ssh.PublicKeyHandler {
	return func(ctx ssh.Context, key ssh.PublicKey) bool {
		// find the keys
		keys, err := keyfinder(ctx)
		if err != nil {
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
