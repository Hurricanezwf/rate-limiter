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

// CMDBLeaderNotify 通知Leader的Http服务地址
type CMDLeaderNotify struct {
	RaftAddr  string `json:"raftAddr,omitempty"`
	HttpdAddr string `json:"httpdAddr,omitempty"`
}

// APIRegistQuotaReq 注册资源配额接口
type APIRegistQuotaReq struct {
	RCTypeID []byte `json:"rcTypeId,omitempty"`
	Quota    uint32 `json:"quota,omitempty"`
}

// APIBorrowReq 借取资源请求格式
type APIBorrowReq struct {
	RCTypeID []byte `json:"rcTypeId,omitempty"`
	ClientID []byte `json:"clientId,omitempty"`
	Expire   int64  `json:"expire,omitempty"`
}

// APIBorrowResp 借取资源响应格式
type APIBorrowResp struct {
	RCID string `json:"rcId,omitempty"`
}

// APIReturnReq 归还单个资源接口
type APIReturnReq struct {
	ClientID []byte `json:"clientId,omitempty"`
	RCID     string `json:"rcId,omitempty"`
}

// APIReturnAllReq 归还用户所有占用资源的接口
type APIReturnAllReq struct {
	ClientID []byte `json:"clientId,omitempty"`
}
