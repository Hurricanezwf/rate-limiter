package meta

import "fmt"

var builders = make(map[string]builder)

type builder func(rcTypeId []byte, quota uint32) Interface

func RegistBuilder(name string, f builder) {
	if f == nil {
		panic(fmt.Sprintf("MetaBuilder for '%s' is nil", name))
	}
	if _, exist := builders[name]; exist {
		panic(fmt.Sprintf("MetaBuilder for '%s' had been existed", name))
	}
	builders[name] = f
}
