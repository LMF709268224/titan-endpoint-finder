# titan-endpoint-finder

## ✨ 核心功能
### 动态接入点管理​​
1.支持本地映射（map[string][]string）与远程JSON配置（AWS S3/CDN）双配置源 优先级：​​本地配置 < 远程配置​​
### ​智能负载均衡​​
1.基于随机选择的节点调度策略（避免单点过载）
2.实时健康检查（TCP/UDP双协议支持）
### ​客户端识别​​
1.自动获取客户端公网IP（通过服务端API）
### ​自愈能力​​
1.周期性接入点健康检查（默认间隔：10分钟）

## 🚀 快速开始
### 获取项目代码
go get https://github.com/LMF709268224/titan-endpoint-finder.git

### 代码示例
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

	// 获取可用接入点
	targetKey := "https://service1.api.com"
	endpointAddr := client.SelectOne(targetKey)
	fmt.Println("Selected Endpoint:", endpointAddr)

	// 获取客户端公网IP
	fmt.Println("Client Public IP:", client.GetClientPublicIP())
}

## ⚙️ 配置机制
### 本地配置
​​结构​​：map[服务标识][]接入点地址
​​示例​​：
localConfig := map[string][]string{
    "https://service1.api.com": {"192.168.1.10:8080", "192.168.1.11:8080"},
    "https://service2.api.com": {"10.0.0.5:3000"},
}
### 远程配置
​​URL示例​​：https://pcdn.titannet.io/config1.json
​​JSON格式​​：
{
  "https://service1.api.com": ["192.168.1.10:8080", "192.168.1.11:8080"],
  "https://service2.api.com": ["10.0.0.5:3000"]
}