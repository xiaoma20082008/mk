package util

import "strconv"

type Options struct {
	values map[string]string
}

func (o *Options) GetBool(name string, def bool) bool {
	if vstr, exists := o.values[name]; exists {
		v, err := strconv.ParseBool(vstr)
		if err == nil {
			return v
		}
	}
	return def
}

func (o *Options) GetInt(name string, def int32) int32 {
	if vstr, exists := o.values[name]; exists {
		v, err := strconv.ParseInt(vstr, 10, 32)
		if err == nil {
			return int32(v)
		}
	}
	return def
}

func (o *Options) GetString(name string, def string) string {
	if vstr, exists := o.values[name]; exists {
		return vstr
	}
	return def
}

func NewOptions() *Options {
	return &Options{
		values: map[string]string{},
	}
}
