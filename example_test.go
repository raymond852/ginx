package ginx

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type exampleAPIIF interface {
	API
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
	DocDefineRef("#EnumTest2Field", []interface{}{"val3", "val4"})
	DocDefineRef("#DescTest2Field", "test2 field")
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
	Test2      string  `json:"test2" doc:"required enum(#EnumTest2Field) desc(#DescTest2Field)"`
	Number     float64 `json:"number" doc:"desc(number test)"`
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

func (e exampleAPI) RouteGroup() *RouteGroup {
	docTag := "example"
	rg := NewRouteGroup("/test")

	rg.
		Put("/json/:%s", "path").
		To(e.JSONTest).
		Doc().
		Summary("json test").
		Description("json test description").
		Tag(docTag).
		Header(Header("Test-Header").Schema("val1").Enum("val1", "val2").Description("Testing header")).
		Query(Query("t").Schema("test").Required(false).Description("Testing query")).
		Path(Path("path").Schema("path-test").Description("Testing path")).
		RequestBody(JSONRequestBody(JSONRequest{
			JSONEmbed: JSONEmbed{
				Str:   "test",
				Array: []string{"val1"},
			},
			EmailField: "raymond852@example.com",
			OuterStr:   "val1",
			Test2:      "val3",
			Number:     10,
			Integer:    10,
			Integer32:  32,
			Integer64:  64,
		}).Required(true).Description("test JSON request body")).
		Response("200", JSONResponseBody(JSONResponse{
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
		RequestBody(MultiPartFormRequestBody(MultipartFormRequest{
			Number:    10,
			Integer:   100,
			File:      multipart.FileHeader{},
			BoolField: true,
		})).
		Response("200", TextResponseBody("success"), "success")

	rg.
		Patch("/").
		To(e.FormUrlencodedTest).
		Doc().
		Summary("urlencoded form test").
		Description("urlencoded form test description").
		Tag(docTag).
		RequestBody(UrlEncodedFormRequestBody(UrlencodedFormRequest{
			StringField: "string field",
			BoolField:   false,
			Number:      1,
			Integer:     2,
		})).
		Response("200", TextResponseBody("success"), "success")

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

func TestExampleAPI_JSONTest(t *testing.T) {
	g := gin.New()
	Init("example description", "1.0.0", "An example title")
	UseSwaggerUI(g, "/apidoc")
	UseValidator(g, func(ctx *gin.Context, err error) {
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
	})
	AddAPI(g, NewExampleAPI())

	go func() {
		srv := &http.Server{
			Addr:    ":5698",
			Handler: g,
		}
		_ = srv.ListenAndServe()
	}()

	<-time.After(1 * time.Second)
	req, _ := http.NewRequest(http.MethodPut, `http://127.0.0.1:5698/test/json/test-json?t=test`, bytes.NewBuffer([]byte(
		`{
			"emailField": "raymond@test.com",
  			"str": "test",
			"test2": "val3",
			"arr": [
				"val1"
			],
  			"outerstr": "val1",
 	 		"number": 10,
  			"integer": 12
		}`)))
	req.Header.Set("Test-Header", "val1")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	assertNil(t, err)
	assertEqual(t, resp.StatusCode, http.StatusOK)
}

func TestExampleAPI_FormTest(t *testing.T) {
	g := gin.New()
	Init("example description", "1.0.0", "An example title")
	UseSwaggerUI(g, "/apidoc")
	UseValidator(g, func(ctx *gin.Context, err error) {
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
	})
	AddAPI(g, NewExampleAPI())

	go func() {
		srv := &http.Server{
			Addr:    ":5699",
			Handler: g,
		}
		_ = srv.ListenAndServe()
	}()

	<-time.After(1 * time.Second)
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)

	_ = w.WriteField("number", "1")
	_ = w.WriteField("integer", "2")
	_ = w.WriteField("boolField", "true")
	f, _ := w.CreateFormFile("fileField", "test.txt")
	_, _ = f.Write([]byte("test"))
	_ = w.Close()
	req, _ := http.NewRequest(http.MethodPost, `http://127.0.0.1:5699/test`, buf)
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	assertNil(t, err)
	assertEqual(t, resp.StatusCode, http.StatusOK)
}

func TestExampleAPI_FormUrlencodedTest(t *testing.T) {
	g := gin.New()
	Init("example description", "1.0.0", "An example title")
	UseValidator(g, func(ctx *gin.Context, err error) {
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusBadRequest, map[string]interface{}{
				"error": err.Error(),
			})
		}
	})
	UseSwaggerUI(g, "/apidoc")
	AddAPI(g, NewExampleAPI())

	go func() {
		srv := &http.Server{
			Addr:    ":5700",
			Handler: g,
		}
		_ = srv.ListenAndServe()
	}()

	<-time.After(1 * time.Second)

	data := url.Values{}
	data.Set("stringField", "foo")
	data.Set("boolField", "true")
	data.Set("number", "1.5")
	data.Set("integer", "1")

	client := &http.Client{}
	r, _ := http.NewRequest(http.MethodPatch, "http://127.0.0.1:5700/test", strings.NewReader(data.Encode())) // URL-encoded payload
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(r)
	assertNil(t, err)
	assertEqual(t, resp.StatusCode, http.StatusOK)
}
