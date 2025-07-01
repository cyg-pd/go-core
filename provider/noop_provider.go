package provider

type NoopProvider struct{}

func (n *NoopProvider) Boot() error     { return nil }
func (n *NoopProvider) Shutdown() error { return nil }

var _ Provider = (*NoopProvider)(nil)
