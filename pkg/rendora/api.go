package rendora

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type apiRenderArgs struct {
	URI string `json:"uri" binding:"required"`
}

// APIRender provides the http client with HeadlessResponse
func (R *Rendora) APIRender(c *gin.Context) {

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
