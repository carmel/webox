package webox

import (
	"encoding/xml"
	"errors"
	"log"
	"net/http"
	"net/url"
	"webox/cipher"
	"webox/util"

	jsoniter "github.com/json-iterator/go"
)

// NotifyResult ...
type NotifyResult struct {
	ReturnCode string `json:"return_code" xml:"return_code"`
	ReturnMsg  string `json:"return_msg,omitempty" xml:"return_msg,omitempty"`
	AppID      string `json:"appid,omitempty" xml:"appid,omitempty"`
	MchID      string `json:"mch_id,omitempty" xml:"mch_id,omitempty"`
	NonceStr   string `json:"nonce_str,omitempty" xml:"nonce_str,omitempty"`
	PrepayID   string `json:"prepay_id,omitempty" xml:"prepay_id,omitempty"`
	ResultCode string `json:"result_code,omitempty" xml:"result_code,omitempty"`
	ErrCodeDes string `json:"err_code_des,omitempty" xml:"err_code_des,omitempty"`
	Sign       string `json:"sign,omitempty" xml:"sign,omitempty"`
}

// Notifier ...
type Notifier interface {
	ServeHTTP(w http.ResponseWriter, req *http.Request)
}

// ServeHTTPFunc ...
type ServeHTTPFunc func(w http.ResponseWriter, req *http.Request)

// RequestHook ...
type RequestHook func(req Requester) (util.Map, error)

// TokenHook ...
type TokenHook func(w http.ResponseWriter, req *http.Request, token *Token, state string) []byte

// UserHook ...
type UserHook func(w http.ResponseWriter, req *http.Request, user *WechatUser) []byte

// StateHook ...
type StateHook func(w http.ResponseWriter, req *http.Request) string

/*authorizeNotify 监听 */
type authorizeNotify struct {
	*OfficialAccount
	TokenHook
	UserHook
	StateHook
}

// ServeHTTP ...
func (n *authorizeNotify) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	query := req.URL.Query()
	if code := query.Get("code"); code != "" {
		token := n.hookAuthorizeToken(w, req, code, query.Get("state"))
		if token != nil {
			info := n.hookUserInfo(w, req, token)
			if info != nil {

			}
		}
		return
	}

	u := n.hookState(w, req)

	http.Redirect(w, req, u, http.StatusFound)
}

func (n *authorizeNotify) hookState(w http.ResponseWriter, req *http.Request) string {
	if n.StateHook != nil {
		s := n.StateHook(w, req)
		return n.AuthCodeURL(s)
	}
	return n.AuthCodeURL("")
}

func (n *authorizeNotify) hookUserInfo(w http.ResponseWriter, req *http.Request, token *Token) *WechatUser {

	info, e := n.GetUserInfo(token)
	if e != nil {
		log.Println("hookUserInfo err:", e.Error())
		return nil
	}
	if n.UserHook != nil {
		bytes := n.UserHook(w, req, info)
		n.responseWriter(w, bytes)
	}
	return info
}

// NotifyResult ...
func (n *authorizeNotify) responseWriter(w http.ResponseWriter, bytes []byte) {
	e := ResponseWriter(w, JSONResponse(bytes))
	if e != nil {
		log.Println(e)
	}
	return
}

func (n *authorizeNotify) hookAuthorizeToken(w http.ResponseWriter, req *http.Request, code string, state string) *Token {

	token, e := n.Oauth2AuthorizeToken(code)
	if e != nil {
		return nil
	}
	if n.TokenHook != nil {
		bytes := n.TokenHook(w, req, token, state)
		n.responseWriter(w, bytes)
	}
	return token
}

/*messageNotify 监听 */
type messageNotify struct {
	*OfficialAccount
	RequestHook
	cipher cipher.Cipher
	//bizMsg *cipher.BizMsg
}

// DecodeReqInfo ...
func (n *messageNotify) decodeInfo(query url.Values, requester Requester) (util.Map, error) {
	var bodies []byte
	var e error
	encryptType := query.Get("encrypt_type")
	timeStamp := query.Get("timestamp")
	nonce := query.Get("nonce")
	msgSignature := query.Get("msg_signature")
	if encryptType != "aes" {
		p := util.Map{}
		e = xml.Unmarshal(bodies, &p)
		if e != nil {
			log.Println(e)
			return nil, e
		}

		bodies, e = n.cipher.Decrypt(&cipher.BizMsgData{
			RSAEncrypt:   p.GetString("RSAEncrypt"),
			TimeStamp:    timeStamp,
			Nonce:        nonce,
			MsgSignature: msgSignature,
		})

		//错误返回,并记录log
		if e != nil {
			log.Println(e)
			return nil, e
		}
	}
	p := util.Map{}
	e = xml.Unmarshal(bodies, &p)
	if e != nil {
		log.Println(e)
		return nil, e
	}
	return p, e
}

