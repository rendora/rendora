package main

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func isInSlice(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func hasPrefixinSlice(slice []string, str string) bool {
	for _, s := range slice {
		if strings.HasPrefix(str, s) {
			return true
		}
	}
	return false
}

var allowedWords []string

func isBot(ua string) (bool, bool) {
	muaLower := strings.ToLower(ua)

	for _, s := range allowedWords {
		if strings.Index(muaLower, s) >= 0 {
			return true, strings.Index(muaLower, "mobile") >= 0
		}
	}
	return false, false
}

//IsWhitelisted checks whether the current request is whitelisted (i.e. should be SSR'ed) or not
func IsWhitelisted(c *gin.Context) bool {
	mua := c.Request.Header.Get("User-Agent")
	filters := &Rendora.C.Filters

	switch filters.UserAgent.Default {
	case "whitelist":
		isbot, _ := isBot(mua)
		if filters.Preset == "bots" && isbot == false {
			return false
		}
		if isInSlice(filters.UserAgent.Exceptions, mua) {
			return false
		}
		break
	case "blacklist":
		if isInSlice(filters.UserAgent.Exceptions, mua) == false {
			return false
		}

	}

	uri := c.Request.RequestURI

	if len(filters.Paths.Static.Exact) > 0 && isInSlice(filters.Paths.Static.Exact, uri) {
		return false
	}

	if len(filters.Paths.Static.Prefix) > 0 && hasPrefixinSlice(filters.Paths.Static.Prefix, uri) {
		return false
	}

	switch filters.Paths.Default {
	case "blacklist":
		if len(filters.Paths.Exceptions.Exact) > 0 && isInSlice(filters.Paths.Exceptions.Exact, uri) {
			return true
		}

		if len(filters.Paths.Exceptions.Prefix) > 0 && hasPrefixinSlice(filters.Paths.Exceptions.Prefix, uri) {
			return true
		}
		return false
	case "whitelist":
		if len(filters.Paths.Exceptions.Exact) > 0 && isInSlice(filters.Paths.Exceptions.Exact, uri) {
			return false
		}

		if len(filters.Paths.Exceptions.Prefix) > 0 && hasPrefixinSlice(filters.Paths.Exceptions.Prefix, uri) {
			return false
		}
		return true
	default:
		return false
	}

}
