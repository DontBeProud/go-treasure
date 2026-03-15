package gtknacos

import (
	gtnacos "github.com/DontBeProud/go-treasure/gt-nacos"
	gtconfpb "github.com/DontBeProud/go-treasure/pb/gt-conf-pb"
)

// NewKratosRegistryClient 用于兼容kratos接口
func NewKratosRegistryClient(client gtnacos.RegistryCenterClient, conf *gtconfpb.NacosRegistryConfig) *Registry {
	opts := make([]Option, 0)
	if conf != nil {
		if conf.Group != "" {
			opts = append(opts, WithGroup(conf.Group))
		}
		if conf.Prefix != "" {
			opts = append(opts, WithPrefix(conf.Prefix))
		}
		if conf.Cluster != "" {
			opts = append(opts, WithCluster(conf.Cluster))
		}
		if conf.Weight != 0 {
			opts = append(opts, WithWeight(float64(conf.Weight)))
		}
		if conf.Kind != "" {
			opts = append(opts, WithDefaultKind(conf.Kind))
		}
	}
	return New(client.GetClient(), opts...)
}
