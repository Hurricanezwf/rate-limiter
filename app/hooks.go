package app

var startHooks []hookfunc
var stopHooks []hookfunc

type hookfunc func() error

func AddStartHook(f hookfunc) {
	startHooks = append(startHooks, f)
}

func AddStopHook(f hookfunc) {
	stopHooks = append(stopHooks, f)
}
