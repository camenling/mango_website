package model

type InstallPlist struct {
	Name       string
	BundleID   string
	Version    string
	IconUrl    string
	AppFileUrl string
}

type SignInput struct {
	P12Path             string
	MobileprovisionPath string
	Password            string
	Theme               int    // 1–5，对应五套 App Icon / 站点主题
	Accent              string // 强调色 #RRGGBB，空或不合法时用风格默认色
	Glass               bool   // 玻璃渐变开关（iOS 26 以上生效）
	AppSource           string // 明文内置软件源，http/https，可多行
}

type SignResult struct {
	TaskId        string `json:"taskId"`
	InstallUrl    string `json:"installUrl"` // 对外短链 https://域名/i/xxxx
	PlistUrl      string `json:"plistUrl"`   // getInstallPlist 地址（供调试）
	IpaUrl        string `json:"ipaUrl"`
	ItmsUrl       string `json:"itmsUrl,omitempty"` // 完整 itms-services（一般不展示）
	Message       string `json:"message"`
	P12Name       string `json:"p12Name,omitempty"`
	ProvisionName string `json:"provisionName,omitempty"`
}

type CertCheckInput struct {
	P12Path             string
	MobileprovisionPath string
	Password            string
}

type CertInfo struct {
	Status          string `json:"status"`
	TeamName        string `json:"teamName"`
	TeamId          string `json:"teamId"`
	ExpiryDate      string `json:"expiryDate"`
	AppName         string `json:"appName"`
	BundleId        string `json:"bundleId"`
	ProvisionExpiry string `json:"provisionExpiry"`
	Message         string `json:"message"`
}
