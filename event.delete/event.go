package event

var mgr *EventManager

func Open(conf *Config) error {
	mgr = &EventManager{}
	return mgr.Open(conf)
}

func PushBack(e *Event) error {
	return mgr.PushBack(e)
}
