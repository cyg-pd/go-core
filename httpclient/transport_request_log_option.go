package httpclient

type logOption interface{ apply(*requestLog) }
type logOptionFunc func(*requestLog)

func (o logOptionFunc) apply(conf *requestLog) { o(conf) }

func WithMaxBodySize(maxSize int) logOption {
	return logOptionFunc(func(conf *requestLog) {
		conf.maxBodySize = maxSize
	})
}
