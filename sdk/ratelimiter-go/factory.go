package ratelimiter

import "fmt"

var builders = make(map[string]builder)

type builder func(conf *ClientConfig) (Interface, error)

func RegistBuilder(name string, f builder) {
	if f == nil {
		panic(fmt.Sprintf("ClientBuilder for '%s' is nil", name))
	}
	if _, exist := builders[name]; exist {
		panic(fmt.Sprintf("ClientBuilder for '%s' had been existed", name))
	}
	builders[name] = f
}
