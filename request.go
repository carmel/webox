package webox

import (
	"bytes"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"
	"webox/util"

	jsoniter "github.com/json-iterator/go"
)

// BodyType ...
type BodyType string

// BodyTypeNone ...
const (
	BodyTypeNone      BodyType = "none"
	BodyTypeJSON      BodyType = "json"
	BodyTypeXML       BodyType = "xml"
	BodyTypeMultipart BodyType = "multipart"
	BodyTypeForm      BodyType = "form"
)

// RequestBody ...
type RequestBody struct {
	BodyType       BodyType
	BodyInstance   any
	RequestBuilder RequestBuilderFunc
}

// RequestBuilderFunc ...
type RequestBuilderFunc func(method, url string, i any) (*http.Request, error)

// TODO
var buildMultipart = buildNothing
var buildForm = buildNothing

var builder = map[BodyType]RequestBuilderFunc{
	BodyTypeXML:       buildXML,
	BodyTypeJSON:      buildJSON,
	BodyTypeForm:      buildForm,
	BodyTypeMultipart: buildMultipart,
	BodyTypeNone:      buildNothing,
}

func buildXML(method, url string, i any) (*http.Request, error) {
	request, e := http.NewRequest(method, url, xmlReader(i))
	if e != nil {
		return nil, e
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	return request, nil
}

// xmlReader ...
func xmlReader(v any) io.Reader {
	var reader io.Reader
	switch v := v.(type) {
	case string:

		reader = strings.NewReader(v)
	case []byte:

		reader = bytes.NewReader(v)
	case util.Map:

		reader = bytes.NewReader(v.ToXML())
	case io.Reader:

		return v
	default:

		if v0, e := xml.Marshal(v); e == nil {

			reader = bytes.NewReader(v0)
		}
	}
	return reader
}

func buildJSON(method, url string, i any) (*http.Request, error) {
	request, e := http.NewRequest(method, url, jsonReader(i))
	if e != nil {
		return nil, e
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	return request, nil
}

func buildNothing(method, url string, i any) (*http.Request, error) {
	request, e := http.NewRequest(method, url, nil)
	if e != nil {
		return nil, e
	}
	return request, nil
}

// jsonReader ...
func jsonReader(v any) io.Reader {
	var reader io.Reader
	switch v := v.(type) {
	case string:

		reader = strings.NewReader(v)
	case []byte:

		reader = bytes.NewReader(v)
	case util.Map:

		reader = bytes.NewReader(v.ToJSON())
	case io.Reader:

		return v
	default:

		if v0, e := jsoniter.Marshal(v); e == nil {
			reader = bytes.NewReader(v0)
		}
	}
	return reader
}

// buildBody ...
func buildBody(v any, tp BodyType) *RequestBody {
	build, b := builder[tp]
	if !b {
		build = buildNothing
	}
	return &RequestBody{
		BodyType:       tp,
		RequestBuilder: build,
		BodyInstance:   v,
	}
}

/*Requester Requester */
type Requester interface {
	Type() BodyType
	BodyReader
}

// Request ...
type Request struct {
	bytes []byte
	err   error
}

// Type ...
func (r *Request) Type() BodyType {
	return BodyTypeNone
}

// xmlResponse ...
type xmlRequest struct {
	Request
	data util.Map
}

// Type ...
func (r *xmlRequest) Type() BodyType {
	return BodyTypeXML
}

// XMLRequest ...
func XMLRequest(bytes []byte) Requester {
	return &xmlRequest{
		Request: Request{
			bytes: bytes,
		},
	}
}

// ToMap ...
func (r *xmlRequest) ToMap() util.Map {
	maps, e := r.Result()
	if e != nil {
		return nil
	}
	return maps
}

// Unmarshal ...
func (r *xmlRequest) Unmarshal(v any) error {
	return xml.Unmarshal(r.bytes, v)
}

// Result ...
func (r *xmlRequest) Result() (util.Map, error) {
	if r.data != nil {
		return r.data, nil
	}
	r.data = make(util.Map)
	e := r.Unmarshal(&r.data)
	return r.data, e
}

// jsonResponse ...
type jsonRequest struct {
	Request
	data util.Map
}

// Type ...
func (r *jsonRequest) Type() BodyType {
	return BodyTypeJSON
}

// JSONRequest ...
func JSONRequest(bytes []byte) Requester {
	return &jsonRequest{
		Request: Request{
			bytes: bytes,
		},
	}
}

// ToMap ...
func (r *jsonRequest) ToMap() util.Map {
	maps, e := r.Result()
	if e != nil {
		return nil
	}
	return maps
}

// Unmarshal ...
func (r *jsonRequest) Unmarshal(v any) error {
	return jsoniter.Unmarshal(r.bytes, v)
}

// Result ...
func (r *jsonRequest) Result() (util.Map, error) {
	r.data = make(util.Map)
	e := r.Unmarshal(&r.data)
	return r.data, e
}

// ToMap ...
func (r *Request) ToMap() util.Map {
	return nil
}

// Unmarshal ...
func (r *Request) Unmarshal(v any) error {
	return r.err
}

// Result ...
func (r *Request) Result() (util.Map, error) {
	return nil, r.err
}

// ErrRequest ...
func ErrRequest(err error) Requester {
	return &Request{
		bytes: nil,
		err:   err,
	}
}

// Bytes ...
func (r *Request) Bytes() []byte {
	return r.bytes
}

// Error ...
func (r *Request) Error() error {
	return r.err
}

// BuildRequester ...
func BuildRequester(req *http.Request) Requester {
	ct := req.Header.Get("Content-Type")
	body, err := readBody(req.Body)
	if err != nil {
		log.Println(body, err)
		return ErrRequest(err)
	}

	if strings.Index(ct, "xml") != -1 ||
		bytes.Index(body, []byte("<xml")) != -1 {
		return XMLRequest(body)
	}
	return JSONRequest(body)
	//return ErrResponder(xerrors.New("error with code " + req.Status))
}

// RebuildRequester ...
func RebuildRequester(req Requester, data util.Map) Requester {
	switch req.Type() {
	case BodyTypeXML:
		return XMLRequest(data.ToXML())
	case BodyTypeJSON:
	default:

	}
	return JSONRequest(data.ToJSON())
}
