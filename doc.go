package ginx

import (
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const (
	ParamQuery  = "query"
	ParamPath   = "path"
	ParamHeader = "header"

	RootTag = "doc"

	DocTagFieldRequired        = "required"
	DocTagFieldFormat          = "format"
	DocTagFieldPattern         = "pattern"
	DocTagFieldDescription     = "desc"
	DocTagFieldEnum            = "enum"
	DocTagFieldNullable        = "nullable"
	DocTagFieldMinItems        = "minItems"
	DocTagFieldMaxItems        = "maxItems"
	DocTagFieldMaximum         = "maximum"
	DocTagFieldMinimum         = "minimum"
	DocTagFieldStringMaxLength = "maxLength"
	DocTagFieldStringMinLength = "minLength"
)

var DocRoot *openapi3.T

func Init(description, version, title string) {
	DocRoot = &openapi3.T{
		OpenAPI: "3.0.0",
		Info: &openapi3.Info{
			Description: description,
			Version:     version,
			Title:       title,
		},
		Paths: openapi3.NewPaths(),
	}
}

var Refs = make(map[string]interface{})

var ginToOpenAPIPathPattern = regexp.MustCompile(`/(:)([^/]+)`)

func DocDefineRef(key string, val interface{}) {
	if _, exists := Refs[key]; exists {
		panic(fmt.Sprintf("docRef key=%s already exists", key))
	}
	Refs[key] = val
}

func Header(name string) *docParam {
	return &docParam{name: name, required: true}
}

func Path(name string) *docParam {
	return &docParam{name: name, required: true}
}

func Query(name string) *docParam {
	return &docParam{name: name, required: true}
}

func MultiPartFormRequestBody(prototype interface{}) *docRequestBody {
	schemeRef := extractSchema(prototype)
	return &docRequestBody{
		required: true,
		contents: map[string]*openapi3.SchemaRef{
			"multipart/form-data": schemeRef,
		},
	}
}

func UrlEncodedFormRequestBody(prototype interface{}) *docRequestBody {
	schemeRef := extractSchema(prototype)
	return &docRequestBody{
		required: true,
		contents: map[string]*openapi3.SchemaRef{
			"application/x-www-form-urlencoded": schemeRef,
		},
	}
}

func JSONRequestBody(prototype interface{}) *docRequestBody {
	schemeRef := extractSchema(prototype)
	if b, err := json.Marshal(prototype); err == nil {
		var obj map[string]interface{}
		if err = json.Unmarshal(b, &obj); err == nil {
			schemeRef.Value.Example = obj
		}
	}
	return &docRequestBody{
		required: true,
		contents: map[string]*openapi3.SchemaRef{
			"application/json": schemeRef,
		},
	}
}

func JSONResponseBody(prototype interface{}) *docResponse {
	schemeRef := extractSchema(prototype)
	if b, err := json.Marshal(prototype); err == nil {
		var inInterface map[string]interface{}
		if err = json.Unmarshal(b, &inInterface); err == nil {
			schemeRef.Value.Example = inInterface
		}
	}
	return &docResponse{
		contents: map[string]*openapi3.SchemaRef{
			"application/json": schemeRef,
		},
	}
}

func TextResponseBody(text string) *docResponse {
	return &docResponse{
		contents: map[string]*openapi3.SchemaRef{
			"text/plain": extractSchema(text),
		},
	}
}

func CSVResponseBody(text string) *docResponse {
	return &docResponse{
		contents: map[string]*openapi3.SchemaRef{
			"text/csv": extractSchema(text),
		},
	}
}

func PDFResponseBody() *docResponse {
	return &docResponse{
		contents: map[string]*openapi3.SchemaRef{
			"application/pdf": extractSchema(""),
		},
	}
}

func newDocPath(route *route) *docPath {
	var p *openapi3.PathItem
	ph := ginToOpenAPIPathPattern.ReplaceAllString(route.httpPath, `/{$2}`)
	if pathItemObj := DocRoot.Paths.Value(ph); pathItemObj != nil {
		p = pathItemObj
	} else {
		p = &openapi3.PathItem{}
	}
	op := &openapi3.Operation{}
	op.Responses = openapi3.NewResponses()
	switch route.httpMethod {
	case http.MethodGet:
		p.Get = op
	case http.MethodPut:
		p.Put = op
	case http.MethodPost:
		p.Post = op
	case http.MethodDelete:
		p.Delete = op
	case http.MethodOptions:
		p.Options = op
	case http.MethodHead:
		p.Head = op
	case http.MethodPatch:
		p.Patch = op
	case http.MethodTrace:
		p.Trace = op
	}
	DocRoot.Paths.Set(ph, p)
	return &docPath{p, op}
}

type docPath struct {
	pathItem  *openapi3.PathItem
	operation *openapi3.Operation
}

func (d *docPath) Tag(tag string) *docPath {
	d.operation.Tags = append(d.operation.Tags, tag)
	return d
}

func (d *docPath) Summary(summary string) *docPath {
	d.operation.Summary = summary
	return d
}

func (d *docPath) Description(description string) *docPath {
	d.operation.Description = description
	return d
}

func (d *docPath) Header(header *docParam) *docPath {
	d.operation.Parameters = append(d.operation.Parameters, header.ToOpenAPIParam(ParamHeader))
	return d
}

func (d *docPath) Path(path *docParam) *docPath {
	d.operation.Parameters = append(d.operation.Parameters, path.ToOpenAPIParam(ParamPath))
	return d
}

func (d *docPath) Query(query *docParam) *docPath {
	d.operation.Parameters = append(d.operation.Parameters, query.ToOpenAPIParam(ParamQuery))
	return d
}

func (d *docPath) RequestBody(body *docRequestBody) *docPath {
	d.operation.RequestBody = body.ToOpenAPIRequestBody()
	return d
}

func (d *docPath) Response(httpCode string, resp *docResponse, desc string) *docPath {
	if d.operation.Responses == nil {
		d.operation.Responses = openapi3.NewResponses()
	}

	resp.description = &desc
	d.operation.Responses.Set(httpCode, resp.ToOpenAPIResponse())
	return d
}

type docParam struct {
	name        string
	description string
	required    bool
	schemaRef   *openapi3.SchemaRef
}

func (d *docParam) Description(desc string) *docParam {
	d.description = desc
	return d
}

func (d *docParam) Required(required bool) *docParam {
	d.required = required
	return d
}

func (d *docParam) Schema(example interface{}) *docParam {
	d.schemaRef = extractSchema(example)
	return d
}

func (d *docParam) Enum(values ...interface{}) *docParam {
	if scheme := d.schemaRef.Value; scheme != nil {
		d.schemaRef.Value.Enum = values
		return d
	} else {
		panic("scheme is nil, must call Scheme(example interface{}) first")
	}
}

func (d *docParam) ToOpenAPIParam(in string) *openapi3.ParameterRef {
	return &openapi3.ParameterRef{
		Ref: "",
		Value: &openapi3.Parameter{
			Name:        d.name,
			In:          in,
			Description: d.description,
			Required:    d.required,
			Schema:      d.schemaRef,
		},
	}
}

type docRequestBody struct {
	description *string
	required    bool
	contents    map[string]*openapi3.SchemaRef
}

func (d *docRequestBody) Required(required bool) *docRequestBody {
	d.required = required
	return d
}

func (d *docRequestBody) Description(desc string) *docRequestBody {
	d.description = &desc
	return d
}

func (d *docRequestBody) ToOpenAPIRequestBody() *openapi3.RequestBodyRef {
	b := openapi3.NewRequestBody().WithRequired(d.required)
	if d.description != nil {
		b.WithDescription(*d.description)
	}
	b.Content = openapi3.NewContent()
	for mediaType, schemaRef := range d.contents {
		b.Content[mediaType] = &openapi3.MediaType{
			Schema: schemaRef,
		}
	}
	return &openapi3.RequestBodyRef{
		Ref:   "",
		Value: b,
	}
}

type docResponse struct {
	description *string
	contents    map[string]*openapi3.SchemaRef
}

func (d *docResponse) ToOpenAPIResponse() *openapi3.ResponseRef {
	b := openapi3.NewResponse()
	b.Content = openapi3.NewContent()
	b.Description = d.description
	for mediaType, schemaRef := range d.contents {
		b.Content[mediaType] = &openapi3.MediaType{
			Schema: schemaRef,
		}
	}
	return &openapi3.ResponseRef{
		Ref:   "",
		Value: b,
	}
}

func extractSchema(prototype interface{}) *openapi3.SchemaRef {
	if prototype == nil {
		return nil
	}
	if vh := extractFileHeader(prototype); vh != nil {
		sch := openapi3.NewStringSchema()
		sch.Format = "binary"
		return openapi3.NewSchemaRef("", sch)
	}
	if vh := extractString(prototype); vh != nil {
		sch := openapi3.NewStringSchema()
		if vh.Value.Kind() != reflect.Invalid {
			sch.Example = vh.Value.Interface()
		}
		return openapi3.NewSchemaRef("", sch)
	}
	if vh := extractInteger(prototype); vh != nil {
		sch := openapi3.NewIntegerSchema()
		if vh.Value.Kind() != reflect.Invalid {
			sch.Example = vh.Value.Interface()
		}
		return openapi3.NewSchemaRef("", sch)
	}
	if vh := extractNumber(prototype); vh != nil {
		sch := openapi3.NewFloat64Schema()
		if vh.Value.Kind() != reflect.Invalid {
			sch.Example = vh.Value.Interface()
		}
		return openapi3.NewSchemaRef("", sch)
	}
	if vh := extractBoolean(prototype); vh != nil {
		sch := openapi3.NewBoolSchema()
		if vh.Value.Kind() != reflect.Invalid {
			sch.Example = vh.Value.Interface()
		}
		return openapi3.NewSchemaRef("", sch)
	}
	if vh := extractMap(prototype); vh != nil {
		sch := openapi3.NewObjectSchema()
		if vh.Value.Kind() != reflect.Invalid {
			sch.Example = vh.Value.Interface()
		}
		return openapi3.NewSchemaRef("", sch)
	}
	if vh := extractArray(prototype); vh != nil {
		sch := openapi3.NewArraySchema()
		t := openapi3.Types([]string{"array"})
		sch.Type = &t
		if vh.Value.Len() == 0 {
			panic(fmt.Sprintf("prototype=%+v should has one or more elements in an array", prototype))
		}
		sch.Items = extractSchema(vh.Value.Index(0).Interface())
		return openapi3.NewSchemaRef("", sch)
	}
	if vh := extractStruct(prototype); vh != nil {
		objSch := openapi3.NewObjectSchema()
		objSch.Properties = make(map[string]*openapi3.SchemaRef)
		num := vh.Type.NumField()
		for i := 0; i < num; i++ {
			field := vh.Type.Field(i)
			val := vh.Value.Field(i)
			if val.Kind() == reflect.Struct && field.Anonymous {
				sr := extractSchema(val.Interface())
				objSch.Required = append(objSch.Required, sr.Value.Required...)
				for k, v := range sr.Value.Properties {
					objSch.Properties[k] = v
				}
				continue
			}

			fTag := field.Tag.Get("json")
			fieldNameArr := strings.Split(fTag, ",")
			if len(fieldNameArr[0]) == 0 || fieldNameArr[0] == "-" {
				fTag = field.Tag.Get("form")
				fieldNameArr = strings.Split(fTag, ",")
				if len(fieldNameArr[0]) == 0 || fieldNameArr[0] == "-" {
					continue
				}
			}
			fn := fieldNameArr[0]

			docStr := field.Tag.Get(RootTag)
			if len(docStr) == 0 {
				continue
			}

			schemeRef := extractSchema(val.Interface())
			if schemeRef == nil {
				continue
			}
			kvs := parseDocTag(docStr)
			for _, schemaFieldKV := range kvs {
				switch schemaFieldKV[0] {
				case DocTagFieldMaxItems:
					if n, err := strconv.ParseInt(schemaFieldKV[1], 10, 64); err != nil {
						panic(fmt.Sprintf("field %s maxItems %s is not a number", schemaFieldKV[0], schemaFieldKV[1]))
					} else {
						un := uint64(n)
						schemeRef.Value.MaxItems = &un
					}
				case DocTagFieldMinItems:
					if n, err := strconv.ParseInt(schemaFieldKV[1], 10, 64); err != nil {
						panic(fmt.Sprintf("field %s minItems %s is not a number", schemaFieldKV[0], schemaFieldKV[1]))
					} else {
						schemeRef.Value.MinItems = uint64(n)
					}
				case DocTagFieldNullable:
					schemeRef.Value.Nullable = true
				case DocTagFieldRequired:
					objSch.Required = append(objSch.Required, fn)
				case DocTagFieldFormat:
					schemeRef.Value.Format = schemaFieldKV[1]
				case DocTagFieldPattern:
					if val, exists := Refs[schemaFieldKV[1]]; exists {
						schemeRef.Value.Pattern = val.(string)
					} else {
						schemeRef.Value.Pattern = schemaFieldKV[1]
					}
				case DocTagFieldDescription:
					if val, exists := Refs[schemaFieldKV[1]]; exists {
						schemeRef.Value.Description = val.(string)
					} else {
						schemeRef.Value.Description = schemaFieldKV[1]
					}
				case DocTagFieldEnum:
					if val, exists := Refs[schemaFieldKV[1]]; exists {
						schemeRef.Value.Enum = val.([]interface{})
					} else {
						sp := strings.Split(schemaFieldKV[1], ";")
						var enums []interface{}
						for _, item := range sp {
							enums = append(enums, item)
						}
						schemeRef.Value.Enum = enums
					}
				case DocTagFieldStringMaxLength:
					lessThanOrEqualTo, err := strconv.Atoi(schemaFieldKV[1])
					if err != nil {
						panic(fmt.Sprintf("prototype=%+v field=%s tag=%s value invalid, err=%v", prototype, field.Name, schemaFieldKV[0], err.Error()))
					}
					max := uint64(lessThanOrEqualTo)
					schemeRef.Value.MaxLength = &max

				case DocTagFieldMaximum:
					lessThanOrEqualTo, err := strconv.ParseFloat(schemaFieldKV[1], 64)
					if err != nil {
						panic(fmt.Sprintf("prototype=%+v field=%s tag=%s value invalid, err=%v", prototype, field.Name, schemaFieldKV[0], err.Error()))
					}
					schemeRef.Value.Max = &lessThanOrEqualTo

				case DocTagFieldStringMinLength:
					greaterThanOrEqual, err := strconv.Atoi(schemaFieldKV[1])
					if err != nil {
						panic(fmt.Sprintf("prototype=%+v field=%s tag=%s value invalid, err=%v", prototype, field.Name, schemaFieldKV[0], err.Error()))
					}
					schemeRef.Value.MinLength = uint64(greaterThanOrEqual)

				case DocTagFieldMinimum:
					greaterThanOrEqual, err := strconv.ParseFloat(schemaFieldKV[1], 64)
					if err != nil {
						panic(fmt.Sprintf("prototype=%+v field=%s tag=%s value invalid, err=%v", prototype, field.Name, schemaFieldKV[0], err.Error()))
					}
					schemeRef.Value.Min = &greaterThanOrEqual

				}
			}
			objSch.Properties[fn] = schemeRef
		}
		return openapi3.NewSchemaRef("", objSch)
	}

	panic(fmt.Sprintf("prototype=%+v not supported", prototype))
}

func parseDocTag(tagContent string) [][]string {
	var ret [][]string
	stateNextField := 0
	stateFieldKey := 1
	stateFieldVal := 2
	state := stateNextField
	var parenthesesStack []struct{}
	var key, val string
	for _, char := range tagContent {
		switch state {
		case stateNextField:
			parenthesesStack = []struct{}{}
			if char != ' ' {
				state = stateFieldKey
				key = string(char)
			}
		case stateFieldKey:
			if char == ' ' {
				state = stateNextField
				ret = append(ret, []string{key, val})
				key = ""
				val = ""
			} else if char == '(' {
				state = stateFieldVal
			} else {
				key += string(char)
			}
		case stateFieldVal:
			if char == '(' {
				if len(val) > 0 && val[len(val)-1] == '\\' {
					val = val[:len(val)-1] + "("
				} else {
					parenthesesStack = append(parenthesesStack, struct{}{})
					val += string(char)
				}
			} else if char == ')' {
				if len(val) > 0 && val[len(val)-1] == '\\' {
					val = val[:len(val)-1] + ")"
				} else {
					if len(parenthesesStack) == 0 {
						state = stateNextField
						ret = append(ret, []string{key, val})
						key = ""
						val = ""
					} else {
						parenthesesStack = parenthesesStack[:len(parenthesesStack)-1]
						val += string(char)
					}
				}
			} else {
				val += string(char)
			}
		}
	}
	if state == stateFieldKey && len(ret) == 0 {
		ret = append(ret, []string{key, val})
	} else if state != stateNextField {
		panic("invalid doc tag, content=" + tagContent)
	}

	return ret
}

type valHolder struct {
	Type  reflect.Type
	Value reflect.Value
}

func extractMap(mapOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(mapOrPtr, reflect.Map); ret != nil {
		return ret
	}
	return nil
}

func extractArray(arrayOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(arrayOrPtr, reflect.Slice); ret != nil {
		return ret
	}
	return nil
}

func extractInteger(integerOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(integerOrPtr, reflect.Int); ret != nil {
		return ret
	}
	if ret := extractOpenAPIType(integerOrPtr, reflect.Int8); ret != nil {
		return ret
	}
	if ret := extractOpenAPIType(integerOrPtr, reflect.Int16); ret != nil {
		return ret
	}
	if ret := extractOpenAPIType(integerOrPtr, reflect.Int32); ret != nil {
		return ret
	}
	if ret := extractOpenAPIType(integerOrPtr, reflect.Int64); ret != nil {
		return ret
	}
	return nil
}

func extractNumber(floatOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(floatOrPtr, reflect.Float32); ret != nil {
		return ret
	}
	if ret := extractOpenAPIType(floatOrPtr, reflect.Float64); ret != nil {
		return ret
	}
	return nil
}

func extractString(stringOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(stringOrPtr, reflect.String); ret != nil {
		return ret
	}
	return nil
}

func extractBoolean(boolOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(boolOrPtr, reflect.Bool); ret != nil {
		return ret
	}
	return nil
}

func extractStruct(structOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(structOrPtr, reflect.Struct); ret != nil {
		return ret
	}
	return nil
}

func extractFileHeader(fileHeaderOrPtr interface{}) *valHolder {
	if ret := extractOpenAPIType(fileHeaderOrPtr, reflect.Struct); ret != nil {
		if ret.Type.AssignableTo(reflect.ValueOf(multipart.FileHeader{}).Type()) {
			return ret
		}
	}
	return nil
}

func extractOpenAPIType(valOrValPtr interface{}, kind reflect.Kind) *valHolder {
	t := reflect.TypeOf(valOrValPtr)
	v := reflect.ValueOf(valOrValPtr)

	if t.Kind() == reflect.Ptr {
		if t.Elem().Kind() != kind {
			return nil
		}
		t = t.Elem()
		v = v.Elem()
	} else if t.Kind() != kind {
		return nil
	}
	return &valHolder{
		Type:  t,
		Value: v,
	}
}
