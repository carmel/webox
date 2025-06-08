package webox

// Config 配置文件，用来生成Property各种属性
type Config struct {
	AppID       string
	AppSecret   string
	MchID       string
	MchKey      string
	PemCert     []byte
	PemKEY      []byte
	RootCA      []byte
	Token       string
	AesKey      string
	Scopes      []string
	RedirectURI string
}
