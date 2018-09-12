package main

import (
	"encoding/base64"
	"log"
	"strings"
	"sync/atomic"
	"time"

	ratelimiter "github.com/Hurricanezwf/rate-limiter/sdk/ratelimiter-go"
	"github.com/spf13/cobra"
)

func stressFunc(cmd *cobra.Command, args []string) {
	if len(RCType) <= 0 {
		log.Println("Missing `rctype` flag")
		return
	}
	rcType, err := base64.StdEncoding.DecodeString(RCType)
	if err != nil {
		log.Printf("Resolve resource type failed, %v\n", err)
		return
	}

	var robotMax = Robot
	var ready = make(chan struct{}, 1)
	var action = make(chan struct{})
	var quit = make(chan struct{})
	var summary = make(chan int64, 1)

	// 初始化机器人
	for i := 0; i < robotMax; i++ {
		go stressRobot(i, ready, action, quit, rcType, summary)
	}

	// 等待机器人Ready
	for readyCount := 0; readyCount < robotMax; readyCount++ {
		<-ready
	}

	// 开始压测
	close(ready)
	close(action)

	// 压测一段时间后结束压测
	<-time.After(time.Duration(Duration) * time.Second)
	close(quit)

	// 汇总结果
	var accessCount int64
	for i := 0; i < robotMax; i++ {
		cnt := <-summary
		accessCount += cnt
	}
	close(summary)

	log.Printf("%d reuests in %d seconds, rate=%d次/s\n", accessCount, Duration, accessCount/int64(Duration))
}

func stressRobot(
	seq int,
	ready chan<- struct{},
	action <-chan struct{},
	quit chan struct{},
	rcType []byte,
	summary chan<- int64,
) {
	max := Rate
	ch := make(chan struct{}, max)
	defer close(ch)

	c, err := ratelimiter.New(&ratelimiter.ClientConfig{
		Cluster: strings.Split(Cluster, ","),
	})
	if err != nil {
		log.Fatalln(err.Error())
	}

	// sender
	var accessCount int64
	for i := 0; i < max; i++ {
		go func(accessCount *int64) {
			var err error
			var rcId string
			for _ = range ch {
				rcId, err = c.Borrow(rcType, 30)
				if err != nil {
					log.Println(err.Error())
					continue
				}
				if err = c.Return(rcId); err != nil {
					log.Println(err.Error())
					continue
				}
				atomic.AddInt64(accessCount, 1)
			}
		}(&accessCount)
	}

	ready <- struct{}{}
	<-action

	for {
		select {
		case ch <- struct{}{}:
			// do nothing
		case <-time.After(2 * time.Second):
			log.Println("Robot ", seq, " send request timeout")
		case <-quit:
			break
		}
	}
}
