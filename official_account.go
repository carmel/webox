package webox

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/url"
	"strings"
	"time"
	"webox/api"
	"webox/model"
	"webox/util"
)

// OfficialAccount ...
type OfficialAccount struct {
	*OfficialAccountProperty
	BodyType    BodyType
	oauth       OAuthProperty
	client      *Client
	jssdk       *JSSDK
	AccessToken *AccessToken
	remoteURL   string
	localHost   string
}

// NewOfficialAccount ...
func NewOfficialAccount(config *OfficialAccountProperty, options ...OfficialAccountOption) *OfficialAccount {
	officialAccount := &OfficialAccount{
		OfficialAccountProperty: config,
		BodyType:                BodyTypeJSON,
	}
	officialAccount.parse(options...)
	officialAccount.client = officialAccount.Client()
	return officialAccount
}

func (obj *OfficialAccount) parse(options ...OfficialAccountOption) {
	if options == nil {
		return
	}
	for _, o := range options {
		o(obj)
	}
}

// Client ...
func (obj *OfficialAccount) Client() *Client {
	if obj.client == nil {
		obj.client = NewClient(ClientBodyType(obj.BodyType), ClientAccessToken(obj.AccessToken))
	}
	return obj.client
}

// HandleAuthorizeNotify ...
func (obj *OfficialAccount) HandleAuthorizeNotify(hooks ...any) ServeHTTPFunc {
	return obj.HandleAuthorize(hooks...).ServeHTTP
}

// HandleAuthorize ...
func (obj *OfficialAccount) HandleAuthorize(hooks ...any) Notifier {
	notify := &authorizeNotify{
		OfficialAccount: obj,
	}
	for _, hook := range hooks {
		switch h := hook.(type) {
		case TokenHook:
			notify.TokenHook = h
		case UserHook:
			notify.UserHook = h
		case StateHook:
			notify.StateHook = h
		}
	}
	return notify
}

// GetUserInfo ...
func (obj *OfficialAccount) GetUserInfo(token *Token) (user *WechatUser, e error) {
	p := util.Map{
		"access_token": token.AccessToken,
		"openid":       token.OpenID,
		"lang":         "zh_CN",
	}
	responder := Get(api.SnsUserinfo, p)
	e = responder.Error()
	if e != nil {
		return nil, e
	}

	user = new(WechatUser)
	e = responder.Unmarshal(user)
	if e != nil {
		return nil, e
	}
	return user, nil
}

// Oauth2AuthorizeToken ...
func (obj *OfficialAccount) Oauth2AuthorizeToken(code string) (token *Token, e error) {

	p := util.Map{
		"appid":      obj.AppID,
		"secret":     obj.AppSecret,
		"code":       code,
		"grant_type": "authorization_code",
	}

	uri := obj.RedirectURI()
	if uri != "" {
		p.Set("redirect_uri", uri)
	}

	responder := PostJSON(api.Oauth2AccessToken, p, nil)
	e = responder.Error()
	if e != nil {
		return nil, e
	}

	token = &Token{}
	e = responder.Unmarshal(token)
	if e != nil {
		return nil, e
	}
	return token, nil
}

/*AuthCodeURL 生成授权地址URL*/
func (obj *OfficialAccount) AuthCodeURL(state string) string {

	var buf bytes.Buffer
	buf.WriteString(api.Oauth2Authorize)
	p := util.Map{
		"response_type": "code",
		"appid":         obj.AppID,
	}

	uri := obj.RedirectURI()
	if uri != "" {
		p.Set("redirect_uri", uri)
	}

	if obj.oauth.Scopes != nil {
		p.Set("scope", obj.oauth.Scopes)
	}
	if state != "" {
		// TODO(light): Docs say never to omit state; don't allow empty.
		p.Set("state", state)
	}
	buf.WriteByte('?')
	buf.WriteString(p.URLEncode())
	return buf.String()
}

// RedirectURI ...
func (obj *OfficialAccount) RedirectURI() string {

	if obj.oauth.RedirectURI != "" {
		if strings.Index(obj.oauth.RedirectURI, "http") == 0 {
			return obj.oauth.RedirectURI
		}
		return util.URL(obj.localHost, obj.oauth.RedirectURI)
	}
	return ""
}

// RemoteURL ...
func (obj *OfficialAccount) RemoteURL() string {
	return obj.remoteURL
}

