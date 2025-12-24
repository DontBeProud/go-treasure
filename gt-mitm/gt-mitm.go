package gtmitm

import (
	"github.com/lqqyt2423/go-mitmproxy/proxy"
	log "github.com/sirupsen/logrus"
)

func init() {
	// 默认关闭GO-Mitm-Proxy原生库的日志
	SetMitmProxyLogLevel(log.PanicLevel)
}

// SetMitmProxyLogLevel 设置GO-Mitm-Proxy原生库的日志级别
func SetMitmProxyLogLevel(level log.Level) {
	log.SetLevel(level)
}

type MitmProxy interface {
	Run() error
	GetRequestRouter() *FlowRouter
	GetResponseRouter() *FlowRouter
}

type mitmProxy struct {
	proxy.BaseAddon
	proxyObj       *proxy.Proxy
	requestRouter  *FlowRouter
	responseRouter *FlowRouter
}

func NewMitmProxy(opt *proxy.Options) (MitmProxy, error) {
	obj, err := proxy.NewProxy(opt)
	if err != nil {
		return nil, err
	}
	res := &mitmProxy{
		proxyObj:       obj,
		requestRouter:  NewFlowRouter(),
		responseRouter: NewFlowRouter(),
	}
	res.proxyObj.AddAddon(res)
	return res, nil
}

func (c *mitmProxy) Run() error {
	return c.proxyObj.Start()
}

func (c *mitmProxy) GetRequestRouter() *FlowRouter {
	return c.requestRouter
}

func (c *mitmProxy) GetResponseRouter() *FlowRouter {
	return c.responseRouter
}

func (c *mitmProxy) Request(f *proxy.Flow) {
	c.requestRouter.HandleFlow(f)
}

func (c *mitmProxy) Response(f *proxy.Flow) {
	c.responseRouter.HandleFlow(f)
}
