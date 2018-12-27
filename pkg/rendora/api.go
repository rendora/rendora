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
	"net/http"

	"github.com/gin-gonic/gin"
)

type apiRenderArgs struct {
	URI string `json:"uri" binding:"required"`
}

// APIRender provides the http client with HeadlessResponse
func (R *Rendora) apiRender(c *gin.Context) {

	var args apiRenderArgs
	if err := c.ShouldBindJSON(&args); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := R.getResponse(args.URI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)

	c.Writer.Header()["Content-Type"] = []string{"application/json; charset=utf-8"}

	enc := json.NewEncoder(c.Writer)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(resp); err != nil {
		panic(err)
	}
}
