package proto

type Request struct {
	Action byte `json:"action"`

	RegistServiceBroadcast *CMDRegistServiceBroadcast `json:"registServiceBroadcast,omitempty"`
	RegistService          *APIRegistServiceReq       `json:"registService,omitempty"`
	RegistQuota            *APIRegistQuotaReq         `json:"registQuota,omitempty"`
	Borrow                 *APIBorrowReq              `json:"borrow,omitempty"`
	Return                 *APIReturnReq              `json:"return,omitempty"`
	ReturnAll              *APIReturnAllReq           `json:"returnAll,omitempty"`
}

type Response struct {
	Action byte
	Err    error

	Borrow *APIBorrowResp
	//RegistQuota *APIRegistQuotaResp `json:"registQuota,omitempty"`
	//Return      *APIReturnResp      `json:"return,omitempty"`
	//Disconnect *APIDisconnectResp `json:"disconnect,omitempty"`
}

// 广播所有结点去注册服务
type CMDRegistServiceBroadcast struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	LeaderUrl string `json:"leaderUrl,omitempty"`
}

// 服务注册请求格式
type APIRegistServiceReq struct {
	Timestamp int64  `json:"timestamp,omitempty"`
	RaftAddr  string `json:"raftAddr,omitempty"`
	HttpdAddr string `json:"httpdAddr,omitempty"`
}

type APIRegistQuotaReq struct {
	RCTypeID []byte `json:"rcTypeId,omitempty"`
	Quota    uint32 `json:"quota,omitempty"`
}

//type APIRegistQuotaResp struct {
//	Err string `json:"error,omitempty"`
//}

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

//type APIReturnResp struct {
//	Err string `json:"error,omitempty"`
//}

type APIReturnAllReq struct {
	ClientID []byte `json:"clientId,omitempty"`
}

//type APIDisconnectResp struct {
//	Err string `json:"error,omitempty"`
//}
