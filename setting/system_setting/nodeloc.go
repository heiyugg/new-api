package system_setting

import "one-api/setting/config"

type NodelocOAuthSettings struct {
	Enabled        bool   `json:"enabled"`
	ClientId       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	RedirectUri    string `json:"redirect_uri"`
	AuthEndpoint   string `json:"auth_endpoint"`
	TokenEndpoint  string `json:"token_endpoint"`
	UserInfoEndpoint string `json:"user_info_endpoint"`
}

// 默认配置
var defaultNodelocOAuthSettings = NodelocOAuthSettings{
	Enabled:        false,
	ClientId:       "",
	ClientSecret:   "",
	RedirectUri:    "",
	AuthEndpoint:   "https://conn.nodeloc.cc/oauth2/auth",
	TokenEndpoint:  "https://conn.nodeloc.cc/oauth2/token",
	UserInfoEndpoint: "https://conn.nodeloc.cc/oauth2/userinfo",
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("nodeloc_oauth", &defaultNodelocOAuthSettings)
}

func GetNodelocOAuthSettings() *NodelocOAuthSettings {
	return &defaultNodelocOAuthSettings
}