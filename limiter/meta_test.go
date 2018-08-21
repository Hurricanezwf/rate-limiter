package limiter

import (
	"crypto/md5"
	"testing"

	"github.com/Hurricanezwf/rate-limiter/encoding"
	. "github.com/Hurricanezwf/rate-limiter/proto"
)

var (
	tId   ResourceTypeID
	cId   ClientID
	quota = uint32(4)
)

func TestLimiterMeta(t *testing.T) {
	tIdMd5 := md5.Sum([]byte("create_host"))
	cIdMd5 := md5.Sum([]byte("zwf"))
	tId = tIdMd5[:]
	cId = cIdMd5[:]

	m := NewLimiterMeta(tId, quota)

	for i := 0; i < 5; i++ {
		rcId, err := m.Borrow(cId, 10)
		if err != nil {
			t.Fatal(err.Error())
		} else {
			t.Logf("Get RC %s", rcId)
		}
	}
}

func TestEncodeBorrowRecord(t *testing.T) {
	rd := &BorrowRecord{
		ClientID: encoding.NewBytes([]byte("clientid")),
		RCID:     encoding.NewString("rcid"),
		BorrowAt: encoding.NewInt64(123456),
		ExpireAt: encoding.NewInt64(567890),
	}

	// encode
	b, err := rd.Encode()
	if err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Encode: %#v\n", b)
	}

	// decode
	b = append(b, []byte("zhe shi wo gu yi dao luan de")...)
	rdNew := &BorrowRecord{}
	if _, err = rdNew.Decode(b); err != nil {
		t.Fatal(err.Error())
	} else {
		t.Logf("After Decode: \n")
		t.Logf("ClientID: %#v\n", rdNew.ClientID.Value())
		t.Logf("RCID:     %s\n", rdNew.RCID.Value())
		t.Logf("BorrowAt: %d\n", rdNew.BorrowAt.Value())
		t.Logf("ExpireAt: %d\n", rdNew.ExpireAt.Value())
	}
}
