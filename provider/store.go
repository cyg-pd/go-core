package provider

import (
	"errors"
	"sync"
)

var isBoot = false

var provider = []Provider{}
var mu sync.Mutex

func All() []Provider { return provider }
func Register(p Provider) {
	mu.Lock()
	defer mu.Unlock()

	provider = append(provider, p)
	if isBoot {
		p.Boot()
	}
}

func Boot() error {
	if isBoot {
		return nil
	}

	for _, provider := range provider {
		if err := provider.Boot(); err != nil {
			return err
		}
	}

	isBoot = true
	return nil
}

func Shutdown() error {
	errs := []error{}
	for _, v := range provider {
		errs = append(errs, v.Shutdown())
	}
	return errors.Join(errs...)
}
