package proto

type Request struct {
	Action byte `json:"action,omitempty"`

	RegistQuota *APIRegistQuotaReq `json:"registQuota,omitempty"`
	Borrow      *APIBorrowReq      `json:"borrow,omitempty"`
	Return      *APIReturnReq      `json:"return,omitempty"`
	Disconnect  *APIDisconnectReq  `json:"disconnect,omitempty"`
}

type Response struct {
	Action byte
	Err    error

	Borrow *APIBorrowResp
	//RegistQuota *APIRegistQuotaResp `json:"registQuota,omitempty"`
	//Return      *APIReturnResp      `json:"return,omitempty"`
	//Disconnect *APIDisconnectResp `json:"disconnect,omitempty"`
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

type APIDisconnectReq struct {
	ClientID []byte `json:"clientId,omitempty"`
}

//type APIDisconnectResp struct {
//	Err string `json:"error,omitempty"`
//}
