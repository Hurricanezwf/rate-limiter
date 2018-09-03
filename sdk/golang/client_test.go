package ratelimiter

import (
	"testing"
	"time"
)

var (
	rcTypeId        = []byte("zwf-test")
	quota    uint32 = uint32(10)
	expire   int64  = int64(100)
	cluster         = []string{
		"127.0.0.1:20000",
		"127.0.0.1:20001",
		"127.0.0.1:20002",
	}
)

func TestRegistQuota(t *testing.T) {
	l, err := New(&ClientConfig{
		Cluster: cluster,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer l.Close()

	if err := l.RegistQuota(rcTypeId, quota); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("Regist quota OK")
	}
}

func TestBorrow_(t *testing.T) {
	l, err := New(&ClientConfig{
		Cluster: cluster,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	defer l.Close()

	// 借用资源
	start := time.Now()
	rcId, err := l.Borrow(rcTypeId, expire)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("Borrow OK, rcId=%s, elapse:%v\n", rcId, time.Since(start))
	}
	_ = rcId

	// 模拟做业务处理
	//time.Sleep(5 * time.Second)

	// 归还资源
	start = time.Now()
	if err = l.Return(rcId); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("Return OK, rcId=%s, elapse:%v\n", rcId, time.Since(start))
	}
}

func TestBorrowWithTimeout(t *testing.T) {
	l, err := New(&ClientConfig{
		Cluster: cluster,
	})
	if err != nil {
		t.Fatal(err.Error())
	}
	//defer l.Close()

	// 借用资源
	start := time.Now()
	rcId, err := l.BorrowWithTimeout(rcTypeId, expire, 5*time.Second)
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("Borrow OK, rcId=%s, elapse:%v\n", rcId, time.Since(start))
	}
}