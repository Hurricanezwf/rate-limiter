package main

import (
	"encoding/base64"
	"log"
	"strings"

	ratelimiter "github.com/Hurricanezwf/rate-limiter/sdk/ratelimiter-go"
	"github.com/golang/protobuf/jsonpb"
	"github.com/spf13/cobra"
)

var (
	Cluster       string
	EnableBase64  bool
	RCType        string
	Quota         uint32
	Expire        int64
	ResetInterval int64
	Robot         int
	Worker        int
	Duration      int
)

func init() {
	log.SetFlags(log.LstdFlags)
}

func rootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:       "ratelimiterctl",
		Short:     "ratelimiterctl is a client tool for managing service",
		ValidArgs: []string{"regist", "delete", "borrow", "return", "returnAll", "rcList"},
		Args:      cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			if len(args) <= 0 {
				cmd.Help()
			}
		},
	}
	rootCmd.PersistentFlags().StringVar(&Cluster, "cluster", "127.0.0.1:20000,127.0.0.1:20001,127.0.0.1:20002", "Address of rate-limiter cluster")
	return rootCmd
}

func registCmd() *cobra.Command {
	registCmd := &cobra.Command{
		Use:       "regist",
		Short:     "注册资源配额",
		ValidArgs: []string{"rctype", "quota", "resetInterval"},
		Args:      cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			rcType, err := base64.StdEncoding.DecodeString(RCType)
			if err != nil {
				log.Printf("Resolve resource type failed, %v\n", err)
				return
			}

			c, err := ratelimiter.New(&ratelimiter.ClientConfig{
				Cluster: strings.Split(Cluster, ","),
			})
			if err != nil {
				log.Println(err.Error())
				return
			}
			if err = c.RegistQuota(rcType, Quota, ResetInterval); err != nil {
				log.Println(err.Error())
				return
			}
			log.Println("OK")
		},
	}

	registCmd.PersistentFlags().StringVar(&RCType, "rctype", "", "[Required] Type of resource with base64 encoding")
	registCmd.PersistentFlags().Uint32Var(&Quota, "quota", 0, "[Required] Quota of resource")
	registCmd.PersistentFlags().Int64Var(&ResetInterval, "resetInterval", 0, "[Required] Time interval seconds of being recycled to reuse")

	return registCmd
}

func deleteCmd() *cobra.Command {
	deleteCmd := &cobra.Command{
		Use:       "delete",
		Short:     "删除资源配额",
		ValidArgs: []string{"rctype"},
		Args:      cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			rcType, err := base64.StdEncoding.DecodeString(RCType)
			if err != nil {
				log.Printf("Resolve resource type failed, %v\n", err)
				return
			}

			c, err := ratelimiter.New(&ratelimiter.ClientConfig{
				Cluster: strings.Split(Cluster, ","),
			})
			if err != nil {
				log.Println(err.Error())
				return
			}
			if err = c.DeleteQuota(rcType); err != nil {
				log.Println(err.Error())
				return
			}
			log.Println("OK")
		},
	}

	deleteCmd.PersistentFlags().StringVar(&RCType, "rctype", "", "[Required] Type of resource with base64 encoding")

	return deleteCmd
}

func rcListCmd() *cobra.Command {
	rcListCmd := &cobra.Command{
		Use:       "rcList",
		Short:     "查询资源列表详情",
		ValidArgs: []string{"rctype"},
		Args:      cobra.OnlyValidArgs,
		Run: func(cmd *cobra.Command, args []string) {
			var err error
			var rcType []byte
			if len(RCType) > 0 {
				rcType, err = base64.StdEncoding.DecodeString(RCType)
				if err != nil {
					log.Printf("Resolve resource type failed, %v\n", err)
					return
				}
			}

			c, err := ratelimiter.New(&ratelimiter.ClientConfig{
				Cluster: strings.Split(Cluster, ","),
			})
			if err != nil {
				log.Println(err.Error())
				return
			}
			details, err := c.ResourceList(rcType)
			if err != nil {
				log.Println(err.Error())
				return
			}

			encoder := jsonpb.Marshaler{EmitDefaults: true}
			for i, dt := range details {
				str, _ := encoder.MarshalToString(dt)
				log.Printf("(%d) %s\n", i, str)
			}
			log.Println("OK")
		},
	}

	rcListCmd.PersistentFlags().StringVar(&RCType, "rctype", "", "[Optional] Type of resource with base64 encoding")

	return rcListCmd
}

func stressCmd() *cobra.Command {
	stressCmd := &cobra.Command{
		Use:       "stress",
		Short:     "压力测试",
		ValidArgs: []string{"rctype", "robot", "rate", "duration"},
		Args:      cobra.OnlyValidArgs,
		Run:       stressFunc,
	}

	stressCmd.PersistentFlags().StringVar(&RCType, "rctype", "", "[Optional] Type of resource with base64 encoding")
	stressCmd.PersistentFlags().IntVar(&Robot, "robot", 0, "[Optional] Robot count for testing")
	stressCmd.PersistentFlags().IntVar(&Worker, "worker", 0, "[Optional] The worker number which eath robot had.")
	stressCmd.PersistentFlags().IntVar(&Duration, "duration", 0, "[Optional] Duration of stress test, unit is second")

	return stressCmd
}

//func borrowCmd() *cobra.Command {
//	borrowCmd := &cobra.Command{
//		Use:       "borrow",
//		Short:     "删除资源配额",
//		ValidArgs: []string{"rc", "expire"},
//		Args:      cobra.OnlyValidArgs,
//		Run: func(cmd *cobra.Command, args []string) {
//			if len(args) <= 0 {
//				cmd.Help()
//				return
//			}
//			rcType, err := encoding.StringToBytes(RCType)
//			if err != nil {
//				log.Printf("Resolve resource type failed, %v\n", err)
//				return
//			}
//			c, err := ratelimiter.New(&ratelimiter.ClientConfig{
//				Cluster: strings.Split(Cluster, ","),
//			})
//			if err != nil {
//				log.Println(err.Error())
//				return
//			}
//			if err = c.Borrow(rcType, Expire); err != nil {
//				log.Println(err.Error())
//				return
//			}
//			log.Println("OK")
//		},
//	}
//
//	borrowCmd.PersistentFlags().StringVar(&RCType, "rc", "", "Type of resource with base64 encoding")
//	borrowCmd.PersistentFlags().Int64Var(&Expire, "expire", 0, "Max time before resource was recycled automatically")
//
//	return borrowCmd
//}
