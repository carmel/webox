package api

const ClearQuota = "/cgi-bin/clear_quota"
const GetCallbackIP = "/cgi-bin/getcallbackip"
const NewSandbox = "sandboxnew"
const GetSignKey = "pay/getsignkey"

// APIMCHUS ...
const APIMCHUS = "https://apius.mch.weixin.qq.com"

// APIMCHHK ...
const APIMCHHK = "https://apihk.mch.weixin.qq.com"

// APIMCHDefault ...
const APIMCHDefault = "https://api.mch.weixin.qq.com"

const ApiWeixin = "https://api.weixin.qq.com"
const Oauth2Authorize = "https://open.weixin.qq.com/connect/oauth2/authorize"
const Oauth2AccessToken = "https://api.weixin.qq.com/sns/oauth2/access_token"
const SnsUserinfo = "https://api.weixin.qq.com/sns/userinfo"

const RiskGetPublicKey = "https://fraud.mch.weixin.qq.com/risk/getpublickey"

const MchSubMchManage = "/secapi/mch/submchmanage"
const MchModifymchinfo = "/secapi/mch/modifymchinfo"
const MktAddrecommendconf = "/secapi/mkt/addrecommendconf"
const MchAddSubDevConfig = "/secapi/mch/addsubdevconfig"

const MmpaymkttransfersSendRedPack = "/mmpaymkttransfers/sendredpack"
const MmpaymkttransfersGetHbInfo = "/mmpaymkttransfers/gethbinfo"
const MmpaymkttransfersSendGroupRedPack = "/mmpaymkttransfers/sendgroupredpack"
const MmpaymkttransfersGetTransferInfo = "/mmpaymkttransfers/gettransferinfo"
const MmpaymkttransfersPromotionTransfers = "/mmpaymkttransfers/promotion/transfers"

const MmpaymkttransfersSendCoupon = "/mmpaymkttransfers/send_coupon"
const MmpaymkttransfersQueryCouponStock = "/mmpaymkttransfers/query_coupon_stock"
const MmpaymkttransfersQueryCouponsInfo = "/mmpaymkttransfers/querycouponsinfo"

const MmpaysptransQueryBank = "/mmpaysptrans/query_bank"
const MmpaysptransPayBank = "/mmpaysptrans/pay_bank"

const Sandbox = "/sandboxnew"
const SandboxSignKey = Sandbox + "/pay/getsignkey"

// BizPayURL ...
const BizPayURL = "weixin://wxpay/bizpayurl?"

const AuthCodeToOpenid = "/tools/authcodetoopenid"
const BatchQueryComment = "/billcommentsp/batchquerycomment"
const PayDownloadBill = "/pay/downloadbill"
const PayDownloadFundFlow = "/pay/downloadfundflow"
const PaySettlementquery = "/pay/settlementquery"
const PayQueryexchagerate = "pay/queryexchagerate"
const PayUnifiedOrder = "/pay/unifiedorder"
const PayOrderQuery = "/pay/orderquery"
const PayMicroPay = "/pay/micropay"
const PayCloseOrder = "/pay/closeorder"
const PayRefundQuery = "/pay/refundquery"

const PayReverse = "/secapi/pay/reverse"
const PayRefund = "/secapi/pay/refund"

// GetTicket api address suffix
const GetTicket = "/cgi-bin/ticket/getticket"

const WeboxLocal = "http://localhost"
const NotifyCB = "notify_cb"
const RefundedCB = "refunded_cb"
const ScannedCB = "scanned_cb"
const DefaultKeepAlive = 30
const DefaultTimeout = 30

/*AccessTokenKey 键值 */
const AccessTokenKey = "access_token"
const AccessToken = "/cgi-bin/token"

const GetKFList = "/cgi-bin/customservice/getkflist"

const MenuCreate = "/cgi-bin/menu/create"
const GetMenu = "/cgi-bin/menu/get"
const DeleteMenu = "/cgi-bin/menu/delete"
const AddMenuConditional = "/cgi-bin/menu/addconditional"
const DeleteMenuConditional = "/cgi-bin/menu/delconditional"
const TryMatchMenu = "/cgi-bin/menu/trymatch"

const SetIndustryTemplate = "/cgi-bin/template/api_set_industry"
const GetIndustryTemplate = "/cgi-bin/template/get_industry"
const AddTemplate = "/cgi-bin/template/api_add_template"
const GetAllPrivateTemplate = "/cgi-bin/template/get_all_private_template"
const DelPrivateTemplate = "/cgi-bin/template/del_private_template"
const SendMessageTemplate = "/cgi-bin/message/template/send"

const UploadMedia = "/cgi-bin/media/upload"
const UploadImg = "/cgi-bin/media/uploadimg"
const GetMedia = "/cgi-bin/media/get"
const GetMediaJssdk = "/cgi-bin/media/get/jssdk"

const CreateTags = "/cgi-bin/tags/create"
const GetTags = "/cgi-bin/tags/get"
const TagsUpdate = "/cgi-bin/tags/update"
const TagsDelete = "/cgi-bin/tags/delete"

const TagsMembersBatchTagging = "/cgi-bin/tags/members/batchtagging"
const TagsMembersBatchUntagging = "/cgi-bin/tags/members/batchuntagging"
const TagsGetIDList = "/cgi-bin/tags/getidlist"
const TagsMembersGetBlackList = "/cgi-bin/tags/members/getblacklist"
const TagsMembersBatchBlackList = "/cgi-bin/tags/members/batchblacklist"
const TagsMembersBatchUnblackList = "/cgi-bin/tags/members/batchunblacklist"

