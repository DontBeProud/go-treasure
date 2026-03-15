package gtnacos

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// Client nacos客户端抽象类
type Client interface {
	// GetClientParam 获取客户端参数信息
	GetClientParam() *vo.NacosClientParam
	// NewConfigCenterClient 创建配置中心客户端
	NewConfigCenterClient() (ConfigCenterClient, func(), error)
	// NewRegistryCenterClient 创建注册中心客户端
	NewRegistryCenterClient() (RegistryCenterClient, func(), error)
}

// nacosClient nacos客户端
type nacosClient struct {
	clientParam *vo.NacosClientParam
}

// GetClientParam 获取客户端参数信息
func (c *nacosClient) GetClientParam() *vo.NacosClientParam {
	return c.clientParam
}

// NewClient 创建NacosClient
func NewClient(nc *gtconfpb.NacosConfig) (c Client, cleanup func(), err error) {
	if nc == nil {
		return nil, func() {}, errors.New("nacos config is nil")
	}

	serverCfg, err := parseServerConfig(nc)
	if err != nil {
		return nil, func() {}, err
	}
	clientCfg := parseClientConfig(nc)

	return &nacosClient{clientParam: &vo.NacosClientParam{
		ServerConfigs: serverCfg,
		ClientConfig:  clientCfg,
	}}, func() {}, nil
}

func parseServerConfig(nc *gtconfpb.NacosConfig) ([]constant.ServerConfig, error) {
	epNum := len(nc.Endpoint)
	if epNum == 0 {
		return nil, errors.New("nacos config endpoints is empty")
	}

	serverCfg := make([]constant.ServerConfig, epNum)
	for idx, ep := range nc.Endpoint {
		if ep == "" {
			return nil, errors.New("nacos config endpoint is empty")
		}
		switch strings.Count(ep, ":") {
		// 不包含冒号，即仅传入了终端地址
		case 0:
			serverCfg[idx] = *constant.NewServerConfig(ep, 8848)
		case 1:
			sp := strings.Split(ep, ":")
			ip := sp[0]
			port, _err := strconv.Atoi(sp[1])
			if _err != nil {
				return nil, fmt.Errorf("%s: invalid nacos endpoint", ep)
			}
			serverCfg[idx] = *constant.NewServerConfig(ip, uint64(port))
		default:
			return nil, fmt.Errorf("%s: invalid nacos endpoint", ep)
		}
	}

	return serverCfg, nil
}

func parseClientConfig(nc *gtconfpb.NacosConfig) *constant.ClientConfig {
	clientCfg := &constant.ClientConfig{
		NotLoadCacheAtStart: true,
		NamespaceId:         nc.Namespace,
		TimeoutMs:           10 * 1000, // 默认10秒
		LogLevel:            "error",
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
	}

	// 超时时间
	if nc.Timeout != nil {
		if d := nc.Timeout.AsDuration(); d.Milliseconds() > 0 {
			clientCfg.TimeoutMs = uint64(d.Milliseconds())
		}
	}

	// 如果同时配置了 ak/sk 则用户名/密码失效
	if nc.AccessKey != "" && nc.SecretKey != "" {
		clientCfg.OpenKMS = true
		clientCfg.AccessKey = nc.AccessKey
		clientCfg.SecretKey = nc.SecretKey
		clientCfg.RegionId = nc.RegionID
	} else {
		clientCfg.Username = nc.UserName
		clientCfg.Password = nc.Password
	}

	if nc.LogDir != nil {
		clientCfg.LogDir = *nc.LogDir
	}

	if nc.CacheDir != nil {
		clientCfg.CacheDir = *nc.CacheDir
	}

	return clientCfg
}
