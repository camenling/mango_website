package install

import (
	"net/url"
	"strings"
)

func ItmsInstallURL(plistURL string) string {
	plistURL = strings.TrimSpace(plistURL)
	if plistURL == "" {
		return ""
	}
	return "itms-services://?action=download-manifest&url=" + url.QueryEscape(plistURL)
}
