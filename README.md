### Introduction
Maintaining API documentation and source code could be a cumbersome task

ginx is trying to solve the above problem by clear doc generation such that allowing developer to accelerate API prototyping

ginx is a lightweight extension to [gin](https://github.com/gin-gonic/gin), which allows REST style API creation with [OpenAPI 3.0](https://github.com/getkin/kin-openapi). 
 
The library is inspired by [go-restful](https://github.com/emicklei/go-restful)

### Example
example.go
```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/raymond852/ginx"
    "mime/multipart"
    "net/http"
)

type exampleAPIIF interface {
	ginx.API
	JSONTest(ctx *gin.Context)
	MultiPartFormTest(ctx *gin.Context)
	FormUrlencodedTest(ctx *gin.Context)
}

func NewExampleAPI() exampleAPIIF {
	return &exampleAPI{}
}

type exampleAPI struct {
}

func init() {
	ginx.DocDefineRef("#EnumTest2Field", []interface{}{"val3", "val4"})
	ginx.DocDefineRef("#DescTest2Field", "test2 field")
}

type MultipartFormRequest struct {
	BoolField bool                 `form:"boolField" doc:"required desc(bool test)"`
	Number    float64              `form:"number" doc:"required desc(number test)"`
	Integer   int                  `form:"integer" doc:"required desc(integer test)"`
	File      multipart.FileHeader `form:"fileField" doc:"required desc(file to upload)"`
}

type UrlencodedFormRequest struct {
	StringField string  `form:"stringField" doc:"required desc(string field)"`
	BoolField   bool    `form:"boolField" doc:"required desc(bool test)"`
	Number      float64 `form:"number" doc:"required desc(number test)"`
	Integer     int     `form:"integer" doc:"required desc(integer test)"`
}

type JSONRequest struct {
	JSONEmbed
    EmailField string  `json:"emailField" doc:"required format(email)"`
	OuterStr   string  `json:"outerstr" doc:"required enum(val1;val2)"`
	Test2      string  `json:"test2" doc:"enum(#EnumTest2Field) desc(#DescTest2Field)"`
	Number     float64 `json:"number" doc:"required desc(number test)"`
	Integer    int     `json:"integer" doc:"required desc(integer test) minimum(5) maximum(100)"`
	Integer32  int32   `json:"integer32" doc:"desc(integer test)"`
	Integer64  int64   `json:"integer64" doc:"desc(integer test)"`
}

type JSONEmbed struct {
	Str   string   `json:"str" doc:"required desc(number test) pattern([^abc]+) maxLength(10) minLength(1)"`
	Array []string `json:"arr" doc:"desc(array) minItems(1) maxItems(5)"`
}

type JSONResponse struct {
	FieldOne string `json:"field1" doc:"desc(example field one)"`
	FieldTwo bool   `json:"field2" doc:"desc(example field 2)"`
}

func (e exampleAPI) RouteGroup() *ginx.RouteGroup {
	docTag := "example"
	rg := ginx.NewRouteGroup("/test")

	rg.
		Put("/json/:%s", "path").
		To(e.JSONTest).
		Doc().
		Summary("json test").
		Description("json test description").
		Tag(docTag).
		Header(ginx.Header("Test-Header").Schema("val1").Enum("val1", "val2").Description("Testing header")).
		Query(ginx.Query("t").Schema("test").Required(false).Description("Testing query")).
		Path(ginx.Path("path").Schema("path-test").Description("Testing path")).
		RequestBody(ginx.JSONRequestBody(JSONRequest{
            EmailField: "raymond852@example.com",
			JSONEmbed: JSONEmbed{
				Str:   "test",
				Array: []string{"val1"},
			},
			OuterStr: "val1",
			Test2: "val3",
			Number:   10,
			Integer:  12,
		}).Required(true).Description("test JSON request body")).
		Response("200", ginx.JSONResponseBody(JSONResponse{
			FieldOne: "abc",
			FieldTwo: false,
		}), "success")

	rg.
		Post("/").
		To(e.MultiPartFormTest).
		Doc().
		Summary("multipart form test").
		Description("multipart form test description").
		Tag(docTag).
		RequestBody(ginx.MultiPartFormRequestBody(MultipartFormRequest{
			Number:    10,
			Integer:   100,
			File:      multipart.FileHeader{},
			BoolField: true,
		})).
		Response("200", ginx.TextResponseBody("success"), "success")

	rg.
		Patch("/").
		To(e.FormUrlencodedTest).
		Doc().
		Summary("urlencoded form test").
		Description("urlencoded form test description").
		Tag(docTag).
		RequestBody(ginx.UrlEncodedFormRequestBody(UrlencodedFormRequest{
			StringField: "string field",
			BoolField:   false,
			Number:      1,
			Integer:     2,
		})).
		Response("200", ginx.TextResponseBody("success"), "success")

	return rg
}

func (e exampleAPI) JSONTest(ctx *gin.Context) {
	ctx.AbortWithStatusJSON(http.StatusOK, JSONResponse{
		FieldOne: "fieldOne",
		FieldTwo: false,
	})
}

func (e exampleAPI) MultiPartFormTest(ctx *gin.Context) {
	req := &MultipartFormRequest{}
	if err := ctx.ShouldBind(req); err != nil {
		ctx.String(http.StatusBadRequest, "bad request")
		ctx.Abort()
	} else {
		ctx.String(http.StatusOK, "success")
		ctx.Abort()
	}
}

func (e exampleAPI) FormUrlencodedTest(ctx *gin.Context) {
	req := &UrlencodedFormRequest{}
	if err := ctx.ShouldBind(req); err != nil {
		ctx.String(http.StatusBadRequest, "bad request")
		ctx.Abort()
	} else {
		ctx.String(http.StatusOK, "success")
		ctx.Abort()
	}
}
```
main.go
```go
package main

import (
    "github.com/raymond852/ginx"
    "github.com/gin-gonic/gin"
    "mime/multipart"
    "net/http"
    "net/url"
    "strings"
    "testing"
    "time"
    "bytes"
    "net/http"
)

func main(){
	g := gin.New()
	ginx.Init("example description", "1.0.0", "An example title")
	ginx.UseSwaggerUI(g, "/apidoc")
	ginx.UseValidator(g, func(ctx *gin.Context, err error) {
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
	})
	ginx.AddAPI(g, NewExampleAPI())

	srv := &http.Server{
		Addr:    ":5699",
		Handler: g,
	}
	_ = srv.ListenAndServe()

}
```
#### Swagger UI gin middleware

`gin.UseSwaggerUI` 

open browser and go to http://localhost:5699/apidoc and you will able to browse the Swagger doc

#### OpenAPI 3.0 request validator gin middleware

`ginx.UseValidator`

the middleware will validate the request against the OpenAPI 3.0 spec

### Features
1. Tag for generate request body and response body which align with openapi 3.0 spec
2. Request validation
3. Open API 3.0 doc generation and preview using swagger UI