// DecodeReqInfo ...
func (n *messageNotify) encodeInfo(p util.Map, ts, nonce string) ([]byte, error) {
	var e error
	bodies, e := n.cipher.Encrypt(&cipher.BizMsgData{
		Text:      string(p.ToXML()),
		TimeStamp: ts,
		Nonce:     nonce,
	})
	//错误返回,并记录log
	if e != nil {
		log.Println(e)
		return nil, e
	}
	return bodies, nil
}

// ServeHTTP ...
func (n *messageNotify) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var e error

	if n.RequestHook == nil {
		log.Println(errors.New("null notify callback "))
		return
	}
	requester := BuildRequester(req)
	if e = requester.Error(); e != nil {
		log.Println(e)
		return
	}

	query, e := url.ParseQuery(req.URL.RawQuery)
	if e != nil {
		log.Println(e)
		return
	}
	maps, e := n.decodeInfo(query, requester)
	if e != nil {
		log.Println(e)
		return
	}

	r, e := n.RequestHook(RebuildRequester(requester, maps))
	if e != nil {
		log.Println(e)
		return
	}

	_, e = w.Write(r.ToXML())

	if e != nil {
		log.Println(e)
		return
	}
}

/*Notifier 监听 */
type paymentPaidNotify struct {
	*Payment
	RequestHook
}

// ServerHttp ...
func (n *paymentPaidNotify) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var e error
	requester := BuildRequester(req)
	resp := NotifyTypeResponder(requester.Type(), NotifySuccess())
	defer func() {
		e = resp.Write(w)
		log.Println(e)
	}()

	if e = requester.Error(); e != nil {
		log.Println(e.Error())
		resp.SetNotifyResult(NotifyFail(e.Error()))
		return
	}
	reqData := requester.ToMap()
	if util.ValidateSign(reqData, n.GetKey()) {
		if n.RequestHook == nil {
			log.Println(errors.New("null notify callback "))
			return
		}
		_, e = n.RequestHook(requester)
		if e != nil {
			log.Println(e.Error())
			resp.SetNotifyResult(NotifyFail(e.Error()))
		}
	}

}

/*Notifier 监听 */
type paymentRefundedNotify struct {
	cipher cipher.Cipher
	RequestHook
}

// ServeHTTP ...
func (obj *paymentRefundedNotify) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var e error
	if obj.RequestHook == nil {
		log.Println(errors.New("null notify callback"))
		return
	}

	requester := BuildRequester(req)
	resp := NotifyTypeResponder(requester.Type(), NotifySuccess())
	defer func() {
		e = resp.Write(w)
		log.Println(e)
	}()

	if e = requester.Error(); e != nil {
		log.Println(e.Error())
		resp.SetNotifyResult(NotifyFail(e.Error()))
		return
	}
	reqData := requester.ToMap()
	reqInfo := reqData.GetString("req_info")
	reqData.Set("reqInfo", obj.DecodeReqInfo(reqInfo))

	_, e = obj.RequestHook(requester)
	if e != nil {
		log.Println(e.Error())
		resp.SetNotifyResult(NotifyFail(e.Error()))
	}
}

// DecodeReqInfo ...
func (obj *paymentRefundedNotify) DecodeReqInfo(info string) util.Map {
	maps := util.Map{}
	dec, _ := obj.cipher.Decrypt(info)
	e := xml.Unmarshal(dec, &maps)
	if e != nil {
		log.Println(e)
	}
	return maps
}

/*Notifier 监听 */
type paymentScannedNotify struct {
	*Payment
	RequestHook
}

