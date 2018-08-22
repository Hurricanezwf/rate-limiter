package utils

import (
	"bytes"
	"fmt"
	"log"
	"reflect"
	"time"

	"github.com/astaxie/beego/httplib"
)

var EnableHTTPDebug bool

type HttpOptions struct {
	Headers        map[string]string
	Body           interface{} // Body表示struct结构,上层无须Marshal
	Params         map[string]string
	ConnectTimeout time.Duration
	RWTimeout      time.Duration
	Retry          int
}

func HttpHead(url string, option *HttpOptions) error {
	return httpSend(httplib.Head(url), option, nil)
}

func HttpGet(url string, option *HttpOptions, result interface{}) error {
	return httpSend(httplib.Get(url), option, result)
}

func HttpPost(url string, option *HttpOptions, result interface{}) error {
	return httpSend(httplib.Post(url), option, result)
}

func HttpPut(url string, option *HttpOptions, result interface{}) error {
	return httpSend(httplib.Put(url), option, result)
}

func HttpDelete(url string, option *HttpOptions, result interface{}) error {
	return httpSend(httplib.Delete(url), option, result)
}

func httpSend(req *httplib.BeegoHTTPRequest, option *HttpOptions, result interface{}) error {
	if req == nil {
		return fmt.Errorf("Http request instance is nil")
	}

	var (
		err error
		ok  bool
		b   []byte
	)

	// 1. Set Option
	if option != nil {
		// set headers
		for k, v := range option.Headers {
			req.Header(k, v)
		}

		// set params
		for k, v := range option.Params {
			req.Param(k, v)
		}

		// set rwtimeout
		if option.ConnectTimeout < time.Second || option.RWTimeout < time.Second {
			option.ConnectTimeout = 30 * time.Second
			option.RWTimeout = time.Minute
		}
		req.SetTimeout(option.ConnectTimeout, option.RWTimeout)

		// set retries
		req.Retries(option.Retry)

		// set body
		if option.Body != nil {
			if b, ok = option.Body.([]byte); !ok {
				b, err = json.Marshal(option.Body)
				if err != nil {
					return fmt.Errorf("Marshal body failed, %v", err)
				}
			}
			if len(b) > 0 {
				req.Body(b)
			}
		}
	}

	// 2. Request And Response
	resp, err := req.Debug(EnableHTTPDebug).Response()
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if EnableHTTPDebug {
		log.Printf("\n", string(req.DumpRequest()), "\n", "\n")
	}

	// 3. Read Response Content
	n := int64(0)
	buf := BytesBuffer.Get(bytes.MinRead)
	defer BytesBuffer.Put(buf)

	if n, err = buf.ReadFrom(resp.Body); err != nil {
		return fmt.Errorf("Read http body failed, %v", err)
	}

	// 4. Resolve Response
	if resp.StatusCode != 200 {
		return fmt.Errorf("StatusCode(%d) != 200, %s", resp.StatusCode, buf.String())
	}
	if result != nil {
		log.Printf("%s: %s\n", reflect.TypeOf(result), buf.String())
		if err = JsonUnmarshal(buf.Next(int(n)), result); err != nil {
			return fmt.Errorf("Bad response format, %v", err)
		}
	}
	return nil
}
