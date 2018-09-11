package limiter

import "fmt"

var builders = make(map[string]builder)

type builder func() Interface

func RegistBuilder(name string, f builder) {
	if f == nil {
		panic(fmt.Sprintf("LimiterBuilder for '%s' to regist is nil", name))
	}
	if _, exist := builders[name]; exist {
		panic(fmt.Sprintf("LimiterBuilder for '%s' had been existed", name))
	}
	builders[name] = f
}
