package cluster

import "fmt"

var builders = make(map[string]builder)

type builder func() Interface

func RegistBuilder(name string, f buidler) {
	if f == nil {
		panic(fmt.Sprintf("ClusterBuilder for '%s' to regist is nil", name))
	}
	if _, exist := builders[name]; exist {
		panic(fmt.Sprintf("ClusterBuilder for '%s' had been existed", name))
	}
	builders[name] = f
}
