package ginx

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"testing"
)

func TestRouteGroup_Delete(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.Delete("delete").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/test/delete")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodDelete)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_DeleteAbsolutePath(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.DeleteAbsolutePath("/delete").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/delete")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodDelete)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_Patch(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.Patch("patch").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/test/patch")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodPatch)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_PatchAbsolutePath(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.PatchAbsolutePath("/patch").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/patch")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodPatch)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_Put(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.Put("put").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/test/put")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodPut)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_PutAbsolutePath(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.PutAbsolutePath("/put").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/put")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodPut)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_Get(t *testing.T) {
	underTest := NewRouteGroup("/test")
	pathParam := 1
	underTest.Get("get/%d", pathParam).To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/test/get/1")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodGet)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_GetAbsolutePath(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.GetAbsolutePath("/get").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/get")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodGet)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_Post(t *testing.T) {
	underTest := NewRouteGroup("/test")
	pathParam := "param"
	underTest.Post("post/:%s", pathParam).To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/test/post/:param")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodPost)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_PostAbsolutePath(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.PostAbsolutePath("/post").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/post")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodPost)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_Options(t *testing.T) {
	underTest := NewRouteGroup("/test")
	pathParam := 1
	underTest.Options("options/%d", pathParam).To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/test/options/1")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodOptions)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_OptionsAbsolutePath(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.OptionsAbsolutePath("/options").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/options")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodOptions)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_Trace(t *testing.T) {
	underTest := NewRouteGroup("/test")
	pathParam := 1
	underTest.Trace("trace/%d", pathParam).To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/test/trace/1")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodTrace)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}

func TestRouteGroup_TraceAbsolutePath(t *testing.T) {
	underTest := NewRouteGroup("/test")
	underTest.TraceAbsolutePath("/trace").To(func(context *gin.Context) {})
	assertEqual(t, 1, len(underTest.getRoutes()))
	assertEqual(t, underTest.routes[0].httpPath, "/trace")
	assertEqual(t, underTest.routes[0].httpMethod, http.MethodTrace)
	assertEqual(t, len(underTest.routes[0].handlers), 1)
}