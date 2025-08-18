package external

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func SetupFlags(f *pflag.FlagSet, v *viper.Viper, key string, service string) {
	f.String("external-"+key+"-url", "http://example.com", service+" URL")
	f.StringSlice("external-"+key+"-header", []string{}, service+" Custom Header (e.g., 'Key:Value')")
	_ = v.BindPFlag("external."+key+".url", f.Lookup("external-"+key+"-url"))
	_ = v.BindPFlag("external."+key+".header", f.Lookup("external-"+key+"-header"))
}

type External struct {
	URL     *url.URL    `json:"url" yaml:"url" mapstructure:"url"`
	Headers http.Header `json:"headers" yaml:"headers" mapstructure:"headers"`
}

func (c *External) resolvePath(path string) string {
	if c.URL == nil {
		return path
	}
	return c.URL.ResolveReference(&url.URL{Path: path}).String()
}

func (c *External) NewRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, c.resolvePath(path), body)
	if err != nil {
		return nil, err
	}

	req.Header = c.Headers.Clone()
	if host := req.Header.Get("host"); len(host) > 0 {
		req.Header.Del(host)
		req.Host = host
	}

	return req, nil
}

func New(v *viper.Viper, name string) (*External, error) {
	var c External

	if vv := v.GetString("external." + name + ".url"); len(vv) > 0 {
		var err error
		if c.URL, err = url.Parse(vv); err != nil {
			return nil, err
		}
	}

	if vv := v.GetStringSlice("external." + name + ".header"); len(vv) > 0 {
		c.Headers = make(http.Header, len(vv))
		for _, h := range vv {
			parts := strings.SplitN(h, ":", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			c.Headers.Add(key, value)
		}
	}

	return &c, nil
}
