package easyreq

import (
	"context"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"sync"

	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

// Bind
const (
	inTagPath   = "path"
	inTagQuery  = "query"
	inTagBody   = "body"
	inTagForm   = "form"
	inTagHeader = "header"

	tagNameIn = "pos"
	tagSep    = ","
)

// Bind bind params from Path, Query, Body, Form. Donot support binary stream(files, images etc.)
// Support Tag `pos`, specified that where we can get this value, only support one
// - path: from url path, don't support nested struct
// - query: from url query, don't support nested struct
// - body: from request's body, default use json, support nested struct
// - form: from request form
// - required: this value is not null
// pathQueryier get variables from path, GET /api/v1/users/:id , get id
/*
type Example struct {
	ID   string `json:"id"   pos:"path:id"`             // path value default is required
	Name string `json:"name" pos:"query:name,required"` // query specified that get
}
*/
func Bind(req *http.Request, params interface{}, pathQueryier ...interface{}) (err error) {
	paramsVal := reflect.ValueOf(params)
	if paramsVal.Kind() != reflect.Ptr {
		err = errors.New("can't bind to nonpointer value")
		return
	}

	for paramsVal.Kind() == reflect.Ptr {
		if paramsVal.IsNil() {
			paramsVal.Set(reflect.New(paramsVal.Type().Elem()))
		}

		paramsVal = paramsVal.Elem()
	}

	if paramsVal.Kind() != reflect.Struct {
		err = errors.New("can't bind to nonstruct value")
		return
	}

	var (
		typ         = paramsVal.Type()
		wg          = sync.WaitGroup{}
		ctx, cancel = context.WithCancel(context.Background())
		easy        = &easyReq{
			ctx:          ctx,
			req:          req,
			once:         &sync.Once{},
			pathQueryier: pathQueryier,
		}
	)

	defer cancel()

	for i := 0; i < paramsVal.NumField(); i++ {
		field := paramsVal.Field(i)
		fieldType := typ.Field(i)
		wg.Add(1)
		go func() {
			err = easy.bindFieldWithCtx(field, fieldType)
			if err != nil {
				cancel()
			}
			wg.Done()
		}()
	}

	wg.Wait()

	if easy.hasJSONBody {
		err = json.NewDecoder(req.Body).Decode(params)
	}

	return
}

type easyReq struct {
	ctx          context.Context
	once         *sync.Once
	pathQueryier []interface{}
	req          *http.Request
	hasJSONBody  bool
}

func (e *easyReq) bindFieldWithCtx(field reflect.Value, fieldType reflect.StructField) (err error) {
	var (
		errCh  = make(chan error, 1)
		doneCh = make(chan struct{}, 1)
	)
	go func() {
		e.bindField(field, fieldType, errCh)
		doneCh <- struct{}{}
	}()

	select {
	case <-e.ctx.Done():
		return
	case err = <-errCh:
		return
	case <-doneCh:
	}

	return
}

func (e *easyReq) bindField(field reflect.Value, fieldType reflect.StructField, errCh chan error) {
	if fieldType.Anonymous {
		r := reflect.New(field.Type())
		err := Bind(e.req, r.Interface(), e.pathQueryier...)
		if err != nil {
			errCh <- err
			return
		}
		field.Set(r.Elem())
	}

	if len(fieldType.Tag.Get("json")) > 0 {
		e.hasJSONBody = true
	}

	var (
		loc, name = getInTagLocAndName(fieldType)
		values    = make([]string, 0, 1)
	)

	switch loc {
	case inTagPath:
		pathVal := getValueFromPath(name, e.pathQueryier)
		values = append(values, pathVal)
	case inTagQuery:
		values = e.req.URL.Query()[name]
	case inTagHeader:
		values = e.req.Header.Values(name)
	case inTagForm:
		e.once.Do(func() {
			e.req.ParseForm()
		})

		values = e.req.PostForm[name]
	}

	var reflectVal reflect.Value
	switch len(values) {
	case 0:
		return
	case 1:
		reflectVal = BindValue(values[0], field.Type())
	default:
		reflectVal = sliceBinder(values, field.Type())
	}

	if reflectVal.Type().ConvertibleTo(field.Type()) {
		if reflectVal.Type() == field.Type() {
			if field.Type().Kind() == reflect.Array || field.Type().Kind() == reflect.Slice {
				field.Set(reflect.AppendSlice(field, reflectVal))
			} else {
				field.Set(reflectVal)
			}
		} else {
			field.Set(reflectVal.Convert(field.Type()))
		}
	}

}

func getInTagLocAndName(fieldType reflect.StructField) (loc, name string) {
	inTag := fieldType.Tag.Get(tagNameIn)
	if len(inTag) == 0 {
		loc = inTagBody
		name = fieldType.Name
		return
	}

	splits := strings.Split(inTag, tagSep)
	locs := strings.Split(splits[0], ":")
	if len(locs) != 2 {
		return
	}

	loc = locs[0]
	name = locs[1]

	return
}

type giner interface {
	Param(string) string
}

type httprouter interface {
	ByName(string) string
}

func getValueFromPath(name string, pathQueryier ...interface{}) string {
	if len(pathQueryier) == 0 {
		return ""
	}

	if g, ok := pathQueryier[0].(giner); ok {
		return g.Param(name)
	}

	if h, ok := pathQueryier[0].(httprouter); ok {
		return h.ByName(name)
	}

	return ""
}