/*
ClearQuota 公众号的所有api调用（包括第三方帮其调用）次数进行清零
HTTP请求方式:POST
HTTP调用: https://api.weixin.qq.com/cgi-bin/clear_quota?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) ClearQuota() Responder {

	u := util.URL(obj.RemoteURL(), api.ClearQuota)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"appid": obj.AppID})
}

/*
GetCallbackIP 请求微信的服务器IP列表
HTTP请求方式: GET
HTTP调用:https://api.weixin.qq.com/cgi-bin/getcallbackip?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetCallbackIP() Responder {

	u := util.URL(obj.RemoteURL(), api.GetCallbackIP)
	return obj.Client().Get(context.Background(), u, nil)
}

// SendMessage 根据OpenID列表群发【订阅号不可用，服务号认证后可用】
// 接口调用请求说明
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/message/mass/send?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) SendMessage(msg util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.SendMessageMass)
	return obj.Client().Post(context.Background(), u, nil, msg)
}

// MessageSendAll 根据标签进行群发【订阅号与服务号认证后均可用】
// 接口调用请求说明
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/message/mass/sendall?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) MessageSendAll(msg util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.MessageMassSendall)
	return obj.Client().Post(context.Background(), u, nil, msg)
}

// MessagePreview 预览接口【订阅号与服务号认证后均可用】
// 开发者可通过该接口发送消息给指定用户，在手机端查看消息的样式和排版。为了满足第三方平台开发者的需求，在保留对openID预览能力的同时，增加了对指定微信号发送预览的能力，但该能力每日调用次数有限制（100次），请勿滥用。
// 接口调用请求说明
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/message/mass/preview?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) MessagePreview(msg util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.MessageMassPreview)
	return obj.Client().Post(context.Background(), u, nil, msg)

}

// DeleteMessage 删除群发【订阅号与服务号认证后均可用】
// 群发之后，随时可以通过该接口删除群发。
// 接口调用请求说明
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/message/mass/delete?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) DeleteMessage(msgID string) Responder {

	u := util.URL(obj.RemoteURL(), api.DeleteMessageMass)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"msg_id": msgID})

}

// MessageStatus 查询群发消息发送状态【订阅号与服务号认证后均可用】
// 接口调用请求说明
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/message/mass/get?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) MessageStatus(msgID string) Responder {

	u := util.URL(obj.RemoteURL(), api.GetMessageMass)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"msg_id": msgID})

}

// MessageSendText ...
func (obj *OfficialAccount) MessageSendText() {

	//TODO:
}

// CreateCardLandingPage 创建货架接口
//
//	HTTP请求方式: POST
//	URL:https://api.weixin.qq.com/card/landingpage/create?access_token=$TOKEN
//	func (c *OfficialAccount) CreateLandingPage(page *CardLandingPage) Responder {
func (obj *OfficialAccount) CreateCardLandingPage(p util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.CreateCardLandingPage)
	return obj.Client().Post(context.Background(), u, nil, p)

}

// CardDeposit 导入code接口
//
//	HTTP请求方式: POST
//	URL:http://api.weixin.qq.com/card/code/deposit?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) CardDeposit(cardID string, code []string) Responder {

	u := util.URL(obj.RemoteURL(), api.DepositCardCode)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"card_id": cardID,
		"code":    code,
	})

}

// CountDepositCardCode 查询导入code数目
//
//	HTTP请求方式: POST
//	URL:http://api.weixin.qq.com/card/code/getdepositcount?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) CountDepositCardCode(cardID string) Responder {

	u := util.URL(obj.RemoteURL(), api.CountDepositCardCode)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"card_id": cardID,
	})
}

// CheckCardCode 核查code接口
//
//	HTTP请求方式: POST
//	HTTP调用:http://api.weixin.qq.com/card/code/checkcode?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) CheckCardCode(cardID string, code []string) Responder {

	u := util.URL(obj.RemoteURL(), api.CheckCardCode)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"card_id": cardID,
		"code":    code,
	})
}

// GetCardCode 查询Code接口
//
//	HTTP请求方式: POST
//	HTTP调用:https://api.weixin.qq.com/card/code/get?access_token=TOKEN
//	参数说明:
//	参数名	必填	类型	示例值	描述
//	code	是	string(20)	110201201245	单张卡券的唯一标准。
//	card_id	否	string(32)	pFS7Fjg8kV1I dDz01r4SQwMkuCKc	卡券ID代表一类卡券。自定义code卡券必填。
//	check_consume	否	bool	true	是否校验code核销状态，填入true和false时的code异常状态返回数据不同。
func (obj *OfficialAccount) GetCardCode(p util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.GetCardCode)
	return obj.Client().Post(context.Background(), u, nil, p)
}

// GetCardHTML 图文消息群发卡券
//
//	HTTP请求方式: POST
//	URL:https://api.weixin.qq.com/card/mpnews/gethtml?access_token=TOKEN
func (obj *OfficialAccount) GetCardHTML(cardID string) Responder {

	u := util.URL(obj.RemoteURL(), api.GetCardMPNewsHTML)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"card_id": cardID,
	})
}

// SetCardTestWhiteListByID 设置测试白名单(by openid)
func (obj *OfficialAccount) SetCardTestWhiteListByID(list ...string) Responder {

	return obj.SetCardTestWhiteList("openid", list)
}

// SetCardTestWhiteListByName 设置测试白名单(by username)
func (obj *OfficialAccount) SetCardTestWhiteListByName(list ...string) Responder {

	return obj.SetCardTestWhiteList("username", list)
}

// SetCardTestWhiteList 设置测试白名单
//
//	HTTP请求方式: POST
//	URL:https://api.weixin.qq.com/card/testwhitelist/set?access_token=TOKEN
func (obj *OfficialAccount) SetCardTestWhiteList(typ string, list []string) Responder {

	u := util.URL(obj.RemoteURL(), api.SetCardTestWhiteList)
	return obj.Client().Post(context.Background(), u, nil, util.Map{typ: list})
}

// CreateCardQrCode 创建二维码
//
//	HTTP请求方式: POST
//	URL:https://api.weixin.qq.com/card/qrcode/create?access_token=TOKEN
func (obj *OfficialAccount) CreateCardQrCode(action *model.QrCodeAction) Responder {

	u := util.URL(obj.RemoteURL(), api.CreateCardQrcode)
	return obj.Client().Post(context.Background(), u, nil, action)
}

// CardCreate 创建卡券
//
//	HTTP请求方式: POST
//	URL: https://api.weixin.qq.com/card/create?access_token=ACCESS_TOKEN
//	type *OneCard or Map
func (obj *OfficialAccount) CardCreate(maps util.MapAble) Responder {

	u := util.URL(obj.RemoteURL(), api.CreateCard)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"card": maps})
}

// GetCard 查看卡券详情
//
//	开发者可以调用该接口查询某个card_id的创建信息、审核状态以及库存数量。
//	接口调用请求说明
//	HTTP请求方式: POSTURL:https://api.weixin.qq.com/card/get?access_token=TOKEN
func (obj *OfficialAccount) GetCard(cardID string) Responder {

	u := util.URL(obj.RemoteURL(), api.GetCard)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"card_id": cardID})
}

// GetCardApplyProtocol 卡券开放类目查询接口
//
//	HTTP请求方式: GET
//	URL:https://api.weixin.qq.com/card/getapplyprotocol?access_token=TOKEN
func (obj *OfficialAccount) GetCardApplyProtocol() Responder {

	u := util.URL(obj.RemoteURL(), api.GetCardApplyProtocol)
	return obj.Client().Get(context.Background(), u, nil)
}

// GetCardColors 卡券开放类目查询接口
//
//	HTTP请求方式: GET
//	URL:https://api.weixin.qq.com/card/getcolors?access_token=TOKEN
func (obj *OfficialAccount) GetCardColors() Responder {

	u := util.URL(obj.RemoteURL(), api.GetCardColors)
	return obj.Client().Get(context.Background(), u, nil)
}

// CheckinBoardingpass 更新飞机票信息接口
//
//	接口调用请求说明
//	http请求方式: POST
//	URL:https://api.weixin.qq.com/card/boardingpass/checkin?access_token=TOKEN
func (obj *OfficialAccount) CheckinBoardingpass(p util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.CheckinCardBoardingpass)
	return obj.Client().Post(context.Background(), u, nil, p)
}

// CardCategories 卡券开放类目查询接口
//
//	接口说明
//	通过调用该接口查询卡券开放的类目ID，类目会随业务发展变更，请每次用接口去查询获取实时卡券类目。
//	注意：
//	1.本接口查询的返回值还有卡券资质ID,此处的卡券资质为：已微信认证的公众号通过微信公众平台申请卡券功能时，所需的资质。
//	2.对于第三方强授权模式，子商户无论选择什么类目，均提交营业执照即可，所以不用考虑此处返回的资质字段，返回值仅参考类目ID即可。
//	接口详情
//	接口调用请求说明
//	https请求方式: GET https://api.weixin.qq.com/card/getapplyprotocol?access_token=TOKEN
func (obj *OfficialAccount) CardCategories() Responder {

	u := util.URL(obj.RemoteURL(), api.GetCardApplyprotocol)
	return obj.Client().Get(context.Background(), u, nil)
}

// GetCardBatch 批量查询卡券列表
//
//	接口调用请求说明
//	HTTP请求方式: POST URL:https://api.weixin.qq.com/card/batchget?access_token=TOKEN
func (obj *OfficialAccount) GetCardBatch(offset, count int, statusList []model.CardStatus) Responder {

	p := util.Map{
		"offset":      offset,
		"count":       count,
		"status_list": statusList,
	}
	u := util.URL(obj.RemoteURL(), api.GetCardBatch)
	return obj.Client().Post(context.Background(), u, nil, p)
}

// UpdateCard 更改卡券信息接口
//
//	接口说明
//	支持更新所有卡券类型的部分通用字段及特殊卡券（会员卡、飞机票、电影票、会议门票）中特定字段的信息。
//	接口调用请求说明
//	HTTP请求方式: POST URL:https://api.weixin.qq.com/card/update?access_token=TOKEN
func (obj *OfficialAccount) UpdateCard(cardID string, p util.Map) Responder {

	p = util.CombineMaps(util.Map{"card_id": cardID}, p)
	u := util.URL(obj.RemoteURL(), api.UpdateCard)
	return obj.Client().Post(context.Background(), u, nil, p)
}

// CardDelete 删除卡券接口
// 删除卡券接口允许商户删除任意一类卡券。删除卡券后，该卡券对应已生成的领取用二维码、添加到卡包JS API均会失效。 注意：如用户在商家删除卡券前已领取一张或多张该卡券依旧有效。即删除卡券不能删除已被用户领取，保存在微信客户端中的卡券。
// 接口调用请求说明
// HTTP请求方式: POST URL:https://api.weixin.qq.com/card/delete?access_token=TOKEN
func (obj *OfficialAccount) CardDelete(cardID string) Responder {

	u := util.URL(obj.RemoteURL(), api.DeleteCard)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"card_id": cardID})
}

// GetUserCardlist ...
func (obj *OfficialAccount) GetUserCardlist(openID, cardID string) Responder {

	u := util.URL(obj.RemoteURL(), api.GetUserCardlist)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"openid": openID, "card_id": cardID})
}

// SetCardPayCell ...
func (obj *OfficialAccount) SetCardPayCell(cardID string, isOpen bool) Responder {

	u := util.URL(obj.RemoteURL(), api.SetCardPayCell)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"is_open": isOpen, "card_id": cardID})
}

// ModifyCardStock ...
func (obj *OfficialAccount) ModifyCardStock(cardID string, option util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.ModifyCardStock)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"card_id": cardID})
}

// GetCardAPITicket get ticket
func (obj *OfficialAccount) GetCardAPITicket(refresh bool) (string, error) {
	jssdk, err := obj.JSSDK()
	if err != nil {
		return "", err
	}
	return jssdk.GetTicket("wx_card", refresh), nil
}

// JSSDK ...
func (obj *OfficialAccount) JSSDK() (*JSSDK, error) {
	if obj.jssdk == nil {
		return nil, errors.New("must add jssdk on new")
	}
	return obj.jssdk, nil
}

/*
OpenComment 打开文章评论
https 请求方式: POST
https://api.weixin.qq.com/cgi-bin/comment/open?access_token=ACCESS_TOKEN
失败:

	{"errcode":88000,"errmsg":"without comment privilege"}
*/
func (obj *OfficialAccount) OpenComment(id, index int) Responder {
	u := util.URL(obj.RemoteURL(), api.OpenComment)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id": id,
		"index":       index,
	})
}

