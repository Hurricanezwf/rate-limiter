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
	hostAddr = "127.0.0.1:20001"

	tId   []byte
	cId   []byte
	quota = uint32(4)
)

func TestRegist(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	for {
		url := fmt.Sprintf("http://%s/v1/registQuota", hostAddr)

		b, err := json.Marshal(APIRegistQuotaReq{
			RCTypeID: tId,
			Quota:    10,
		})
		if err != nil {
			t.Fatal(err.Error())
		}

		buf := bytes.NewBuffer(b)
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

	for {
		url := fmt.Sprintf("http://%s/v1/borrow", hostAddr)

		b, err := json.Marshal(APIBorrowReq{
			RCTypeID: tId,
			ClientID: cId,
			Expire:   1000,
		})
		if err != nil {
			t.Fatal(err.Error())
		}

		buf := bytes.NewBuffer(b)
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

	for {
		url := fmt.Sprintf("http://%s/v1/return", hostAddr)

		b, err := json.Marshal(APIReturnReq{
			RCID:     "462ec3705249fd4358a1bcd02ce5e43f_rc#0",
			ClientID: cId,
		})
		if err != nil {
			t.Fatal(err.Error())
		}

		buf := bytes.NewBuffer(b)
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

	for {
		url := fmt.Sprintf("http://%s/v1/returnAll", hostAddr)

		b, err := json.Marshal(APIReturnAllReq{
			ClientID: cId,
		})
		if err != nil {
			t.Fatal(err.Error())
		}

		buf := bytes.NewBuffer(b)
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

	for {
		url := fmt.Sprintf("http://%s/v1/snapshot", hostAddr)

		rp, err := http.Get(url)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf := bytes.NewBuffer(nil)
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
	for {
		url := fmt.Sprintf("http://%s/v1/restore", hostAddr)

		rp, err := http.Get(url)
		if err != nil {
			t.Fatal(err.Error())
		}
		buf := bytes.NewBuffer(nil)
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
