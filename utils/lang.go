package utils

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// getLang 获取当前语言
func GetLang(ctx *gin.Context) string {
	lang, err := ctx.Cookie("lnk_lang")
	if err != nil {
		lang = ""
	}

	if lang == "" {
		// 如果cookie中没有则从url中获取
		lang = ctx.Query("lang")
		if lang == "" {
			acceptLanguage := ctx.GetHeader("accept-Language")
			if strings.Contains(acceptLanguage, "en") {
				lang = "en-US"
			} else {
				lang = "zh-CN"
			}
		}
	}

	return lang
}