/*
CloseComment 关闭评论
https 请求方式: POST
https://api.weixin.qq.com/cgi-bin/comment/close?access_token=ACCESS_TOKEN
失败:
{"errcode":88000,"errmsg":"without comment privilege"}
*/
func (obj *OfficialAccount) CloseComment(id, index int) Responder {
	u := util.URL(obj.RemoteURL(), api.CloseComment)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id": id,
		"index":       index,
	})
}

/*
ListComment 获取文章评论
https 请求方式: POST
https://api.weixin.qq.com/cgi-bin/comment/list?access_token=ACCESS_TOKEN
失败:
{"errcode":88000,"errmsg":"without comment privilege"}
*/
func (obj *OfficialAccount) ListComment(id, index, begin, count, typ int) Responder {
	u := util.URL(obj.RemoteURL(), api.ListComment)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id": id,
		"index":       index,
		"begin":       begin,
		"count":       count,
		"type":        typ,
	})

}

/*
ElectComment  将评论标记精选
https 请求方式: POST
https://api.weixin.qq.com/cgi-bin/comment/markelect?access_token=ACCESS_TOKEN
参数	是否必须	类型	说明
id	是	int	群发返回的msg_data_id
index	否	int	多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
user_comment_id	是	int	用户评论id
*/
func (obj *OfficialAccount) ElectComment(id, index, userCommentID int) Responder {
	u := util.URL(obj.RemoteURL(), api.ElectComment)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id":     id,
		"index":           index,
		"user_comment_id": userCommentID,
	})
}

/*
UnelectComment 将评论取消精选

	https 请求方式: POST
	https://api.weixin.qq.com/cgi-bin/comment/unmarkelect?access_token=ACCESS_TOKEN

参数	是否必须	类型	说明
id	是	int	群发返回的msg_data_id
index	否	int	多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
user_comment_id	是	int	用户评论id
*/
func (obj *OfficialAccount) UnelectComment(id, index, userCommentID int) Responder {
	u := util.URL(obj.RemoteURL(), api.UnelectComment)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id":     id,
		"index":           index,
		"user_comment_id": userCommentID,
	})
}

/*
DeleteComment 删除评论
https 请求方式: POST
https://api.weixin.qq.com/cgi-bin/comment/delete?access_token=ACCESS_TOKEN
参数	是否必须	类型	说明
id	是	int	群发返回的msg_data_id
index	否	int	多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
user_comment_id	是	int	用户评论id
*/
func (obj *OfficialAccount) DeleteComment(id, index, userCommentID int) Responder {
	u := util.URL(obj.RemoteURL(), api.DeleteComment)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id":     id,
		"index":           index,
		"user_comment_id": userCommentID,
	})
}

