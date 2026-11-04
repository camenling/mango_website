package install

import "mangosign/internal/model"

func MakeInstallPlist(info model.InstallPlist) string {
	icon := info.IconUrl
	if icon == "" {
		return simplePlist(info)
	}
	return fullPlist(info)
}

func simplePlist(info model.InstallPlist) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict>
  <key>items</key>
  <array>
    <dict>
      <key>assets</key>
      <array>
        <dict>
          <key>kind</key>
          <string>software-package</string>
          <key>url</key>
          <string><![CDATA[` + info.AppFileUrl + `]]></string>
        </dict>
      </array>
      <key>metadata</key>
      <dict>
        <key>bundle-identifier</key>
        <string>` + info.BundleID + `</string>
        <key>bundle-version</key>
        <string><![CDATA[` + info.Version + `]]></string>
        <key>kind</key>
        <string>software</string>
        <key>title</key>
        <string><![CDATA[` + info.Name + `]]></string>
      </dict>
    </dict>
  </array>
</dict>
</plist>`
}

func fullPlist(info model.InstallPlist) string {
	return `<?xml version="1.0" encoding="UTF-8"?>
<plist version="1.0"><dict>
  <key>items</key>
  <array>
    <dict>
      <key>assets</key>
      <array>
        <dict>
          <key>kind</key>
          <string>software-package</string>
          <key>url</key>
          <string><![CDATA[` + info.AppFileUrl + `]]></string>
        </dict>
        <dict>
          <key>kind</key>
          <string>display-image</string>
          <key>needs-shine</key>
          <integer>0</integer>
          <key>url</key>
          <string><![CDATA[` + info.IconUrl + `]]></string>
        </dict>
        <dict>
          <key>kind</key>
          <string>full-size-image</string>
          <key>needs-shine</key>
          <true/>
          <key>url</key>
          <string><![CDATA[` + info.IconUrl + `]]></string>
        </dict>
      </array>
      <key>metadata</key>
      <dict>
        <key>bundle-identifier</key>
        <string>` + info.BundleID + `</string>
        <key>bundle-version</key>
        <string><![CDATA[` + info.Version + `]]></string>
        <key>kind</key>
        <string>software</string>
        <key>title</key>
        <string><![CDATA[` + info.Name + `]]></string>
      </dict>
    </dict>
  </array>
</dict>
</plist>`
}
