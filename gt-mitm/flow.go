package gtmitm

import (
	"regexp"
	"sync"

	"github.com/lqqyt2423/go-mitmproxy/proxy"
)

// FlowHandlerFunc 处理函数类型
type FlowHandlerFunc func(*FlowContext)

// FlowRouter 路由器
type FlowRouter struct {
	globalHandler []FlowHandlerFunc
	routes        []*flowRouter
	mutex         sync.RWMutex
}

// FlowContext 基于 proxy.Flow 的上下文
type FlowContext struct {
	*proxy.Flow
	Params   map[string]string
	handlers []FlowHandlerFunc
	index    int
	abort    bool
}

// flowRouter 路由结构（添加 host 支持）
type flowRouter struct {
	method    string
	host      string
	hostRegex *regexp.Regexp // host 的正则表达式
	path      string
	regex     *regexp.Regexp
	handlers  []FlowHandlerFunc
}

// NewFlowRouter 创建新路由器
func NewFlowRouter() *FlowRouter {
	return &FlowRouter{
		routes:        make([]*flowRouter, 0),
		globalHandler: make([]FlowHandlerFunc, 0),
		mutex:         sync.RWMutex{},
	}
}

// Use 注册全局中间件
func (r *FlowRouter) Use(handlers ...FlowHandlerFunc) {
	r.globalHandler = append(r.globalHandler, handlers...)
}

// 添加路由（支持 host）
func (r *FlowRouter) addRoute(host, method, path string, handlers ...FlowHandlerFunc) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	pattern := pathToRegex(path)
	hostPattern := hostToRegex(host)

	route := &flowRouter{
		method:    method,
		host:      host,
		hostRegex: hostPattern,
		path:      path,
		regex:     pattern,
		handlers:  handlers,
	}

	r.routes = append(r.routes, route)
}

// ========== 带 Host 的路由方法 ==========

// GETWithHost 注册带 host 的 GET 路由
func (r *FlowRouter) GETWithHost(host, path string, handlers ...FlowHandlerFunc) {
	r.addRoute(host, "GET", path, handlers...)
}

// POSTWithHost 注册带 host 的 POST 路由
func (r *FlowRouter) POSTWithHost(host, path string, handlers ...FlowHandlerFunc) {
	r.addRoute(host, "POST", path, handlers...)
}

// PUTWithHost 注册带 host 的 PUT 路由
func (r *FlowRouter) PUTWithHost(host, path string, handlers ...FlowHandlerFunc) {
	r.addRoute(host, "PUT", path, handlers...)
}

// DELETEWithHost 注册带 host 的 DELETE 路由
func (r *FlowRouter) DELETEWithHost(host, path string, handlers ...FlowHandlerFunc) {
	r.addRoute(host, "DELETE", path, handlers...)
}

// PATCHWithHost 注册带 host 的 PATCH 路由
func (r *FlowRouter) PATCHWithHost(host, path string, handlers ...FlowHandlerFunc) {
	r.addRoute(host, "PATCH", path, handlers...)
}

// ANYWithHost 注册带 host 的任意方法路由
func (r *FlowRouter) ANYWithHost(host, path string, handlers ...FlowHandlerFunc) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		r.addRoute(host, method, path, handlers...)
	}
}

// ========== 原有路由方法（兼容性，匹配所有 host） ==========

// GET 注册 GET 路由（匹配所有 host）
func (r *FlowRouter) GET(path string, handlers ...FlowHandlerFunc) {
	r.addRoute("", "GET", path, handlers...)
}

// POST 注册 POST 路由（匹配所有 host）
func (r *FlowRouter) POST(path string, handlers ...FlowHandlerFunc) {
	r.addRoute("", "POST", path, handlers...)
}

// PUT 注册 PUT 路由（匹配所有 host）
func (r *FlowRouter) PUT(path string, handlers ...FlowHandlerFunc) {
	r.addRoute("", "PUT", path, handlers...)
}

// DELETE 注册 DELETE 路由（匹配所有 host）
func (r *FlowRouter) DELETE(path string, handlers ...FlowHandlerFunc) {
	r.addRoute("", "DELETE", path, handlers...)
}

// PATCH 注册 PATCH 路由（匹配所有 host）
func (r *FlowRouter) PATCH(path string, handlers ...FlowHandlerFunc) {
	r.addRoute("", "PATCH", path, handlers...)
}

// ANY 注册任意方法路由（匹配所有 host）
func (r *FlowRouter) ANY(path string, handlers ...FlowHandlerFunc) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		r.addRoute("", method, path, handlers...)
	}
}

// ========== 工具函数 ==========

// 路径转正则表达式
func pathToRegex(path string) *regexp.Regexp {
	pattern := "^" + path + "$"
	// 替换 :param 为 ([^/]+)
	pattern = regexp.MustCompile(`:([^/]+)`).ReplaceAllString(pattern, `([^/]+)`)
	// 替换 *param 为 (.+)
	pattern = regexp.MustCompile(`\*([^/]+)`).ReplaceAllString(pattern, `(.+)`)
	return regexp.MustCompile(pattern)
}

