package gtnacos

import (
	"sync"

	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
)

// RegistryCenterClient 注册中心客户端抽象类
type RegistryCenterClient interface {
	// GetRootClient 获取根客户端
	GetRootClient() Client

	// GetClient => naming_client.INamingClient
	GetClient() naming_client.INamingClient

	// CloseClient Close the GRPC client
	CloseClient()
}

type registryCenterClient struct {
	rootClient Client
	client     naming_client.INamingClient
	once       *sync.Once
}

// NewRegistryCenterClient 创建注册中心客户端
func (c *nacosClient) NewRegistryCenterClient() (client RegistryCenterClient, cleanup func(), err error) {
	param := c.GetClientParam()
	_client, err := clients.NewNamingClient(*param)
	if err != nil {
		return nil, func() {}, err
	}
	client = &registryCenterClient{
		rootClient: c,
		client:     _client,
		once:       &sync.Once{},
	}
	return client, func() {
		client.CloseClient()
	}, nil
}

// GetRootClient 获取根客户端
func (c *registryCenterClient) GetRootClient() Client {
	return c.rootClient
}

// GetClient => naming_client.INamingClient
func (c *registryCenterClient) GetClient() naming_client.INamingClient {
	return c.client
}

// CloseClient Close the GRPC client
func (c *registryCenterClient) CloseClient() {
	c.once.Do(c.client.CloseClient)
}
