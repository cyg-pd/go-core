package application

type option interface{ apply(*Application) }
type optionFunc func(*Application)

func (fn optionFunc) apply(cfg *Application) { fn(cfg) }