/*
AddCommentReply 回复评论
https 请求方式: POST
https://api.weixin.qq.com/cgi-bin/comment/reply/add?access_token=ACCESS_TOKEN
参数	是否必须	类型	说明
id	是	int	群发返回的msg_data_id
index	否	int	多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
user_comment_id	是	int	评论id
content	是	string	回复内容
*/
func (obj *OfficialAccount) AddCommentReply(id, index, userCommentID int, content string) Responder {
	u := util.URL(obj.RemoteURL(), api.AddCommentReply)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id":     id,
		"index":           index,
		"user_comment_id": userCommentID,
		"content":         content,
	})
}

/*
DeleteCommentReply 删除回复
https 请求方式: POST
https://api.weixin.qq.com/cgi-bin/comment/reply/delete?access_token=ACCESS_TOKEN
参数	是否必须	类型	说明
id	是	int	群发返回的msg_data_id
index	否	int	多图文时，用来指定第几篇图文，从0开始，不带默认操作该msg_data_id的第一篇图文
user_comment_id	是	int	评论id
*/
func (obj *OfficialAccount) DeleteCommentReply(id, index, userCommentID int) Responder {
	u := util.URL(obj.RemoteURL(), api.DeleteCommentReply)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"msg_data_id":     id,
		"index":           index,
		"user_comment_id": userCommentID,
	})
}

/*
GetCurrentAutoReplyInfo ...
http请求方式: GET（请使用https协议）
https://api.weixin.qq.com/cgi-bin/get_current_autoreply_info?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetCurrentAutoReplyInfo() Responder {
	u := util.URL(obj.RemoteURL(), api.GetCurrentAutoReplyInfo)
	return obj.Client().Get(context.Background(), u, nil)
}

/*
GetCurrentSelfMenuInfo ...
http请求方式: GET（请使用https协议）
https://api.weixin.qq.com/cgi-bin/get_current_selfmenu_info?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetCurrentSelfMenuInfo() Responder {
	u := util.URL(obj.RemoteURL(), api.GetCurrentSelfMenuInfo)
	return obj.Client().Get(context.Background(), u, nil)
}
func (obj *OfficialAccount) Get(uri, beginDate, endDate string) Responder {
	u := util.URL(obj.RemoteURL(), uri)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"begin_date": beginDate, "end_date": endDate})
}

/*
GetUserSummary 获取用户增减数据（getusersummary）	7
https://api.weixin.qq.com/datacube/getusersummary?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetUserSummary(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUserSummary,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
GetUserCumulate 获取累计用户数据（getusercumulate）	7
https://api.weixin.qq.com/datacube/getusercumulate?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetUserCumulate(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUserCumulate,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
GetArticleSummary 获取图文群发每日数据（getarticlesummary）	1
https://api.weixin.qq.com/datacube/getarticlesummary?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetArticleSummary(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetArticleSummary,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
DataCubeGetArticleTotal 获取图文群发总数据（getarticletotal）	1
https://api.weixin.qq.com/datacube/getarticletotal?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetArticleTotal(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetArticleTotal,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
GetUserRead 获取图文统计数据（getuserread）	3
https://api.weixin.qq.com/datacube/getuserread?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetUserRead(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUserRead,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
GetUserReadHour 获取图文统计分时数据（getuserreadhour）	1
https://api.weixin.qq.com/datacube/getuserreadhour?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetUserReadHour(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUserReadHour,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
GetUserShare 获取图文分享转发数据（getusershare）	7
https://api.weixin.qq.com/datacube/getusershare?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetUserShare(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUserShare,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
DataCubeGetUserShareHour 获取图文分享转发分时数据（getusersharehour）	1
https://api.weixin.qq.com/datacube/getusersharehour?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetUserShareHour(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUserShareHour,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*
DataCubeGetUpstreamMsg 获取消息发送概况数据（getupstreammsg）	7
https://api.weixin.qq.com/datacube/getupstreammsg?access_token=ACCESS_TOKEN
*/
func (obj *OfficialAccount) GetUpstreamMsg(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUpstreamMsg,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

// DataCubeGetUpstreamMsgHour 获取消息分送分时数据（getupstreammsghour）	1
// https://api.weixin.qq.com/datacube/getupstreammsghour?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetUpstreamMsgHour(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUpstreamMsgHour,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

// DataCubeGetUpstreamMsgWeek 获取消息发送周数据（getupstreammsgweek）	30
// https://api.weixin.qq.com/datacube/getupstreammsgweek?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetUpstreamMsgWeek(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUpstreamMsgWeek,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

// DataCubeGetUpstreamMsgDist 获取消息发送分布数据（getupstreammsgdist）	15
// https://api.weixin.qq.com/datacube/getupstreammsgdist?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetUpstreamMsgDist(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUpstreamMsgDist,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

// DataCubeGetUpstreamMsgDistWeek 获取消息发送分布周数据（getupstreammsgdistweek）	30
// https://api.weixin.qq.com/datacube/getupstreammsgdistweek?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetUpstreamMsgDistWeek(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUpstreamMsgDistWeek,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

// DataCubeGetUpstreamMsgDistMonth 获取消息发送分布月数据（getupstreammsgdistmonth）	30
// https://api.weixin.qq.com/datacube/getupstreammsgdistmonth?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetUpstreamMsgDistMonth(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetUpstreamMsgDistMonth,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

// DataCubeGetInterfaceSummary 获取接口分析数据（getinterfacesummary）	30
// https://api.weixin.qq.com/datacube/getinterfacesummary?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetInterfaceSummary(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetInterfaceSummary,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

// DataCubeGetInterfaceSummaryHour 获取接口分析分时数据（getinterfacesummaryhour）	1
// https://api.weixin.qq.com/datacube/getinterfacesummaryhour?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetInterfaceSummaryHour(beginDate, endDate time.Time) Responder {

	return obj.Get(
		api.GetInterfaceSummaryHour,
		beginDate.Format(api.DatacubeTimeLayout),
		endDate.Format(api.DatacubeTimeLayout),
	)
}

/*MediaType MediaType */
type MediaType string

/*media types */
const (
	MediaTypeImage MediaType = "image"
	MediaTypeVoice MediaType = "voice"
	MediaTypeVideo MediaType = "video"
	MediaTypeThumb MediaType = "thumb"
)

