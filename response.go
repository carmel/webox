package webox

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"webox/util"

	jsoniter "github.com/json-iterator/go"
	"golang.org/x/text/transform"
)

/*Responder Responder */
type Responder interface {
	io.ReadCloser
	BodyReader
	Type() BodyType
}

// Response ...
type Response struct {
	bytes  []byte
	reader io.ReadCloser
	err    error
}

// Type ...
func (r *Response) Type() BodyType {
	return BodyTypeNone
}

// Read ...
func (r *Response) Read(p []byte) (n int, err error) {
	return r.reader.Read(p)
}

// Close ...
func (r *Response) Close() error {
	return r.reader.Close()
}

// xmlResponse ...
type xmlResponse struct {
	Response
	data util.Map
}

// Type ...
func (r *xmlResponse) Type() BodyType {
	return BodyTypeXML
}

// XMLResponseReader ...
func XMLResponseReader(reader io.ReadCloser) Responder {
	return &xmlResponse{
		Response: Response{
			reader: reader,
		},
	}
}

// XMLResponse ...
func XMLResponse(bytes []byte) Responder {
	return &xmlResponse{
		Response: Response{
			bytes: bytes,
		},
	}
}

// ToMap ...
func (r *xmlResponse) ToMap() util.Map {
	maps, e := r.Result()
	if e != nil {
		return nil
	}
	return maps
}

// Unmarshal ...
func (r *xmlResponse) Unmarshal(v any) error {
	return xml.Unmarshal(r.bytes, v)
}

// Result ...
func (r *xmlResponse) Result() (util.Map, error) {
	if r.data != nil {
		return r.data, nil
	}
	r.data = make(util.Map)
	e := r.Unmarshal(&r.data)
	return r.data, e
}

// Error ...
func (r *xmlResponse) Error() error {
	var e ErrRes
	if r.err != nil {
		return r.err
	}
	_ = jsoniter.Unmarshal(r.bytes, &e)
	if e.ErrCode != 0 {
		return fmt.Errorf("code:%d,msg:%s", e.ErrCode, e.ErrMsg)
	}
	return nil
}

// jsonResponse ...
type jsonResponse struct {
	Response
	data util.Map
}

// Type ...
func (r *jsonResponse) Type() BodyType {
	return BodyTypeJSON
}

// JSONResponseReader ...
func JSONResponseReader(reader io.ReadCloser) Responder {
	return &jsonResponse{
		Response: Response{
			reader: reader,
		},
	}
}

// JSONResponse ...
func JSONResponse(bytes []byte) Responder {
	return &jsonResponse{
		Response: Response{
			bytes: bytes,
		},
	}
}

// ToMap ...
func (r *jsonResponse) ToMap() util.Map {
	maps, e := r.Result()
	if e != nil {
		return nil
	}
	return maps
}

// Unmarshal ...
func (r *jsonResponse) Unmarshal(v any) error {
	return jsoniter.Unmarshal(r.bytes, v)
}

// Result ...
func (r *jsonResponse) Result() (util.Map, error) {
	r.data = make(util.Map)
	e := r.Unmarshal(&r.data)
	return r.data, e
}

// Error ...
func (r *jsonResponse) Error() error {
	var e ErrRes
	if r.err != nil {
		return r.err
	}
	_ = jsoniter.Unmarshal(r.bytes, &e)
	if e.ErrCode != 0 {
		return fmt.Errorf("code:%d,msg:%s", e.ErrCode, e.ErrMsg)
	}
	return nil
}

// ToMap ...
func (r *Response) ToMap() util.Map {
	return nil
}

// Unmarshal ...
func (r *Response) Unmarshal(v any) error {
	return r.err
}

// Result ...
func (r *Response) Result() (util.Map, error) {
	return nil, r.err
}

// ErrResponder ...
func ErrResponder(err error) Responder {
	return &Response{
		bytes: nil,
		err:   err,
	}
}

// Bytes ...
func (r *Response) Bytes() []byte {
	return r.bytes
}

// ErrRes ...
type ErrRes struct {
	ErrCode int64
	ErrMsg  string
}

// Error ...
func (r *Response) Error() error {
	return r.err
}

// ResponseWriter ...
func ResponseWriter(w http.ResponseWriter, responder Responder) error {
	w.WriteHeader(http.StatusOK)
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		switch responder.Type() {
		case BodyTypeXML:
			header["Content-Type"] = []string{"application/xml; charset=utf-8"}
		case BodyTypeJSON:
			header["Content-Type"] = []string{"application/json; charset=utf-8"}
		}
	}
	b := responder.Bytes()
	if b == nil {
		return nil
	}

	_, err := w.Write(b)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// BuildResponder ...
func BuildResponder(resp *http.Response) Responder {
	ct := resp.Header.Get("Content-Type")
	body, err := readBody(resp.Body)
	if err != nil {
		log.Println(body, err)
		return ErrResponder(err)
	}

	if resp.StatusCode == 200 {
		if strings.Index(ct, "xml") != -1 ||
			bytes.Index(body, []byte("<xml")) != -1 {
			return XMLResponse(body)
		}
		return JSONResponse(body)
	}
	log.Println("error with " + resp.Status)
	return ErrResponder(errors.New("error with code " + resp.Status))
}

// SaveTo ...
func SaveTo(response Responder, path string) error {
	var err error
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_SYNC, os.ModePerm)
	if err != nil {

		return err
	}
	defer func() {
		err = file.Close()
	}()

	if _, err = io.Copy(file, response); err != nil {
		return err
	}

	//_, err = file.Write(response.Bytes())
	//if err != nil {
	//	return err
	//}
	return nil
}

// SaveEncodingTo ...
func SaveEncodingTo(response Responder, path string, t transform.Transformer) (err error) {
	var file *os.File
	file, err = os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_SYNC, os.ModePerm)
	if err != nil {

		return err
	}
	defer func() {
		err = file.Close()
	}()
	writer := transform.NewWriter(file, t)
	if _, err = io.Copy(writer, response); err != nil {
		return err
	}

	//_, err = writer.Write(response.Bytes())
	//if err != nil {
	//	return err
	//}
	defer func() {
		err = writer.Close()
	}()
	return nil
}
