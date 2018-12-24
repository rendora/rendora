package main

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func isKeywordInSlice(slice []string, str string) bool {
	for _, s := range slice {
		if strings.Index(str, s) >= 0 {
			return true
		}
	}
	return false
}

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

//IsWhitelisted checks whether the current request is whitelisted (i.e. should be SSR'ed) or not
func IsWhitelisted(c *gin.Context) bool {
	mua := c.Request.Header.Get("User-Agent")
	muaLower := strings.ToLower(mua)
	filters := &Rendora.C.Filters

	lenKeywords := len(filters.UserAgent.Exceptions.Keywords)
	lenExceptions := len(filters.UserAgent.Exceptions.Exact)

	switch filters.UserAgent.Default {
	case "whitelist":

		if lenKeywords > 0 && isKeywordInSlice(filters.UserAgent.Exceptions.Keywords, muaLower) {
			return false
		}
		if lenExceptions > 0 && isInSlice(filters.UserAgent.Exceptions.Exact, mua) {
			return false
		}
		break
	case "blacklist":
		if lenKeywords == 0 && lenExceptions == 0 {
			return false
		}
		if lenKeywords > 0 && isKeywordInSlice(filters.UserAgent.Exceptions.Keywords, muaLower) == false {
			return false
		}

		if lenExceptions > 0 && isInSlice(filters.UserAgent.Exceptions.Exact, mua) == false {
			return false
		}

	}

	uri := c.Request.RequestURI

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
