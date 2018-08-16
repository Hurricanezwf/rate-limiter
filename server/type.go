package server

import "github.com/Hurricanezwf/rate-limiter/limiter"

type APIGetReq struct {
	CID limiter.ClientID      `json:"clientId"`
	PID limiter.RequestPathID `json:"requestPathId"`
}
