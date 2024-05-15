package options

import (
	"time"
)

type Options struct {
	Expire    time.Duration
	Serialize Serialize
}

func NewOptions() *Options {
	return &Options{}
}

func (o *Options) SetSerialize(serialize Serialize) *Options {
	o.Serialize = serialize
	return o
}

func (o *Options) SetExpire(exp time.Duration) *Options {
	o.Expire = exp
	return o
}

func (o *Options) Merge(options ...*Options) *Options {
	opt := *o

	for _, option := range options {
		if option == nil {
			continue
		}

		if option.Serialize.Valid() {
			opt.Serialize = option.Serialize
		}

		if option.Expire > 0 {
			opt.Expire = option.Expire
		}
	}

	return &opt
}
