package middleware

import (
	"bytes"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/rendora/rendora/utils"

	"github.com/gin-gonic/gin"
)

var (
	templateName = "index"
	templateData TemplateData
)

type (
	replaceWriter struct {
		gin.ResponseWriter
	}

	TemplateData struct {
		Keywords    string
		Description string
		Title       string
	}
)

func (r *replaceWriter) WriteString(data string) (int, error) {
	var buf bytes.Buffer

	t := template.Must(template.New(templateName).Parse(data))
	err := t.Execute(&buf, templateData)
	if err != nil {
		return r.ResponseWriter.WriteString(data)
	}

	r.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	return r.ResponseWriter.WriteString(buf.String())
}

func (r *replaceWriter) Write(data []byte) (int, error) {
	var buf bytes.Buffer

	t := template.Must(template.New(templateName).Parse(string(data)))
	err := t.Execute(&buf, templateData)
	if err != nil {
		return r.ResponseWriter.Write(data)
	}

	r.ResponseWriter.Header().Set("Content-Length", strconv.Itoa(buf.Len()))

	return r.ResponseWriter.Write(buf.Bytes())
}

func getTemplateData(lang string) (td TemplateData) {
	switch lang {
	case "zh-CN":
		td = TemplateData{
			Keywords:    "外教,直播课,免费课,美国外教,英语课,汉语课,对外汉语,HSK,背单词,自学英语,学英语的网站,外教一对一,英语口语,在线英语,初级英语口语,英语流利说,英文翻译,免费外教",
			Description: "英文有问题？双语帮问答栏当天解答！想练口语？双语帮天天免费外教课！想背单词？小麦大米是单词收割机！看视频？玩游戏？刷题？英文+中文，双语尽在双语帮！",
			Title:       "双语帮 - 双语好，样样好",
		}
	case "en-US":
		td = TemplateData{
			Keywords:    "外教,直播课,免费课,美国外教,英语课,汉语课,对外汉语,HSK,背单词,自学英语,学英语的网站,外教一对一,英语口语,在线英语,初级英语口语,英语流利说,英文翻译,免费外教",
			Description: "Questions? Ask Bilingo! Spoken English? Bilingo has free lessons! Vocabulary? WheatRice is your MiniApp! Videos? Games? Quizzes? Bilingo has it all, English or Chinese!",
			Title:       "Bilingo - Bilingo means better",
		}
	default:
		td = TemplateData{
			Keywords:    "外教,直播课,免费课,美国外教,英语课,汉语课,对外汉语,HSK,背单词,自学英语,学英语的网站,外教一对一,英语口语,在线英语,初级英语口语,英语流利说,英文翻译,免费外教",
			Description: "英文有问题？双语帮问答栏当天解答！想练口语？双语帮天天免费外教课！想背单词？小麦大米是单词收割机！看视频？玩游戏？刷题？英文+中文，双语尽在双语帮！",
			Title:       "双语帮 - 双语好，样样好",
		}
	}

	return td
}

// ReplaceHTML replace index page's data.
func ReplaceHTML() gin.HandlerFunc {
	return func(c *gin.Context) {
		ext := filepath.Ext(c.Request.RequestURI)
		if ext == "" && c.Request.Method == http.MethodGet {
			c.Writer = &replaceWriter{
				ResponseWriter: c.Writer,
			}

			templateData = getTemplateData(utils.GetLang(c))
		}
	}
}