/*String transfer MediaType to string */
func (m MediaType) String() string {
	return string(m)
}

// MaterialAddNews 新增永久素材
// http请求方式: POST，https协议
// https://api.weixin.qq.com/cgi-bin/material/add_news?access_token=ACCESS_TOKEN
// func (m *Material) AddNews(articles []*media.Article) core.Responder {
func (obj *OfficialAccount) MaterialAddNews(p util.Map) Responder {

	u := util.URL(obj.RemoteURL(), api.AddNews)
	return obj.Client().Post(context.Background(), u, nil, p)
}

// MaterialAddMaterial 新增其他类型永久素材
// http请求方式: POST，需使用https
// https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=ACCESS_TOKEN&type=TYPE
func (obj *OfficialAccount) MaterialAddMaterial(filePath string, mediaType MediaType) Responder {

	if mediaType == MediaTypeVideo {
		log.Println("please use MaterialUploadVideo() function")
	}
	u := util.URL(obj.RemoteURL(), api.AddMaterial)
	p := obj.AccessToken.KeyMap()
	p.Set("type", mediaType)
	return Upload(u, p, util.Map{"media": filePath})
}

// MaterialUploadVideo 新增其他类型永久素材
// http请求方式: POST，需使用https
// https://api.weixin.qq.com/cgi-bin/material/add_material?access_token=ACCESS_TOKEN&type=TYPE
func (obj *OfficialAccount) MaterialUploadVideo(filePath string, title, introduction string) Responder {

	u := util.URL(obj.RemoteURL(), api.AddMaterial)
	p := obj.AccessToken.KeyMap()
	p.Set("type", MediaTypeVideo)
	return Upload(u, p, util.Map{
		"media": filePath,
		"description": util.Map{
			"title":        title,
			"introduction": introduction,
		}})
}

// GetMaterial 获取永久素材
// http请求方式: POST,https协议
// https://api.weixin.qq.com/cgi-bin/material/get_material?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetMaterial(mediaID string) Responder {

	u := util.URL(obj.RemoteURL(), api.GetMaterial)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"media_id": mediaID})
}

// MaterialDel 删除永久素材
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/material/del_material?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) MaterialDel(mediaID string) Responder {

	u := util.URL(obj.RemoteURL(), api.DelMaterial)
	resp := obj.Client().Post(context.Background(), u, nil, util.Map{"media_id": mediaID})
	return resp

}

/*Article Article */
type Article struct {
	Title              string `json:"title"`                           // 标题
	ThumbMediaID       string `json:"thumb_media_id"`                  // 图文消息的封面图片素材id（必须是永久mediaID）
	Author             string `json:"author,omitempty"`                // 作者
	Digest             string `json:"digest,omitempty"`                // 图文消息的摘要，仅有单图文消息才有摘要，多图文此处为空。如果本字段为没有填写，则默认抓取正文前64个字。
	ShowCoverPic       string `json:"show_cover_pic"`                  // 	是否显示封面，0为false，即不显示，1为true，即显示
	Content            string `json:"content"`                         // 图文消息的具体内容，支持HTML标签，必须少于2万字符，小于1M，且此处会去除JS,涉及图片url必须来源 "上传图文消息内的图片获取URL"接口获取。外部图片url将被过滤。
	ContentSourceURL   string `json:"content_source_url"`              // 图文消息的原文地址，即点击“阅读原文”后的URL
	NeedOpenComment    uint32 `json:"need_open_comment,omitempty"`     // (新增字段）	否	Uint32	是否打开评论，0不打开，1打开
	OnlyFansCanComment uint32 `json:"only_fans_can_comment,omitempty"` // （新增字段）	否	Uint32	是否粉丝才可评论，0所有人可评论，1粉丝才可评论
}

// MaterialUpdateNews 修改永久图文素材
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/material/update_news?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) MaterialUpdateNews(mediaID string, index int, articles []*Article) Responder {

	u := util.URL(obj.RemoteURL(), api.DelNews)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"media_id": mediaID,
		"index":    index,
		"articles": articles,
	})
}

// MaterialGetCount 获取素材总数
// http请求方式: GET
// https://api.weixin.qq.com/cgi-bin/material/get_materialcount?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) MaterialGetCount() Responder {

	u := util.URL(obj.RemoteURL(), api.GetMaterialcount)
	return obj.Client().Get(context.Background(), u, nil)
}

// MaterialBatchGet 获取素材列表
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/material/batchget_material?access_token=ACCESS_TOKEN
// 参数说明
// 参数	是否必须	说明
// type	是	素材的类型，图片（image）、视频（video）、语音 （voice）、图文（news）
// offset	是	从全部素材的该偏移位置开始返回，0表示从第一个素材 返回
// count	是	返回素材的数量，取值在1到20之间
func (obj *OfficialAccount) MaterialBatchGet(mediaType MediaType, offset, count int) Responder {

	u := util.URL(obj.RemoteURL(), api.ListMaterial)
	return obj.Client().Post(context.Background(), u, nil, util.Map{
		"type":   mediaType.String(),
		"offset": offset,
		"count":  count,
	})
}

/*
MediaUpload 媒体文件上传接口
https://api.weixin.qq.com/cgi-bin/media/upload?access_token=ACCESS_TOKEN&type=TYPE
参数	是否必须	说明
access_token	是	调用接口凭证
type	是	媒体文件类型，分别有图片（image）、语音（voice）、视频（video）和缩略图（thumb）
media	是	form-data中媒体文件标识，有filename、filelength、content-type等信息
*/
func (obj *OfficialAccount) MediaUpload(filePath string, mediaType MediaType) Responder {

	u := util.URL(obj.RemoteURL(), api.UploadMedia)
	p := obj.AccessToken.KeyMap()
	p.Set("type", mediaType)
	return Upload(u, p, util.Map{"media": filePath})
}

/*
MediaUploadThumb UploadVoice
see Upload
*/
func (obj *OfficialAccount) MediaUploadThumb(filePath string) Responder {
	return obj.MediaUpload(filePath, MediaTypeThumb)
}

/*
MediaUploadVoice UploadVoice
see Upload
*/
func (obj *OfficialAccount) MediaUploadVoice(filePath string) Responder {
	return obj.MediaUpload(filePath, MediaTypeVoice)
}

/*
MediaUploadVideo UploadVideo
see Upload
*/
func (obj *OfficialAccount) MediaUploadVideo(filePath string) Responder {
	return obj.MediaUpload(filePath, MediaTypeVideo)
}

