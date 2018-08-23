package tests

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	. "github.com/Hurricanezwf/rate-limiter/proto"
)

var (
	hostAddr = "127.0.0.1:20000"

	tId   []byte
	cId   []byte
	quota = uint32(4)
)

func TestRegist(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	url := fmt.Sprintf("http://%s/v1/registQuota", hostAddr)
	buf := bytes.NewBuffer(nil)

	json.NewEncoder(buf).Encode(APIRegistQuotaReq{
		RCTypeID: tId,
		Quota:    10,
	})

	for {
		rp, err := http.Post(url, "application/json", buf)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf.Reset()
		buf.ReadFrom(rp.Body)
		rp.Body.Close()

		if rp.StatusCode == 307 {
			url = rp.Header.Get("Location")
			continue
		}
		if rp.StatusCode != 200 {
			t.Fatalf("StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
		}
		break
	}
}

func TestBorrow(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	url := fmt.Sprintf("http://%s/v1/borrow", hostAddr)
	buf := bytes.NewBuffer(nil)

	json.NewEncoder(buf).Encode(APIBorrowReq{
		RCTypeID: tId,
		ClientID: cId,
		Expire:   1000,
	})

	for {
		rp, err := http.Post(url, "application/json", buf)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf.Reset()
		buf.ReadFrom(rp.Body)
		rp.Body.Close()

		if rp.StatusCode == 307 {
			url = rp.Header.Get("Location")
			continue
		}
		if rp.StatusCode != 200 {
			t.Fatalf("StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
		}
		break
	}
}

func TestReturnOne(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	url := fmt.Sprintf("http://%s/v1/return", hostAddr)
	buf := bytes.NewBuffer(nil)

	json.NewEncoder(buf).Encode(APIReturnReq{
		RCID:     "462ec3705249fd4358a1bcd02ce5e43f_rc#0",
		ClientID: cId,
	})

	for {
		rp, err := http.Post(url, "application/json", buf)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf.Reset()
		buf.ReadFrom(rp.Body)
		rp.Body.Close()

		if rp.StatusCode == 307 {
			url = rp.Header.Get("Location")
			continue
		}
		if rp.StatusCode != 200 {
			t.Fatalf("StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
		}
		break
	}
}

func TestReturnAll(t *testing.T) {
	cIdMd5 := md5.Sum([]byte("zwf"))
	cId = cIdMd5[:]

	url := fmt.Sprintf("http://%s/v1/returnAll", hostAddr)
	buf := bytes.NewBuffer(nil)

	json.NewEncoder(buf).Encode(APIReturnAllReq{
		ClientID: cId,
	})

	for {
		rp, err := http.Post(url, "application/json", buf)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf.Reset()
		buf.ReadFrom(rp.Body)
		rp.Body.Close()

		if rp.StatusCode == 307 {
			url = rp.Header.Get("Location")
			continue
		}
		if rp.StatusCode != 200 {
			t.Fatalf("StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
		}
		break
	}
}

func TestSnapshot(t *testing.T) {
	TestRegist(t)
	TestBorrow(t)
	TestBorrow(t)

	url := fmt.Sprintf("http://%s/v1/snapshot", hostAddr)
	buf := bytes.NewBuffer(nil)

	for {
		rp, err := http.Get(url)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf.Reset()
		buf.ReadFrom(rp.Body)
		rp.Body.Close()

		if rp.StatusCode == 307 {
			url = rp.Header.Get("Location")
			continue
		}
		if rp.StatusCode != 200 {
			t.Fatalf("StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
		}
		break
	}
}

func TestRestore(t *testing.T) {
	url := fmt.Sprintf("http://%s/v1/restore", hostAddr)
	buf := bytes.NewBuffer(nil)

	for {
		rp, err := http.Get(url)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf.Reset()
		buf.ReadFrom(rp.Body)
		rp.Body.Close()

		if rp.StatusCode == 307 {
			url = rp.Header.Get("Location")
			continue
		}
		if rp.StatusCode != 200 {
			t.Fatalf("StatusCode(%d) != 200, %s", rp.StatusCode, buf.String())
		}
		break
	}

	//TestRegist(t)
	for i := 0; i < 10; i++ {
		TestBorrow(t)
		time.Sleep(time.Second)
	}
}
