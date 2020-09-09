package ginx

import (
	"fmt"
	"net/http"
	"path"

	"github.com/gin-gonic/gin"
)

type API interface {
	RouteGroup() *RouteGroup
}

type route struct {
	httpPath   string
	httpMethod string
	handlers   []gin.HandlerFunc
}

func NewRouteGroup(basePath string) *RouteGroup {
	return &RouteGroup{
		basePath: basePath,
	}
}

func (r *route) To(handlerFunc ...gin.HandlerFunc) *route {
	r.handlers = handlerFunc
	return r
}

func (r *route) Doc() *docPath {
	return newDocPath(r)
}

type RouteGroup struct {
	basePath string
	routes   []*route
}

func (rg *RouteGroup) add(httpPath, method string) *route {
	r := &route{
		httpPath:   httpPath,
		httpMethod: method,
	}
	rg.routes = append(rg.routes, r)
	return r
}

func (rg *RouteGroup) getRoutes() []*route {
	return rg.routes
}

func (rg *RouteGroup) resolvePath(relativePath string, pathParams ...interface{}) string {
	rp := fmt.Sprintf(relativePath, pathParams...)
	return path.Join(rg.basePath, rp)
}

func (rg *RouteGroup) Patch(path string, pathParams ...interface{}) *route {
	p := rg.resolvePath(path, pathParams...)
	return rg.add(p, http.MethodPatch)
}

func (rg *RouteGroup) PatchAbsolutePath(path string) *route {
	return rg.add(path, http.MethodPatch)
}

func (rg *RouteGroup) Delete(path string, pathParams ...interface{}) *route {
	p := rg.resolvePath(path, pathParams...)
	return rg.add(p, http.MethodDelete)
}

func (rg *RouteGroup) DeleteAbsolutePath(path string) *route {
	return rg.add(path, http.MethodDelete)
}

func (rg *RouteGroup) Put(path string, pathParams ...interface{}) *route {
	p := rg.resolvePath(path, pathParams...)
	return rg.add(p, http.MethodPut)
}

func (rg *RouteGroup) PutAbsolutePath(path string) *route {
	return rg.add(path, http.MethodPut)
}

func (rg *RouteGroup) Post(path string, pathParams ...interface{}) *route {
	p := rg.resolvePath(path, pathParams...)
	return rg.add(p, http.MethodPost)
}

func (rg *RouteGroup) PostAbsolutePath(path string) *route {
	return rg.add(path, http.MethodPost)
}

func (rg *RouteGroup) Get(path string, pathParams ...interface{}) *route {
	p := rg.resolvePath(path, pathParams...)
	return rg.add(p, http.MethodGet)
}

func (rg *RouteGroup) GetAbsolutePath(path string) *route {
	return rg.add(path, http.MethodGet)
}

func (rg *RouteGroup) Options(path string, pathParams ...interface{}) *route {
	p := rg.resolvePath(path, pathParams...)
	return rg.add(p, http.MethodOptions)
}

func (rg *RouteGroup) OptionsAbsolutePath(path string) *route {
	return rg.add(path, http.MethodOptions)
}

func (rg *RouteGroup) Trace(path string, pathParams ...interface{}) *route {
	p := rg.resolvePath(path, pathParams...)
	return rg.add(p, http.MethodTrace)
}

func (rg *RouteGroup) TraceAbsolutePath(path string) *route {
	return rg.add(path, http.MethodTrace)
}