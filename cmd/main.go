package main

import (
	"context"
	"fmt"

	"github.com/LMF709268224/titan-endpoint-finder/pkg/endpoint"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 初始化方式1：本地配置 + 远程配置
	localConfig := map[string][]string{
		"https://service1.api.com": {"192.168.1.10:8080", "192.168.1.11:8080"},
		"https://service2.api.com": {"10.0.0.5:3000"},
	}

	remoteConfigURL := "https://pcdn.titannet.io/config1.json"

	client, err := endpoint.NewClient(ctx, localConfig, remoteConfigURL)
	if err != nil {
		panic(err)
	}

	// 获取可用端点
	targetKey := "https://service1.api.com"
	endpointAddr := client.GetEndpoint(targetKey)
	fmt.Println("Selected Endpoint:", endpointAddr)

	// 获取客户端公网IP
	fmt.Println("Client Public IP:", client.GetClientPublicIP())
}