// host 转正则表达式
func hostToRegex(host string) *regexp.Regexp {
	if host == "" {
		return regexp.MustCompile(".*")
	}
	pattern := "^" + host + "$"
	// 替换 * 为 [^.]* （通配符）
	pattern = regexp.MustCompile(`\*`).ReplaceAllString(pattern, `[^.]*`)
	// 替换 :param 为 ([^.]+)
	pattern = regexp.MustCompile(`:([^.]+)`).ReplaceAllString(pattern, `([^.]+)`)
	return regexp.MustCompile(pattern)
}

// 匹配路由（添加 host 匹配）
func (r *FlowRouter) matchRoute(host, method, path string) (*flowRouter, map[string]string) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	for _, route := range r.routes {
		// 匹配 host
		if !route.hostRegex.MatchString(host) {
			continue
		}

		// 匹配 method
		if route.method != method && route.method != "ANY" {
			continue
		}

		// 匹配 path
		if matches := route.regex.FindStringSubmatch(path); matches != nil {
			params := make(map[string]string)

			paramNames := extractParamNames(route.path)
			for i, name := range paramNames {
				if i+1 < len(matches) {
					params[name] = matches[i+1]
				}
			}

			return route, params
		}
	}

	return nil, nil
}

// 提取参数名
func extractParamNames(path string) []string {
	var names []string

	colonRegex := regexp.MustCompile(`:([^/]+)`)
	colonMatches := colonRegex.FindAllStringSubmatch(path, -1)
	for _, match := range colonMatches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}

	starRegex := regexp.MustCompile(`\*([^/]+)`)
	starMatches := starRegex.FindAllStringSubmatch(path, -1)
	for _, match := range starMatches {
		if len(match) > 1 {
			names = append(names, match[1])
		}
	}

	return names
}

// HandleFlow 处理
func (r *FlowRouter) HandleFlow(flow *proxy.Flow) {
	method := flow.Request.Method
	path := flow.Request.URL.Path
	host := flow.Request.URL.Hostname()

	handlers := make([]FlowHandlerFunc, len(r.globalHandler))
	copy(handlers, r.globalHandler)
	route, params := r.matchRoute(host, method, path)
	if route != nil {
		handlers = append(handlers, route.handlers...)
	} else {
		params = map[string]string{}
	}

	ctx := &FlowContext{
		Flow:     flow,
		Params:   params,
		handlers: handlers,
		index:    -1,
	}

	ctx.Next()
}

// ========== 上下文方法 ==========

// Next 执行下一个处理器
func (c *FlowContext) Next() {
	c.index++

	for c.index < len(c.handlers) && !c.abort {
		c.handlers[c.index](c)
		c.index++
	}
}

// Abort 中止处理链
func (c *FlowContext) Abort() {
	c.abort = true
}

// Param 获取路径参数
func (c *FlowContext) Param(key string) string {
	return c.Params[key]
}

// ========== 路由组支持 ==========

// FlowGroup 路由组
type FlowGroup struct {
	router   *FlowRouter
	host     string
	prefix   string
	handlers []FlowHandlerFunc
}

// Group 创建路由组（不带 host）
func (r *FlowRouter) Group(prefix string) *FlowGroup {
	return &FlowGroup{
		router: r,
		prefix: prefix,
	}
}

// GroupWithHost 创建带 host 的路由组
func (r *FlowRouter) GroupWithHost(host, prefix string) *FlowGroup {
	return &FlowGroup{
		router: r,
		host:   host,
		prefix: prefix,
	}
}

// Use 为路由组添加中间件
func (g *FlowGroup) Use(middleware ...FlowHandlerFunc) {
	g.handlers = append(g.handlers, middleware...)
}

// 添加组路由
func (g *FlowGroup) addRoute(method, path string, handlers ...FlowHandlerFunc) {
	fullPath := g.prefix + path
	allHandlers := append(append([]FlowHandlerFunc{}, g.handlers...), handlers...)
	g.router.addRoute(g.host, method, fullPath, allHandlers...)
}

// GET 组GET路由
func (g *FlowGroup) GET(path string, handlers ...FlowHandlerFunc) {
	g.addRoute("GET", path, handlers...)
}

// POST 组POST路由
func (g *FlowGroup) POST(path string, handlers ...FlowHandlerFunc) {
	g.addRoute("POST", path, handlers...)
}

// PUT 组PUT路由
func (g *FlowGroup) PUT(path string, handlers ...FlowHandlerFunc) {
	g.addRoute("PUT", path, handlers...)
}

// DELETE 组DELETE路由
func (g *FlowGroup) DELETE(path string, handlers ...FlowHandlerFunc) {
	g.addRoute("DELETE", path, handlers...)
}

// PATCH 组PATCH路由
func (g *FlowGroup) PATCH(path string, handlers ...FlowHandlerFunc) {
	g.addRoute("PATCH", path, handlers...)
}

// ANY 组任意方法路由
func (g *FlowGroup) ANY(path string, handlers ...FlowHandlerFunc) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"}
	for _, method := range methods {
		g.addRoute(method, path, handlers...)
	}
}
