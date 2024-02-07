package kanthorsdk

type Options struct {
	Debug  bool
	Host   string
	Scheme string
}

type Option func(*Options)

func WithDebug(debug bool) Option {
	return func(o *Options) {
		o.Debug = debug
	}
}

func WithHost(host string) Option {
	return func(o *Options) {
		o.Host = host
	}
}

func WithScheme(scheme string) Option {
	return func(o *Options) {
		o.Scheme = scheme
	}
}
