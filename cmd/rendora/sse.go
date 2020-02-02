package rendora

import (
	"io"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	versionFileName = "version.json"
	versionCommand  = "version"
)

func (r *rendora) sse(ctx *gin.Context) {
	ctx.Header("Access-Control-Allow-Origin", "*")

	chanStream := make(chan string, 1)
	go func() {
		file, err := ioutil.ReadFile(strings.Join([]string{r.c.StaticConfig.StaticDir, versionFileName}, string(os.PathSeparator)))
		if err != nil {
			return
		}

		chanStream <- string(file)
	}()

	ctx.Stream(func(w io.Writer) bool {
		ctx.SSEvent(versionCommand, <-chanStream)
		return true
	})
}
