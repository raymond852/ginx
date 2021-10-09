package ginx

import (
	"errors"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gin-gonic/gin"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"
)

var swaggerPathPrefix string

func UseOpenTracing(engine *gin.Engine, tracer opentracing.Tracer) {
	engine.Use(func(c *gin.Context) {
		carrier := opentracing.HTTPHeadersCarrier(c.Request.Header)
		ctx, _ := tracer.Extract(opentracing.HTTPHeaders, carrier)
		sp := tracer.StartSpan(c.Request.Method + " " + c.FullPath(), ext.RPCServerOption(ctx))
		ext.HTTPMethod.Set(sp, c.Request.Method)
		ext.HTTPUrl.Set(sp, c.FullPath())
		ext.SpanKind.Set(sp, "api")
		sp.SetTag("client.ip", c.ClientIP())
		c.Request = c.Request.WithContext(opentracing.ContextWithSpan(c.Request.Context(), sp))

		c.Next()

		ext.HTTPStatusCode.Set(sp, uint16(c.Writer.Status()))
		sp.Finish()
	})
}

func UseSwaggerUI(engine *gin.Engine, swaggerPath string) {
	if !strings.HasPrefix(swaggerPath, "/") {
		swaggerPath = "/" + swaggerPath
	}
	engine.Use(func(context *gin.Context) {
		if context.Request.URL.Path == path.Join(swaggerPath, "swagger.json") && context.Request.Method == http.MethodGet {
			context.AbortWithStatusJSON(http.StatusOK, DocRoot)
		}
		context.Next()
	})
	engine.StaticFS(swaggerPath, AssetFile())
	swaggerPathPrefix = swaggerPath
}

func UseValidator(engine *gin.Engine, validationErrorHandler func(*gin.Context, error)) {
	var router *openapi3filter.Router
	decodeBody := func(body io.Reader, header http.Header, schema *openapi3.SchemaRef, encFn openapi3filter.EncodingFn) (interface{}, error) {
		data, err := ioutil.ReadAll(body)
		if err != nil {
			return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindInvalidFormat, Cause: err}
		}
		switch schema.Value.Type {
		case "integer", "number":
			v, err := strconv.ParseFloat(string(data), 64)
			if err != nil {
				return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindInvalidFormat, Value: string(data), Reason: "an invalid integer", Cause: err}
			}
			return v, nil
		case "boolean":
			v, err := strconv.ParseBool(string(data))
			if err != nil {
				return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindInvalidFormat, Value: string(data), Reason: "an invalid boolean", Cause: err}
			}
			return v, nil
		default:
			return string(data), nil
		}
	}

	// Override the multipart/form-data so that the validation will not based on the content-type of the form field
	openapi3filter.RegisterBodyDecoder("multipart/form-data", func(body io.Reader, header http.Header, schema *openapi3.SchemaRef, encFn openapi3filter.EncodingFn) (interface{}, error) {
		if schema.Value.Type != "object" {
			return nil, errors.New("unsupported JSON schema of request body")
		}

		// Parse form.
		values := make(map[string][]interface{})
		contentType := header.Get(http.CanonicalHeaderKey("Content-Type"))
		_, params, err := mime.ParseMediaType(contentType)
		if err != nil {
			return nil, err
		}
		mr := multipart.NewReader(body, params["boundary"])
		for {
			var part *multipart.Part
			if part, err = mr.NextPart(); err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}

			var (
				name = part.FormName()
				enc  *openapi3.Encoding
			)
			if encFn != nil {
				enc = encFn(name)
			}
			subEncFn := func(string) *openapi3.Encoding { return enc }
			// If the property's schema has type "array" it is means that the form contains a few parts with the same name.
			// Every such part has a type that is defined by an items schema in the property's schema.
			var valueSchema *openapi3.SchemaRef
			var exists bool
			valueSchema, exists = schema.Value.Properties[name]
			if !exists {
				anyProperties := schema.Value.AdditionalPropertiesAllowed
				if anyProperties != nil {
					switch *anyProperties {
					case true:
						//additionalProperties: true
						continue
					default:
						//additionalProperties: false
						return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindOther, Cause: fmt.Errorf("part %s: undefined", name)}
					}
				}
				if schema.Value.AdditionalProperties == nil {
					return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindOther, Cause: fmt.Errorf("part %s: undefined", name)}
				}
				valueSchema, exists = schema.Value.AdditionalProperties.Value.Properties[name]
				if !exists {
					return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindOther, Cause: fmt.Errorf("part %s: undefined", name)}
				}
			}
			if valueSchema.Value.Type == "array" {
				valueSchema = valueSchema.Value.Items
			}

			var value interface{}
			if value, err = decodeBody(part, http.Header(part.Header), valueSchema, subEncFn); err != nil {
				return nil, err
			}
			values[name] = append(values[name], value)
		}

		allTheProperties := make(map[string]*openapi3.SchemaRef)
		for k, v := range schema.Value.Properties {
			allTheProperties[k] = v
		}
		if schema.Value.AdditionalProperties != nil {
			for k, v := range schema.Value.AdditionalProperties.Value.Properties {
				allTheProperties[k] = v
			}
		}
		// Make an object value from form values.
		obj := make(map[string]interface{})
		for name, prop := range allTheProperties {
			vv := values[name]
			if len(vv) == 0 {
				continue
			}
			if prop.Value.Type == "array" {
				obj[name] = vv
			} else {
				obj[name] = vv[0]
			}
		}

		return obj, nil
	})

	engine.Use(func(c *gin.Context) {
		reqUrl := c.Request.URL

		if strings.HasPrefix(reqUrl.Path, swaggerPathPrefix) {
			return
		}
		method := c.Request.Method

		if router == nil {
			router = openapi3filter.NewRouter().WithSwagger(DocRoot)
		}
		route, pathParams, _ := router.FindRoute(method, reqUrl)

		requestValidationInput := &openapi3filter.RequestValidationInput{
			Request:    c.Request,
			PathParams: pathParams,
			Route:      route,
		}

		if err := openapi3filter.ValidateRequest(c, requestValidationInput); err != nil {
			validationErrorHandler(c, err)
		} else {
			validationErrorHandler(c, nil)
		}
		c.Next()
	})
}

func AddAPI(engine *gin.Engine, apiArgs ...API) {
	for _, api := range apiArgs {
		rg := api.RouteGroup().getRoutes()
		for _, r := range rg {
			engine.Handle(strings.ToUpper(r.httpMethod), r.httpPath, r.handlers...)
		}
	}
}
