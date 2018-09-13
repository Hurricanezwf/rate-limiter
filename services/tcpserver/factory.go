package tcpserver

import (
	"fmt"
)

var builders = make(map[string]builder)

type builder func() Interface

func RegistBuilder(name string, f builder) {
	if f == nil {
		panic(fmt.Sprintf("TCPServerBuilder for '%s' is nil", name))
	}
	if _, exist := builders[name]; exist {
		panic(fmt.Sprintf("TCPServerBuilder for '%s' had been existed", name))
	}
	builders[name] = f
}
