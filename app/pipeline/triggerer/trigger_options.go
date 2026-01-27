package triggerer

type config struct {
	DisableTx bool
}

type Option interface {
	Apply(*config)
}

type OptionDisableTx struct {
}

func (o OptionDisableTx) Apply(c *config) {
	c.DisableTx = true
}
