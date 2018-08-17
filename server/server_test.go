package server

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
	tId = ResourceTypeID(tIdMd5[:])
	cId = ClientID(cIdMd5[:])

	url := fmt.Sprintf("http://%s/v1/registQuota", hostAddr)
	opt := utils.HttpOptions{
		Body: Request{
			TID:   tId,
			Quota: 10,
		},
	}

	if err := utils.HttpPost(url, &opt, nil); err != nil {
		t.Fatal(err.Error())
	}
}

func TestBorrow(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = ResourceTypeID(tIdMd5[:])
	cId = ClientID(cIdMd5[:])

	url := fmt.Sprintf("http://%s/v1/borrow", hostAddr)
	opt := utils.HttpOptions{
		Body: Request{
			TID:    tId,
			CID:    cId,
			Expire: 10,
		},
	}

	if err := utils.HttpPost(url, &opt, nil); err != nil {
		t.Fatal(err.Error())
	}
}

func TestReturn(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = ResourceTypeID(tIdMd5[:])
	cId = ClientID(cIdMd5[:])

	url := fmt.Sprintf("http://%s/v1/return", hostAddr)
	opt := utils.HttpOptions{
		Body: Request{
			RCID: ResourceID("462ec3705249fd4358a1bcd02ce5e43f_rc#0"),
			CID:  cId,
		},
	}

	if err := utils.HttpPost(url, &opt, nil); err != nil {
		t.Fatal(err.Error())
	}
}