/*
MediaUploadImage UploadImage
see Upload
*/
func (obj *OfficialAccount) MediaUploadImage(filePath string) Responder {
	return obj.MediaUpload(filePath, MediaTypeImage)
}

/*
MediaGet 获取临时素材
http请求方式: GET,https调用
https://api.weixin.qq.com/cgi-bin/media/get?access_token=ACCESS_TOKEN&media_id=MEDIA_ID
请求示例（示例为通过curl命令获取多媒体文件）
curl -I -G "https://api.weixin.qq.com/cgi-bin/media/get?access_token=ACCESS_TOKEN&media_id=MEDIA_ID"
*/
func (obj *OfficialAccount) MediaGet(mediaID string) Responder {

	u := util.URL(obj.RemoteURL(), api.GetMedia)
	return obj.Client().Get(context.Background(), u, util.Map{"media_id": mediaID})
}

// MediaGetJSSDK 高清语音素材获取接口
// http请求方式: GET,https调用
// https://api.weixin.qq.com/cgi-bin/media/get/jssdk?access_token=ACCESS_TOKEN&media_id=MEDIA_ID
func (obj *OfficialAccount) MediaGetJSSDK(mediaID string) Responder {

	u := util.URL(obj.RemoteURL(), api.GetMediaJssdk)
	return obj.Client().Get(context.Background(), u, util.Map{"media_id": mediaID})
}
func (obj *OfficialAccount) UploadImg(name string, filePath string) Responder {

	u := util.URL(obj.RemoteURL(), api.UploadImg)
	p := obj.AccessToken.KeyMap()
	return Upload(u, p, util.Map{name: filePath})
}

// MediaUploadImg 上传图文消息内的图片获取URL
// http请求方式: POST，https协议
// https://api.weixin.qq.com/cgi-bin/media/uploadimg?access_token=ACCESS_TOKEN
// 调用示例（使用curl命令，用FORM表单方式上传一个图片）:
// curl -F media=@test.jpg "https://api.weixin.qq.com/cgi-bin/media/uploadimg?access_token=ACCESS_TOKEN"
func (obj *OfficialAccount) MediaUploadImg(filePath string) Responder {
	return obj.UploadImg("media", filePath)
}

// MediaUploadImgBuffer 上传图片接口
// HTTP请求方式: POST/FROM
// URL:https://api.weixin.qq.com/cgi-bin/media/uploadimg?access_token=ACCESS_TOKEN
// 调用示例（使用curl命令，用FORM表单方式上传一个图片）:curl –Fbuffer=@test.jpg
func (obj *OfficialAccount) MediaUploadImgBuffer(filePath string) Responder {
	return obj.UploadImg("buffer", filePath)
}

// MenuCreate 创建菜单
// 个性化创建
// https://api.weixin.qq.com/cgi-bin/menu/addconditional?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) MenuCreate(buttons *Button) Responder {

	u := util.URL(obj.RemoteURL(), api.AddMenuConditional)
	if buttons.GetMatchRule() == nil {
		u = util.URL(obj.RemoteURL(), api.MenuCreate)
	}
	return obj.Client().Post(context.Background(), u, nil, buttons)
}

/*
MenuList 自定义菜单查询接口
请求说明
http请求方式:GET
https://api.weixin.qq.com/cgi-bin/menu/get?access_token=ACCESS_TOKEN
返回说明（无个性化菜单时）
参考URL:https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1421141014
*/
func (obj *OfficialAccount) MenuList() Responder {

	u := util.URL(obj.RemoteURL(), api.GetMenu)
	return obj.Client().Get(context.Background(), u, nil)
}

/*
MenuCurrent 获取自定义菜单配置接口
接口调用请求说明
http请求方式: GET（请使用https协议）https://api.weixin.qq.com/cgi-bin/get_current_selfmenu_info?access_token=ACCESS_TOKEN
返回结果说明
参考URL:https://mp.weixin.qq.com/wiki?t=resource/res_main&id=mp1434698695
*/
func (obj *OfficialAccount) MenuCurrent() Responder {

	u := util.URL(obj.RemoteURL(), api.GetCurrentSelfMenuInfo)
	return obj.Client().Get(context.Background(), u, nil)
}

/*
MenuTryMatch 测试个性化菜单匹配结果
http请求方式:POST（请使用https协议）
https://api.weixin.qq.com/cgi-bin/menu/trymatch?access_token=ACCESS_TOKEN
请求示例
{"user_id":"weixin"}
user_id可以是粉丝的OpenID，也可以是粉丝的微信号。
返回结果 该接口将返回菜单配置，示例如下:

	{
	    "button": [
	        {
	            "type": "view",
	            "name": "tx",
	            "url": "http://www.qq.com/",
	            "sub_button": [ ]
	        },
	        {
	            "type": "view",
	            "name": "tx",
	            "url": "http://www.qq.com/",
	            "sub_button": [ ]
	        },
	        {
	            "type": "view",
	            "name": "tx",
	            "url": "http://www.qq.com/",
	            "sub_button": [ ]
	        }
	    ]
	}
*/
func (obj *OfficialAccount) MenuTryMatch(userID string) Responder {

	u := util.URL(obj.RemoteURL(), api.TryMatchMenu)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"user_id": userID})
}

/*
MenuDelete 自定义菜单删除接口
请求说明
http请求方式:GET
https://api.weixin.qq.com/cgi-bin/menu/delete?access_token=ACCESS_TOKEN
返回说明
对应创建接口，正确的Json返回结果:
{"errcode":0,"errmsg":"ok"}
*/
func (obj *OfficialAccount) MenuDelete(menuID int) Responder {

	u := util.URL(obj.RemoteURL(), api.DeleteMenuConditional)
	if menuID == 0 {
		u = util.URL(obj.RemoteURL(), api.DeleteMenu)
		return obj.Client().Get(context.Background(), u, nil)
	}
	return obj.Client().Post(context.Background(), u, nil, util.Map{"menuid": menuID})
}

/*
POIAdd 创建门店
http请求方式	POST/FORM
请求Url	https://api.weixin.qq.com/cgi-bin/poi/addpoi?access_token=ACCESS_TOKEN
POST数据格式	buffer
*/
func (obj *OfficialAccount) POIAdd(biz *model.PoiBaseInfo) Responder {

	u := util.URL(obj.RemoteURL(), api.AddPoi)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"business": biz})
}

/*
POIGet 查询门店信息
http请求方式	POST
请求Url	http://api.weixin.qq.com/cgi-bin/poi/getpoi?access_token=TOKEN
POST数据格式	json
*/
func (obj *OfficialAccount) POIGet(id string) Responder {

	u := util.URL(obj.RemoteURL(), api.AddPoi)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"poi_id": id})
}

