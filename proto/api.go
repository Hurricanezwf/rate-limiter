package proto

type Request struct {
	Action ActionType     `json:"action,omitempty"`
	TID    ResourceTypeID `json:"tId,omitempty"`
	RCID   ResourceID     `json:"rcId,omitempty"`
	CID    ClientID       `json:"cId,omitempty"`
	Expire int64          `json:"expire,omitempty"`
	Quota  uint32         `json:"quota,omitempty"`
}

type Response struct {
	Err    error
	Result interface{}
}
