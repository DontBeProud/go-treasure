package gtnacos

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"

	gtencode "github.com/DontBeProud/go-treasure/gt-encode"
	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/config_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
)

// ConfigChangeListener 配置变化的回调函数
type ConfigChangeListener func(namespace, group, dataId, data string)

// ConfigCenterClient 配置中心客户端抽象类
type ConfigCenterClient interface {
	// GetRootClient 获取根客户端
	GetRootClient() Client

	// GetConfig use to get config from nacos server
	// dataId  require
	// group   require
	// tenant ==>nacos.namespace optional
	GetConfig(conf *gtconfpb.NacosConfigCenterConfig) (string, error)

	// BindConfig use to get config from nacos server, and bind to the struct
	// if contentType is nil, use parsed contentType: etc. [fileName: aa.json] => [contentType: json]
	// dataId  require
	// group   require
	// tenant ==>nacos.namespace optional
	BindConfig(conf *gtconfpb.NacosConfigCenterConfig, v interface{}, contentType *string, ext *gtencode.Unmarshal2StructExt) error

	// PublishConfig use to publish config to nacos server
	// dataId  require
	// group   require
	// content require
	// tenant ==>nacos.namespace optional
	PublishConfig(conf *gtconfpb.NacosConfigCenterConfig, content string) (bool, error)

	// DeleteConfig use to delete config
	// dataId  require
	// group   require
	// tenant ==>nacos.namespace optional
	DeleteConfig(conf *gtconfpb.NacosConfigCenterConfig) (bool, error)

	// ListenConfig use to listen config change,it will callback OnChange() when config change
	// dataId  require
	// group   require
	// onchange require
	// tenant ==>nacos.namespace optional
	ListenConfig(conf *gtconfpb.NacosConfigCenterConfig, listener ConfigChangeListener) error

	// CancelListenConfig use to cancel listen config change
	// dataId  require
	// group   require
	// tenant ==>nacos.namespace optional
	CancelListenConfig(conf *gtconfpb.NacosConfigCenterConfig) error

	// SearchConfig use to search nacos config
	// search  require search=accurate--精确搜索  search=blur--模糊搜索
	// group   option
	// dataId  option
	// tenant ==>nacos.namespace optional
	// pageNo  option,default is 1
	// pageSize option,default is 10
	SearchConfig(searchType string, conf *gtconfpb.NacosConfigCenterConfig, page int, pageSize int) (*model.ConfigPage, error)

	// GetClient => config_client.IConfigClient
	GetClient() config_client.IConfigClient

	// CloseClient Close the GRPC client
	CloseClient()
}

// configCenterClient 配置中心客户端
type configCenterClient struct {
	rootClient Client
	client     config_client.IConfigClient
	once       *sync.Once
}

// NewConfigCenterClient 创建配置中心客户端
func (c *nacosClient) NewConfigCenterClient() (ConfigCenterClient, func(), error) {
	param := c.GetClientParam()
	client, err := clients.NewConfigClient(*param)
	if err != nil {
		return nil, func() {}, err
	}
	res := &configCenterClient{
		rootClient: c,
		client:     client,
		once:       &sync.Once{},
	}
	return res, func() {
		res.CloseClient()
	}, nil
}

// GetRootClient 获取根客户端
func (c *configCenterClient) GetRootClient() Client {
	return c.rootClient
}

// GetClient => config_client.IConfigClient
func (c *configCenterClient) GetClient() config_client.IConfigClient {
	return c.client
}

// GetConfig use to get config from nacos server
func (c *configCenterClient) GetConfig(conf *gtconfpb.NacosConfigCenterConfig) (string, error) {
	if conf == nil {
		return "", errors.New("get config conf is nil")
	}
	return c.client.GetConfig(parseConfigParam(conf))
}

// BindConfig use to get config from nacos server, and bind to the struct
func (c *configCenterClient) BindConfig(conf *gtconfpb.NacosConfigCenterConfig, v interface{}, contentType *string, ext *gtencode.Unmarshal2StructExt) error {
	if v == nil {
		return errors.New("bind config v is nil")
	}
	data, err := c.GetConfig(conf)
	if err != nil {
		return err
	}

	ct := ""
	if contentType != nil {
		ct = *contentType
	} else {
		ct = strings.TrimPrefix(filepath.Ext(conf.DataId), ".")
	}

	return gtencode.Unmarshal2Struct(ct, []byte(data), v, ext)
}

// PublishConfig use to publish config to nacos server
func (c *configCenterClient) PublishConfig(conf *gtconfpb.NacosConfigCenterConfig, content string) (bool, error) {
	p := parseConfigParam(conf)
	p.Content = content
	return c.client.PublishConfig(p)
}

// DeleteConfig use to delete config
func (c *configCenterClient) DeleteConfig(conf *gtconfpb.NacosConfigCenterConfig) (bool, error) {
	return c.client.DeleteConfig(parseConfigParam(conf))
}

// ListenConfig use to listen config change
func (c *configCenterClient) ListenConfig(conf *gtconfpb.NacosConfigCenterConfig, listener ConfigChangeListener) error {
	p := parseConfigParam(conf)
	p.OnChange = listener
	return c.client.ListenConfig(p)
}

// CancelListenConfig use to cancel listen config change
func (c *configCenterClient) CancelListenConfig(conf *gtconfpb.NacosConfigCenterConfig) error {
	return c.client.CancelListenConfig(parseConfigParam(conf))
}

// SearchConfig use to search nacos config
func (c *configCenterClient) SearchConfig(searchType string, conf *gtconfpb.NacosConfigCenterConfig, page int, pageSize int) (*model.ConfigPage, error) {
	if conf == nil {
		return nil, errors.New("search conf is nil")
	}
	return c.client.SearchConfig(vo.SearchConfigParam{
		Search:   searchType,
		DataId:   conf.DataId,
		Group:    conf.Group,
		Tag:      conf.Tag,
		AppName:  conf.AppName,
		PageNo:   page,
		PageSize: pageSize,
	})
}

// CloseClient Close the GRPC client
func (c *configCenterClient) CloseClient() {
	c.once.Do(c.client.CloseClient)
}

func parseConfigParam(conf *gtconfpb.NacosConfigCenterConfig) vo.ConfigParam {
	if conf == nil {
		return vo.ConfigParam{}
	}
	return vo.ConfigParam{
		DataId:           conf.DataId,
		Group:            conf.Group,
		Tag:              conf.Tag,
		ConfigTags:       conf.ConfigTags,
		AppName:          conf.AppName,
		BetaIps:          conf.BetaIps,
		CasMd5:           conf.CasMd5,
		Type:             conf.Type,
		SrcUser:          conf.SrcUser,
		EncryptedDataKey: conf.EncryptedDataKey,
		KmsKeyId:         conf.KmsKeyId,
	}
}
