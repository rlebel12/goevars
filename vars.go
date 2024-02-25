package goenvvars

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
)

var DefaultAllowFallback = defaultAllowFallback

func New(key string, opts ...envVarOpt) *envVar {
	ev := new(envVar)
	ev.key = key
	ev.allowFallback = DefaultAllowFallback
	ev.value, ev.found = os.LookupEnv(key)

	for _, opt := range opts {
		opt(ev)
	}

	return ev
}

func (ev *envVar) Optional() *envVar {
	ev.optional = true
	return ev
}

type fallback struct {
	allow bool
}

type fallbackOpt func(*fallback)

func (ev *envVar) Fallback(value string, opts ...fallbackOpt) *envVar {
	fb := &fallback{
		allow: ev.allowFallback(),
	}

	for _, opt := range opts {
		opt(fb)
	}

	if !ev.found && fb.allow {
		ev.value = value
	}
	return ev
}

func OverrideAllow(allow func() bool) fallbackOpt {
	return func(f *fallback) {
		f.allow = allow()
	}
}

func AllowAlways() fallbackOpt {
	return OverrideAllow(func() bool {
		return true
	})
}

func (ev *envVar) String() string {
	ret, err := ev.TryString()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryString() (string, error) {
	if err := ev.validate(); err != nil {
		return "", fmt.Errorf("invalid string environment variable: %s", ev.value)
	}
	return ev.value, nil
}

func (ev *envVar) Bool() bool {
	ret, err := ev.TryBool()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryBool() (bool, error) {
	if err := ev.validate(); err != nil {
		return false, err
	}
	if ev.value == "" {
		return false, nil
	}
	ret, err := strconv.ParseBool(ev.value)
	if err != nil {
		return false, fmt.Errorf("invalid boolean environment variable: %s", ev.value)
	}
	return ret, nil
}

func (ev *envVar) Int() int {
	ret, err := ev.TryInt()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryInt() (int, error) {
	if err := ev.validate(); err != nil {
		return 0, fmt.Errorf("invalid integer environment variable: %s", ev.value)
	}
	if ev.value == "" {
		return 0, nil
	}
	ret, err := strconv.Atoi(ev.value)
	if err != nil {
		return 0, fmt.Errorf("invalid integer environment variable: %s", ev.value)
	}
	return ret, nil
}

func (ev *envVar) Float64() float64 {
	ret, err := ev.TryFloat64()
	if err != nil {
		panic(err)
	}
	return ret
}

func (ev *envVar) TryFloat64() (float64, error) {
	if err := ev.validate(); err != nil {
		return 0, fmt.Errorf("invalid float environment variable: %s", ev.value)
	}
	if ev.value == "" {
		return 0, nil
	}
	ret, err := strconv.ParseFloat(ev.value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid float environment variable: %s", ev.value)
	}
	return ret, nil
}

// Returns the value of the environment variable as a URL.
// Panics if the value is not a valid URL, but this may happen
// if a scheme is not specified. See the documentation for
// url.Parse for more information.
func (ev *envVar) URL() *url.URL {
	ret, err := ev.TryURL()
	if err != nil {
		panic(err)
	}
	return ret
}

// Returns the value of the environment variable as a URL.
// Fails if the value is not a valid URL, but this may happen
// if a scheme is not specified. See the documentation for
// url.Parse for more information.
func (ev *envVar) TryURL() (*url.URL, error) {
	if err := ev.validate(); err != nil {
		return &url.URL{}, fmt.Errorf("invalid URL environment variable: %s", ev.value)
	}
	if ev.value == "" {
		return &url.URL{}, nil
	}
	ret, err := url.Parse(ev.value)
	if err != nil {
		return &url.URL{}, fmt.Errorf("invalid URL environment variable: %s", ev.value)
	}
	return ret, nil
}

// Returns true if the environment variable with the given key is set and non-empty
func Presence(key string) bool {
	val, ok := os.LookupEnv(key)
	return ok && val != ""
}

type envVar struct {
	key           string
	value         string
	found         bool
	optional      bool
	allowFallback func() bool
}

type envVarOpt func(*envVar)

func (ev *envVar) validate() error {
	if !ev.optional && ev.value == "" {
		return fmt.Errorf("Missing required environment variable: %s", ev.key)
	}
	return nil
}

func defaultAllowFallback() bool {
	return !IsProd()
}