/*
POIUpdate 修改门店服务信息
http请求方式	POST/FROM
请求Url	https://api.weixin.qq.com/cgi-bin/poi/updatepoi?access_token=TOKEN
POST数据格式	buffer
字段说明:
全部字段内容同前。
特别注意:
以上8个字段，若有填写内容则为覆盖更新，若无内容则视为不修改，维持原有内容。 photo_list 字段为全列表覆盖，若需要增加图片，需将之前图片同样放入list 中，在其后增加新增图片。如:已有A、B、C 三张图片，又要增加D、E 两张图，则需要调用该接口，photo_list 传入A、B、C、D、E 五张图片的链接。
成功返回:
{
"errcode":0,
"errmsg":"ok"
}
*/
func (obj *OfficialAccount) POIUpdate(biz *model.PoiBaseInfo) Responder {

	u := util.URL(obj.RemoteURL(), api.UpdatePoi)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"business": biz})

}

/*
POIGetList 查询门店列表

	http请求方式	POST
	请求Url	https://api.weixin.qq.com/cgi-bin/poi/getpoilist?access_token=TOKEN
	POST数据格式	json
	字段	说明	是否必填
	begin	开始位置，0 即为从第一条开始查询	是
	limit	返回数据条数，最大允许50，默认为20	是

成功返回:

	{
	    "errcode":0,
	    "errmsg":"ok"
	    "business_list":[
	    {"base_info":{
	    "sid":"101",
	    "business_name":"麦当劳",
	    "branch_name":"艺苑路店",
	    "address":"艺苑路11号",
	    "telephone":"020-12345678",
	    "categories":["美食,快餐小吃"],
	    "city":"广州市",
	    "province":"广东省",
	    "offset_type":1,
	    "longitude":115.32375,
	    "latitude":25.097486,
	    "photo_list":[{"photo_url":"http: ...."}],
	    "introduction":"麦当劳是全球大型跨国连锁餐厅，1940 年创立于美国，在世界上大约拥有3 万间分店。主要售卖汉堡包，以及薯条、炸鸡、汽水、冰品、沙拉、水果等快餐食品",
	    "recommend":"麦辣鸡腿堡套餐，麦乐鸡，全家桶",
	    "special":"免费wifi，外卖服务",
	    "open_time":"8:00-20:00",
	    "avg_price":35,
	    "poi_id":"285633617",
	    "available_state":3,
	    "district":"海珠区",
	    "update_status":0
	}},

	{"base_info":{
	    "sid":"101",
	    "business_name":"麦当劳",
	    "branch_name":"北京路店",
	    "address":"北京路12号",
	    "telephone":"020-12345689",
	    "categories":["美食,快餐小吃"],
	    "city":"广州市",
	    "province":"广东省",
	    "offset_type":1,
	    "longitude":115.3235,
	    "latitude":25.092386,
	    "photo_list":[{"photo_url":"http: ...."}],
	    "introduction":"麦当劳是全球大型跨国连锁餐厅，1940 年创立于美国，在世界上大约拥有3 万间分店。主要售卖汉堡包，以及薯条、炸鸡、汽水、冰品、沙拉、水果等快餐食品",
	    "recommend":"麦辣鸡腿堡套餐，麦乐鸡，全家桶",
	    "special":"免费wifi，外卖服务",
	    "open_time":"8:00-20:00",
	    "avg_price":35,
	    "poi_id":"285633618",
	    "available_state":4,
	    "district":"越秀区",
	    "update_status":0
	}},

	{"base_info":{
	    "sid":"101",
	    "business_name":"麦当劳",
	    "branch_name":"龙洞店",
	    "address":"迎龙路122号",
	    "telephone":"020-12345659",
	    "categories":["美食,快餐小吃"],
	    "city":"广州市",
	    "province":"广东省",
	    "offset_type":1,
	    "longitude":115.32345,
	    "latitude":25.056686,
	    "photo_list":[{"photo_url":"http: ...."}],
	    "introduction":"麦当劳是全球大型跨国连锁餐厅，1940 年创立于美国，在世界上大约拥有3 万间分店。主要售卖汉堡包，以及薯条、炸鸡、汽水、冰品、沙拉、水果等快餐食品",
	    "recommend":"麦辣鸡腿堡套餐，麦乐鸡，全家桶",
	    "special":"免费wifi，外卖服务",
	    "open_time":"8:00-20:00",
	    "avg_price":35,
	    "poi_id":"285633619",
	    "available_state":2,
	    "district":"天河区",
	    "update_status":0
	}},

],
"total_count":"3",
}
*/
func (obj *OfficialAccount) POIGetList(begin int, limit int) Responder {

	u := util.URL(obj.RemoteURL(), api.GetListPoi)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"begin": begin, "limit": limit})

}

/*
POIDel 删除门店
协议	https
http请求方式	POST/FROM
请求Url	https://api.weixin.qq.com/cgi-bin/poi/delpoi?access_token=TOKEN
POST数据格式	buffer
*/
func (obj *OfficialAccount) POIDel(poiID string) Responder {

	u := util.URL(obj.RemoteURL(), api.DelPoi)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"poi_id": poiID})
}

/*
POIGetCategory 门店类目表
http请求方式	GET
请求Url	http://api.weixin.qq.com/cgi-bin/poi/getwxcategory?access_token=TOKEN
成功返回:
{
"category_list":
["美食,江浙菜,上海菜","美食,江浙菜,淮扬菜","美食,江浙菜,浙江菜","美食,江浙菜,南京菜 ","美食,江浙菜,苏帮菜…"]
}
*/
func (obj *OfficialAccount) POIGetCategory() Responder {

	u := util.URL(obj.RemoteURL(), api.GetWXCategory)
	return obj.Client().Get(context.Background(), u, nil)
}

// TagCreate 创建标签
// http请求方式:POST（请使用https协议）
// https://api.weixin.qq.com/cgi-bin/tags/create?access_token=ACCESS_TOKEN
// 成功:
// {"tag":{"id":100,"name":"testtag"}}
func (obj *OfficialAccount) TagCreate(name string) Responder {

	u := util.URL(obj.RemoteURL(), api.CreateTags)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"tag": util.Map{"name": name}})
}

