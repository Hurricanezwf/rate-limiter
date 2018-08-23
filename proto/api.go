package proto

type Request struct {
	Action byte `json:"action"`

	LeaderNotify *CMDLeaderNotify   `json:"leaderNotify,omitempty"`
	RegistQuota  *APIRegistQuotaReq `json:"registQuota,omitempty"`
	Borrow       *APIBorrowReq      `json:"borrow,omitempty"`
	Return       *APIReturnReq      `json:"return,omitempty"`
	ReturnAll    *APIReturnAllReq   `json:"returnAll,omitempty"`
}

type Response struct {
	Action byte
	Err    error

	Borrow *APIBorrowResp
}

// 通知Leader的Http服务地址
type CMDLeaderNotify struct {
	RaftAddr  string `json:"raftAddr,omitempty"`
	HttpdAddr string `json:"httpdAddr,omitempty"`
}

type APIRegistQuotaReq struct {
	RCTypeID []byte `json:"rcTypeId,omitempty"`
	Quota    uint32 `json:"quota,omitempty"`
}

type APIBorrowReq struct {
	RCTypeID []byte `json:"rcTypeId,omitempty"`
	ClientID []byte `json:"clientId,omitempty"`
	Expire   int64  `json:"expire,omitempty"`
}

type APIBorrowResp struct {
	RCID string `json:"rcId,omitempty"`
}

type APIReturnReq struct {
	ClientID []byte `json:"clientId,omitempty"`
	RCID     string `json:"rcId,omitempty"`
}

type APIReturnAllReq struct {
	ClientID []byte `json:"clientId,omitempty"`
}
