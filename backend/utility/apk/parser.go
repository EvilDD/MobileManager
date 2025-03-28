package apk

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strings"

	"github.com/avast/apkparser"
	"github.com/gogf/gf/v2/frame/g"
)

// APKInfo 存储解析出的APK信息
type APKInfo struct {
	PackageName     string
	VersionName     string
	VersionCode     string
	ApplicationName string
}

// ParseAPK 解析APK文件
func ParseAPK(apkPath string) (*APKInfo, error) {
	info := &APKInfo{}

	// 创建一个buffer来存储XML输出
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("", "\t")

	// 解析APK文件
	zipErr, resErr, manErr := apkparser.ParseApk(apkPath, enc)
	if zipErr != nil {
		return nil, fmt.Errorf("打开APK文件失败: %v", zipErr)
	}
	if manErr != nil {
		return nil, fmt.Errorf("解析AndroidManifest.xml失败: %v", manErr)
	}
	if resErr != nil {
		g.Log().Warning(nil, "解析资源文件失败:", resErr)
	}

	// 解析生成的XML
	decoder := xml.NewDecoder(&buf)
	for {
		token, err := decoder.Token()
		if err != nil {
			break
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "manifest" {
				// 解析manifest标签的属性
				for _, attr := range t.Attr {
					switch attr.Name.Local {
					case "package":
						info.PackageName = attr.Value
						g.Log().Debug(nil, "包名:", info.PackageName)
					case "versionName":
						info.VersionName = attr.Value
						g.Log().Debug(nil, "版本名:", info.VersionName)
					case "versionCode":
						info.VersionCode = attr.Value
						g.Log().Debug(nil, "版本号:", info.VersionCode)
					}
				}
			} else if t.Name.Local == "application" {
				// 解析application标签的属性
				for _, attr := range t.Attr {
					if attr.Name.Local == "label" {
						info.ApplicationName = strings.TrimPrefix(attr.Value, "@string/")
						g.Log().Debug(nil, "应用名称:", info.ApplicationName)
						break
					}
				}
			}
		}
	}

	// 如果应用名为空，使用包名
	if info.ApplicationName == "" {
		info.ApplicationName = info.PackageName
	}

	// 如果版本为空，使用默认值
	if info.VersionName == "" {
		info.VersionName = "1.0.0"
	}

	return info, nil
}
