package webox

import (
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"webox/api"
	"webox/util"
)

// Token represents the credentials used to authorize
// the requests to access protected resources on the OAuth 2.0
// provider's backend.
//
// This type is a mirror of oauth2.Token and exists to break
// an otherwise-circular dependency. Other internal packages
// should convert this Token into an oauth2.Token before use.
type Token struct {
	// AccessToken is the AccessToken that authorizes and authenticates
	// the requests.
	AccessToken string `json:"access_token"`

	// RefreshToken is a AccessToken that's used by the application
	// (as opposed to the user) to refresh the access AccessToken
	// if it expires.
	RefreshToken string `json:"refresh_token"`

	// Expiry is the optional expiration time of the access AccessToken.
	//
	// If zero, TokenSource implementations will reuse the same
	// AccessToken forever and RefreshToken or equivalent
	// mechanisms for that TokenSource will not be used.
	ExpiresIn int64 `json:"expires_in"`

	// wechat openid
	OpenID string `json:"openid"`

	// wechat scope
	Scope string `json:"scope"`
	// Raw optionally contains extra metadata from the server
	// when updating a AccessToken.
	Raw any
}

/*KeyMap get AccessToken's key,value with map */
func (t *Token) KeyMap() util.Map {
	if t.AccessToken == "" {
		return nil
	}
	return util.Map{
		api.AccessTokenKey: t.AccessToken,
	}
}

/*SetExpiresIn set expires time */
func (t *Token) SetExpiresIn(ti time.Time) *Token {
	t.ExpiresIn = ti.Unix()
	return t
}

/*GetExpiresIn get expires time */
func (t *Token) GetExpiresIn() time.Time {
	return time.Unix(t.ExpiresIn, 0)
}

/*GetScopes get AccessToken scopes for get AccessToken*/
func (t *Token) GetScopes() []string {
	return strings.Split(t.Scope, ",")
}

/*SetScopes set AccessToken scopes for get AccessToken*/
func (t *Token) SetScopes(s []string) *Token {
	strings.Join(s, ",")
	return t
}

/*ToJSON transfer AccessToken to json*/
func (t *Token) ToJSON() string {
	s, e := jsoniter.MarshalToString(t)
	if e != nil {
		return ""
	}
	return s
}

/*ParseToken parse AccessToken from string*/
func ParseToken(src string) (*Token, error) {
	var t Token

	e := jsoniter.UnmarshalFromString(src, &t)
	if e != nil {
		return nil, e
	}
	return &t, nil
}
