package main

import (
	"context"
	"fmt"

	"github.com/LMF709268224/titan-endpoint-finder/pkg/endpoint"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	key := "https://test18-scheduler.titannet.io:3456/rpc/v0"

	ms := map[string][]string{}
	ms[key] = []string{"121.14.67.147:5000"}

	// 初始化客户端
	eClient := endpoint.NewClient(ctx, ms, "")
	// 获取一个可用的接入点
	endpointAddr := eClient.GetEndpoint(key)
	fmt.Println("Selected Endpoint:", endpointAddr)
	// 获取客户端ip地址
	fmt.Println("my url:", eClient.GetMyIP())
}
