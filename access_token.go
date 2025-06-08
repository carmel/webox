package webox

import (
	"crypto/md5"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"
	"webox/api"
	"webox/cache"
	"webox/util"
)

/*AccessToken GetToken */
type AccessToken struct {
	*AccessTokenProperty
	remoteURL string
	tokenKey  string
	tokenURL  string
}

/*AccessTokenSafeSeconds token安全时间 */
const AccessTokenSafeSeconds = 500

// RemoteURL ...
func (obj *AccessToken) RemoteURL() string {
	if obj != nil && obj.remoteURL != "" {
		return obj.remoteURL
	}
	return api.ApiWeixin
}

// TokenURL ...
func (obj *AccessToken) TokenURL() string {
	return util.URL(obj.RemoteURL(), tokenURL(obj))
}
func tokenURL(obj *AccessToken) string {
	if obj != nil && obj.tokenURL != "" {
		return obj.tokenURL
	}
	return api.AccessToken
}

/*NewAccessToken NewAccessToken*/
func NewAccessToken(property *AccessTokenProperty, options ...AccessTokenOption) *AccessToken {
	token := &AccessToken{
		AccessTokenProperty: property,
	}
	token.parse(options...)
	return token
}

func (obj *AccessToken) parse(options ...AccessTokenOption) {
	if options == nil {
		return
	}
	for _, o := range options {
		o(obj)
	}
}

/*Refresh 刷新AccessToken */
func (obj *AccessToken) Refresh() *AccessToken {

	obj.getToken(true)
	return obj
}

/*GetRefreshToken 获取刷新token */
func (obj *AccessToken) GetRefreshToken() *Token {

	return obj.getToken(true)
}

/*GetToken 获取token */
func (obj *AccessToken) GetToken() *Token {
	return obj.getToken(false)
}

// KeyMap ...
func (obj *AccessToken) KeyMap() util.Map {
	return MustKeyMap(obj)
}

func (obj *AccessToken) getToken(refresh bool) *Token {
	key := obj.getCacheKey()

	if !refresh && cache.Has(key) {
		if v, b := cache.Get(key).(string); b {
			token, e := ParseToken(v)
			if e != nil {
				log.Println("parse token error")
				return nil
			}

			return token
		}
	}

	token, e := requestToken(obj.TokenURL(), obj.AccessTokenProperty)
	if e != nil {
		log.Println(e)
		return nil
	}

	if v := token.ExpiresIn; v != 0 {
		obj.SetTokenWithLife(token.ToJSON(), v-AccessTokenSafeSeconds)
	} else {
		obj.SetToken(token.ToJSON())
	}
	return token
}

func requestToken(url string, credentials *AccessTokenProperty) (*Token, error) {
	var t Token
	var e error
	token := Get(url, credentials.ToMap())
	if e := token.Error(); e != nil {
		return nil, e
	}
	e = token.Unmarshal(&t)
	if e != nil {
		return nil, e
	}
	return &t, nil
}

/*SetTokenWithLife set string AccessToken with life time */
func (obj *AccessToken) SetTokenWithLife(token string, tts int64) *AccessToken {
	return obj.setToken(token, tts)
}

/*SetToken set string AccessToken */
func (obj *AccessToken) SetToken(token string) *AccessToken {
	return obj.setToken(token, 7200-AccessTokenSafeSeconds)
}

func (obj *AccessToken) setToken(token string, tts int64) *AccessToken {
	cache.Set(obj.getCacheKey(), token, time.Duration(tts*int64(time.Second)))
	return obj
}

func (obj *AccessToken) getCredentials() string {
	cred := strings.Join([]string{obj.GrantType, obj.AppID, obj.AppSecret}, ".")
	c := md5.Sum([]byte(cred))
	return fmt.Sprintf("%x", c[:])
}

func (obj *AccessToken) getCacheKey() string {
	return "webox.access_token." + obj.getCredentials()
}

const accessTokenNil = "nil point AccessToken"
const tokenNil = "nil point token"

/*MustKeyMap get AccessToken's key,value with map when nil or error return nil map */
func MustKeyMap(at *AccessToken) util.Map {
	if m, e := KeyMap(at); e == nil {
		return m
	}
	return util.Map{}
}

/*KeyMap get AccessToken's key,value with map */
func KeyMap(at *AccessToken) (util.Map, error) {
	if at == nil {
		return nil, errors.New(accessTokenNil)
	}
	if token := at.GetToken(); token != nil {
		return token.KeyMap(), nil
	}
	return nil, errors.New(tokenNil)
}

func parseAccessToken(token any) string {
	switch v := token.(type) {
	case Token:
		return v.AccessToken
	case *Token:
		return v.AccessToken
	case string:
		return v
	}
	return ""
}
