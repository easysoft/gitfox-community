package req

const (
	HeaderContentEncoding = "Content-Encoding"
)

type ContextKey int

const (
	authSessionKey ContextKey = iota
)
