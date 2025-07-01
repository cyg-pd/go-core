package provider

type Provider interface {
	Boot() error
	Shutdown() error
}
