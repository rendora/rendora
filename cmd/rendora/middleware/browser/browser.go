package browser

import (
	"crypto/md5"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/rendora/rendora/utils"

	"github.com/gin-gonic/gin"
)

var (
	templateName = "check"
	enUS         = "en-US"
	zhCN         = "zh-CN"
)

type TemplateData struct {
	Title                                                                                       string
	Suggest                                                                                     string
	Logo                                                                                        string
	GoogleDownloadURL, FirefoxDownloadURL, EdgeDownloadURL, SafariDownloadURL, OperaDownloadURL string
}

func Check(ctx *gin.Context) {
	ctx.Next()

	ext := filepath.Ext(ctx.Request.RequestURI)
	if ext == "" && ctx.Request.Method == http.MethodGet && isOldBrowser(ctx) {
		content := []byte(page)
		etag := fmt.Sprintf("%x", md5.Sum(content))
		ctx.Header("ETag", etag)
		ctx.Header("Cache-Control", "no-cache")

		if match := ctx.GetHeader("If-None-Match"); match != "" {
			if strings.Contains(match, etag) {
				ctx.AbortWithStatus(http.StatusNotModified)
				return
			}
		}

		var data TemplateData
		switch utils.GetLang(ctx) {
		case zhCN:
			data = TemplateData{
				Title:              "双语帮 - 双语好，样样好",
				Suggest:            "您当前使用的浏览器版本过低, 建议下载使用以下浏览器最新版本！",
				Logo:               "https://a.links123.cn/common/imgs/ie/logo_shuangyubang@2x.png",
				GoogleDownloadURL:  "https://www.google.cn/intl/zh-CN/chrome/",
				FirefoxDownloadURL: "https://www.firefox.com.cn/",
				EdgeDownloadURL:    "https://www.microsoft.com/zh-cn/windows/microsoft-edge",
				SafariDownloadURL:  "https://www.apple.com/cn/safari/",
				OperaDownloadURL:   "https://www.opera.com/zh-cn/download",
			}
		case enUS:
			data = TemplateData{
				Title:              "Bilingo - Bilingo means better",
				Suggest:            "You need the updated versions of the following browsers to view the site properly:",
				Logo:               "https://a.links123.cn/common/imgs/logo_en.png",
				GoogleDownloadURL:  "https://www.google.cn/intl/en-US/chrome/",
				FirefoxDownloadURL: "https://www.mozilla.org/en-US/firefox/",
				EdgeDownloadURL:    "https://www.microsoft.com/en-us/windows/microsoft-edge",
				SafariDownloadURL:  "https://www.apple.com/safari/",
				OperaDownloadURL:   "https://www.opera.com/",
			}
		default:
			data = TemplateData{
				Title:              "Bilingo - Bilingo means better",
				Suggest:            "You need the updated versions of the following browsers to view the site properly:",
				Logo:               "https://a.links123.cn/common/imgs/logo_en.png",
				GoogleDownloadURL:  "https://www.google.cn/intl/en-US/chrome/",
				FirefoxDownloadURL: "https://www.mozilla.org/en-US/firefox/",
				EdgeDownloadURL:    "https://www.microsoft.com/en-us/windows/microsoft-edge",
				SafariDownloadURL:  "https://www.apple.com/safari/",
				OperaDownloadURL:   "https://www.opera.com/",
			}
		}

		t := template.Must(template.New(templateName).Parse(page))
		err := t.Execute(ctx.Writer, data)
		if err != nil {
			ctx.AbortWithStatus(http.StatusInternalServerError)
		}

		ctx.AbortWithStatus(http.StatusOK)
	}
}

// isOldBrowser check old browser
func isOldBrowser(ctx *gin.Context) bool {
	userAgent := ctx.Request.UserAgent()
	isIE11 := strings.Contains(userAgent, "Trident") && strings.Contains(userAgent, "rv:11.0")
	if isIE11 {
		return true
	}

	isIE := strings.Contains(userAgent, "compatible") && strings.Contains(userAgent, "MSIE")
	if isIE {
		return true
	}

	return false
}