// TagGet 获取公众号已创建的标签
// http请求方式:GET（请使用https协议）
// https://api.weixin.qq.com/cgi-bin/tags/get?access_token=ACCESS_TOKEN
// 成功:
// {"tags":[{"id":2,"name":"星标组","count":0},{"id":100,"name":"testtag","count":0}]}
func (obj *OfficialAccount) TagGet() Responder {

	u := util.URL(obj.RemoteURL(), api.GetTags)
	return obj.Client().Get(context.Background(), u, nil)
}

// QrCodeCreate 创建二维码ticket
//
//	http请求方式: POST
//	URL: https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=TOKEN
//	POST数据格式:json
//	POST数据例子:{"expire_seconds": 604800, "action_name": "QR_SCENE", "action_info": {"scene": {"scene_id": 123}}}
//	或者也可以使用以下POST数据创建字符串形式的二维码参数:
//	{"expire_seconds": 604800, "action_name": "QR_STR_SCENE", "action_info": {"scene": {"scene_str": "test"}}}
//	http请求方式: POST
//	URL: https://api.weixin.qq.com/cgi-bin/qrcode/create?access_token=TOKEN
//	POST数据格式:json
//	POST数据例子:{"action_name": "QR_LIMIT_SCENE", "action_info": {"scene": {"scene_id": 123}}}
//	或者也可以使用以下POST数据创建字符串形式的二维码参数:
//	{"action_name": "QR_LIMIT_STR_SCENE", "action_info": {"scene": {"scene_str": "test"}}}
//	成功:
//	{"ticket":"gQFy7zwAAAAAAAAAAS5odHRwOi8vd2VpeGluLnFxLmNvbS9xLzAyOE1nSDFvTHdkeWkxeVNqTnhxMTcAAgR6E7FaAwQ8AAAA","expire_seconds":60,"url":"http:\/\/weixin.qq.com\/q\/028MgH1oLwdyi1ySjNxq17"}
func (obj *OfficialAccount) QrCodeCreate(action *model.QrCodeAction) Responder {
	//TODO: need fix

	u := util.URL(obj.RemoteURL(), api.CreateQrcode)
	return obj.Client().Post(context.Background(), u, nil, action)
}

// QrCodeShow 显示二维码
// HTTP GET请求（请使用https协议）https://mp.weixin.qq.com/cgi-bin/showqrcode?ticket=TICKET
// 提醒:使用core.SaveTo保存文件
func (obj *OfficialAccount) QrCodeShow(ticket string) Responder {

	u := util.URL(obj.RemoteURL(), api.ShowQrcode)
	return Get(u, util.Map{"ticket": url.QueryEscape(ticket)})
}

// SetIndustryTemplate 设置所属行业
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/template/api_set_industry?access_token=ACCESS_TOKEN
// 主行业	副行业	代码
// IT科技	互联网/电子商务	1
// IT科技	IT软件与服务	2
// IT科技	IT硬件与设备	3
// IT科技	电子技术	4
// IT科技	通信与运营商	5
// IT科技	网络游戏	6
// 金融业	银行	7
// 金融业	基金理财信托	8
// 金融业	保险	9
// 餐饮	餐饮	10
// 酒店旅游	酒店	11
// 酒店旅游	旅游	12
// 运输与仓储	快递	13
// 运输与仓储	物流	14
// 运输与仓储	仓储	15
// 教育	培训	16
// 教育	院校	17
// 政府与公共事业	学术科研	18
// 政府与公共事业	交警	19
// 政府与公共事业	博物馆	20
// 政府与公共事业	公共事业非盈利机构	21
// 医药护理	医药医疗	22
// 医药护理	护理美容	23
// 医药护理	保健与卫生	24
// 交通工具	汽车相关	25
// 交通工具	摩托车相关	26
// 交通工具	火车相关	27
// 交通工具	飞机相关	28
// 房地产	建筑	29
// 房地产	物业	30
// 消费品	消费品	31
// 商业服务	法律	32
// 商业服务	会展	33
// 商业服务	中介服务	34
// 商业服务	认证	35
// 商业服务	审计	36
// 文体娱乐	传媒	37
// 文体娱乐	体育	38
// 文体娱乐	娱乐休闲	39
// 印刷	印刷	40
// 其它	其它	41
func (obj *OfficialAccount) SetIndustryTemplate(id1, id2 string) Responder {

	u := util.URL(api.SetIndustryTemplate)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"industry_id1": id1, "industry_id2": id2})
}

// GetIndustryTemplate 获取设置的行业信息
// http请求方式:GET
// https://api.weixin.qq.com/cgi-bin/template/get_industry?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) GetIndustryTemplate() Responder {

	u := util.URL(api.GetIndustryTemplate)
	return obj.Client().Get(context.Background(), u, nil)
}

// TemplateAdd 获得模板ID
// 获取模板:https://mp.weixin.qq.com/advanced/tmplmsg?action=list&t=tmplmsg/list&token=93895307&lang=zh_CN
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/template/api_add_template?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) TemplateAdd(shortID string) Responder {

	u := util.URL(api.AddTemplate)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"template_id_short": shortID})
}

// TemplateSend 发送模板消息
// http请求方式: POST
// https://api.weixin.qq.com/cgi-bin/message/template/send?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) TemplateSend(template *model.Template) Responder {

	u := util.URL(api.SendMessageTemplate)
	return obj.Client().Post(context.Background(), u, nil, template)
}

// TemplateGetAllPrivate 获取模板列表
// url:https://api.weixin.qq.com/cgi-bin/template/get_all_private_template?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) TemplateGetAllPrivate() Responder {

	u := util.URL(api.GetAllPrivateTemplate)
	return obj.Client().Get(context.Background(), u, nil)
}

// TemplateDelAllPrivate 删除模板
// url:https://api.weixin.qq.com/cgi-bin/template/del_private_template?access_token=ACCESS_TOKEN
func (obj *OfficialAccount) TemplateDelAllPrivate(templateID string) Responder {
	u := util.URL(api.DelPrivateTemplate)
	return obj.Client().Post(context.Background(), u, nil, util.Map{"template_id": templateID})
}

// TicketGet 获取api_ticket
// http请求方式: GET
// https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=ACCESS_TOKEN&type=wx_card
// 成功:
// {"errcode":0,"errmsg":"ok","ticket":"9KwiourQPRN3vx3Nn1c_iX9qGaI3Cf8dwVy7qqYeYKcd3BK4Zd_jSlol7E7baUfgOY0E2ybaw2OrlhkChKaS7w","expires_in":7200}
func (obj *OfficialAccount) TicketGet(typ string) Responder {

	u := util.URL(api.GetTicket)
	return obj.Client().Get(context.Background(), u, util.Map{"type": typ})
}
