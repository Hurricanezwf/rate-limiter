package services

import (
	"crypto/md5"
	"fmt"
	"testing"

	. "github.com/Hurricanezwf/rate-limiter/proto"
	"github.com/Hurricanezwf/toolbox/utils"
)

var (
	hostAddr = "127.0.0.1:17250"

	tId   ResourceTypeID
	cId   ClientID
	quota = 4
)

func TestRegist(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	url := fmt.Sprintf("http://%s/v1/registQuota", hostAddr)
	opt := utils.HttpOptions{
		Body: APIRegistQuotaReq{
			RCTypeID: tId,
			Quota:    1,
		},
	}

	if err := utils.HttpPost(url, &opt, nil); err != nil {
		t.Fatal(err.Error())
	}
}

func TestBorrow(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	url := fmt.Sprintf("http://%s/v1/borrow", hostAddr)
	opt := utils.HttpOptions{
		Body: APIBorrowReq{
			RCTypeID: tId,
			ClientID: cId,
			Expire:   10,
		},
	}

	if err := utils.HttpPost(url, &opt, nil); err != nil {
		t.Fatal(err.Error())
	}
}

func TestReturn(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	url := fmt.Sprintf("http://%s/v1/return", hostAddr)
	opt := utils.HttpOptions{
		Body: APIReturnReq{
			RCID:     "462ec3705249fd4358a1bcd02ce5e43f_rc#0",
			ClientID: cId,
		},
	}

	if err := utils.HttpPost(url, &opt, nil); err != nil {
		t.Fatal(err.Error())
	}
}