// ServeHTTP ...
func (obj *paymentScannedNotify) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	var e error
	var p util.Map

	if obj.RequestHook == nil {
		log.Println(errors.New("null notify callback"))
		return
	}
	requester := BuildRequester(req)
	resp := NotifyTypeResponder(requester.Type(), NotifySuccess())
	defer func() {
		e = resp.Write(w)
		log.Println(e)
	}()

	if e = requester.Error(); e != nil {
		log.Println(e.Error())
		resp.SetNotifyResult(NotifyFail(e.Error()))
		return
	}
	reqData := requester.ToMap()
	if util.ValidateSign(reqData, obj.GetKey()) {

		p, e = obj.RequestHook(requester)
		if e != nil {
			log.Println(e.Error())
			resp.SetNotifyResult(NotifyFailDes(resp.NotifyResult(), e.Error()))
		}
		if !p.Has("prepay_id") {
			log.Println("null prepay_id")
			resp.SetNotifyResult(NotifyFailDes(resp.NotifyResult(), "null prepay_id"))
		} else {
			//公众账号ID	appid	String(32)	是	wx8888888888888888	微信分配的公众账号ID
			//商户号	mch_id	String(32)	是	1900000109	微信支付分配的商户号
			//随机字符串	nonce_str	String(32)	是	5K8264ILTKCH16CQ2502SI8ZNMTM67VS	微信返回的随机字符串
			//预支付ID	prepay_id	String(64)	是	wx201410272009395522657a690389285100	调用统一下单接口生成的预支付ID
			//业务结果	result_code	String(16)	是	SUCCESS	SUCCESS/FAIL
			//错误描述	err_code_des	String(128)	否		当result_code为FAIL时，商户展示给用户的错误提
			//签名	sign	String(32)	是	C380BEC2BFD727A4B6845133519F3AD6	返回数据签名，签名生成算法
			res := resp.NotifyResult()
			res.AppID = obj.AppID
			res.MchID = obj.MchID
			res.NonceStr = util.GenerateNonceStr()
			res.PrepayID = p.GetString("prepay_id")
			res.Sign = util.GenSign(reqData, obj.GetKey())
		}

	}

}

// NotifyResponder ...
type NotifyResponder interface {
	SetNotifyResult(result *NotifyResult)
	NotifyResult() *NotifyResult
	Write(w http.ResponseWriter) error
}

type xmlNotify struct {
	notifyResult *NotifyResult
}

// NotifyResult ...
func (obj *xmlNotify) NotifyResult() *NotifyResult {
	return obj.notifyResult
}

// SetNotifyResult ...
func (obj *xmlNotify) SetNotifyResult(notifyResult *NotifyResult) {
	obj.notifyResult = notifyResult
}

// Write ...
func (obj *xmlNotify) Write(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/xml; charset=utf-8"}
	}
	if obj.notifyResult == nil {
		return errors.New("null notify result")
	}
	_, err := w.Write(obj.notifyResult.ToXML())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

type jsonNotify struct {
	notifyResult *NotifyResult
}

// NotifyResult ...
func (obj *jsonNotify) NotifyResult() *NotifyResult {
	return obj.notifyResult
}

// SetNotifyResult ...
func (obj *jsonNotify) SetNotifyResult(notifyResult *NotifyResult) {
	obj.notifyResult = notifyResult
}

// Write ...
func (obj *jsonNotify) Write(w http.ResponseWriter) error {
	w.WriteHeader(http.StatusOK)
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{"application/json; charset=utf-8"}
	}
	if obj.notifyResult == nil {
		return errors.New("null notify result")
	}
	_, err := w.Write(obj.notifyResult.ToJSON())
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}

// NotifyTypeResponder ...
func NotifyTypeResponder(bodyType BodyType, notifyResult *NotifyResult) NotifyResponder {
	switch bodyType {
	case BodyTypeJSON:
		return &jsonNotify{
			notifyResult: notifyResult,
		}
	case BodyTypeXML:
		return &xmlNotify{
			notifyResult: notifyResult,
		}
	}
	return nil
}

// ToJSON ...
func (obj *NotifyResult) ToJSON() []byte {
	bytes, e := jsoniter.Marshal(obj)
	if e != nil {
		log.Println(e)
		return nil
	}
	return bytes
}

// ToXML ...
func (obj *NotifyResult) ToXML() []byte {
	bytes, e := xml.Marshal(obj)
	if e != nil {
		log.Println(e)
		return nil
	}
	return bytes
}

// NotifySuccess ...
func NotifySuccess() *NotifyResult {
	return &NotifyResult{
		ReturnCode: "SUCCESS",
		ReturnMsg:  "OK",
	}
}

// NotifyFail ...
func NotifyFail(msg string) *NotifyResult {
	return &NotifyResult{
		ReturnCode: "FAIL",
		ReturnMsg:  msg,
	}
}

// NotifyFailDes ...
func NotifyFailDes(r *NotifyResult, msg string) *NotifyResult {
	r.ResultCode = "FAIL"
	r.ErrCodeDes = msg
	return r
}
