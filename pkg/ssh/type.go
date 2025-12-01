package ssh

type contextKey string

const (
	principalKey   = contextKey("principalKey")
	publicKeyIDKey = contextKey("publicKeyID")
)

var (
	allowedCommands = []string{
		"git-upload-pack",
		"git-receive-pack",
	}
	defaultCiphers = []string{
		"chacha20-poly1305@openssh.com",
		"aes128-ctr",
		"aes192-ctr",
		"aes256-ctr",
		"aes128-gcm@openssh.com",
		"aes256-gcm@openssh.com",
	}
	defaultKeyExchanges = []string{
		"curve25519-sha256",
		"curve25519-sha256@libssh.org",
		"ecdh-sha2-nistp256",
		"ecdh-sha2-nistp384",
		"ecdh-sha2-nistp521",
		"diffie-hellman-group14-sha256",
		"diffie-hellman-group14-sha1",
	}
	defaultMACs = []string{
		"hmac-sha2-256-etm@openssh.com",
		"hmac-sha2-512-etm@openssh.com",
		"hmac-sha2-256",
		"hmac-sha2-512",
	}
	defaultServerKeyPath = "gitfox.rsa"
	KeepAliveMsg         = "keepalive@openssh.com"
)