const UserTagGet = "/cgi-bin/user/tag/get"
const UserInfoUpdateRemark = "/cgi-bin/user/info/updateremark"
const UserInfo = "/cgi-bin/user/info"
const UserInfoBatchGet = "/cgi-bin/user/info/batchget"
const UserGet = "/cgi-bin/user/get"

const CreateQrcode = "/cgi-bin/qrcode/create"
const ShowQrcode = "/cgi-bin/showqrcode"

const SendMessageMass = "/cgi-bin/message/mass/send"
const MessageMassSendall = "/cgi-bin/message/mass/sendall"
const MessageMassPreview = "cgi-bin/message/mass/preview"
const DeleteMessageMass = "/cgi-bin/message/mass/delete"
const GetMessageMass = "/cgi-bin/message/mass/get"

// DatacubeTimeLayout time format for datacube
const DatacubeTimeLayout = "2006-01-02"

// const Tags_members_batchuntagging_URL_SUFFIX = "/cgi-bin/tags/members/batchuntagging"
// const Tags_members_batchtagging_URL_SUFFIX = "/cgi-bin/tags/members/batchtagging"
// const Tags_members_batchuntagging_URL_SUFFIX = "/cgi-bin/tags/members/batchuntagging"
const GetUserSummary = "/datacube/getusersummary"
const GetUserCumulate = "/datacube/getusercumulate"
const GetArticleSummary = "/datacube/getarticlesummary"
const GetArticleTotal = "/datacube/getarticletotal"
const GetUserRead = "/datacube/getuserread"
const GetUserReadHour = "/datacube/getuserreadhour"
const GetUserShare = "/datacube/getusershare"
const GetUserShareHour = "/datacube/getusersharehour"

const GetUpstreamMsg = "/datacube/getupstreammsg"
const GetUpstreamMsgHour = "/datacube/getupstreammsghour"
const GetUpstreamMsgWeek = "/datacube/getupstreammsgweek"
const GetUpstreamMsgDist = "/datacube/getupstreammsgdist"
const GetUpstreamMsgMonth = "/datacube/getupstreammsgmonth"
const GetUpstreamMsgDistWeek = "/datacube/getupstreammsgdistweek"
const GetUpstreamMsgDistMonth = "/datacube/getupstreammsgdistmonth"
const GetInterfaceSummary = "/datacube/getinterfacesummary"
const GetInterfaceSummaryHour = "/datacube/getinterfacesummaryhour"

const AddNews = "/cgi-bin/material/add_news"
const AddMaterial = "/cgi-bin/material/add_material"
const GetMaterial = "/cgi-bin/material/get_material"
const DelMaterial = "/cgi-bin/material/del_material"
const DelNews = "/cgi-bin/material/update_news"
const GetMaterialcount = "/cgi-bin/material/get_materialcount"
const ListMaterial = "/cgi-bin/material/batchget_material"
const OpenComment = "/cgi-bin/comment/open"
const CloseComment = "/cgi-bin/comment/close"
const ListComment = "/cgi-bin/comment/list"
const ElectComment = "/cgi-bin/comment/markelect"
const UnelectComment = "/cgi-bin/comment/unmarkelect"
const DeleteComment = "/cgi-bin/comment/delete"
const AddCommentReply = "/cgi-bin/comment/reply/add"
const DeleteCommentReply = "/cgi-bin/comment/reply/delete"

// const Oauth2AccessToken = "/sns/oauth2/access_token"
const Oauth2RefreshToken = "/sns/oauth2/refresh_token"
const Oauth2Userinfo = "/sns/userinfo"
const Oauth2Auth = "/sns/auth"
const DefaultOauthRedirect = "/oauth_redirect"
const SnsapiBase = "snsapi_base"
const SnsapiUserinfo = "snsapi_userinfo"
const CreateCardLandingPage = "/card/landingpage/create"
const DepositCardCode = "/card/code/deposit"
const CountDepositCardCode = "/card/code/getdepositcount"
const CreateCardQrcode = "/card/qrcode/create"
const CheckCardCode = "/card/code/checkcode"
const GetCardCode = "/card/code/get"
const GetCardMPNewsHTML = "/card/mpnews/gethtml"
const SetCardTestWhiteList = "/card/testwhitelist/set"
const CreateCard = "/card/create"
const GetCard = "/card/get"
const GetCardApplyProtocol = "/card/getapplyprotocol"
const GetCardColors = "/card/getcolors"
const GetCardApplyprotocol = "/card/getapplyprotocol"
const GetCardBatch = "/card/batchget"
const UpdateCard = "/card/update"
const DeleteCard = "/card/delete"
const GetUserCardlist = "/card/user/getcardlist"
const SetCardPayCell = "card/paycell/set"
const ModifyCardStock = "card/modifystock"
const CheckinCardBoardingpass = "/card/boardingpass/checkin"
const AddPoi = "/cgi-bin/poi/addpoi"
const PoiGetPoi = "/cgi-bin/poi/getpoi"
const UpdatePoi = "/cgi-bin/poi/updatepoi"
const GetListPoi = "/cgi-bin/poi/getpoilist"
const DelPoi = "/cgi-bin/poi/delpoi"
const GetWXCategory = "/cgi-bin/poi/getwxcategory"
const GetCurrentAutoReplyInfo = "/cgi-bin/get_current_autoreply_info"
const GetCurrentSelfMenuInfo = "/cgi-bin/get_current_selfmenu_info"

// POST ...
const POST = "POST"

// GET ...
const GET = "GET"
