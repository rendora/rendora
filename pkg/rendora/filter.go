/*
Copyright 2018 George Badawi.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package rendora

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

//isWhitelisted checks whether the current request is whitelisted (i.e. should be SSR'ed) or not
func (R *Rendora) isWhitelisted(c *gin.Context) bool {
	mua := c.Request.Header.Get("User-Agent")
	muaLower := strings.ToLower(mua)
	filters := &R.c.Filters

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
