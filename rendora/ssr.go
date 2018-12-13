package main

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type reqBody struct {
	URL string `json:"url"`
}

//HeadlessResponse contains the status code, DOM content and headers of the response coming from the headless chrome instance
type HeadlessResponse struct {
	Status  int               `json:"status"`
	Content string            `json:"content"`
	Headers map[string]string `json:"headers"`
	Latency float64           `json:"latency"`
}

func getHeadlessExternal(uri string) (*HeadlessResponse, error) {
	client := &http.Client{}
	bd := reqBody{
		URL: Rendora.C.Target.URL + uri,
	}

	s, err := json.Marshal(bd)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, Rendora.C.Headless.URL+"/pages", bytes.NewBuffer(s))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Rendora-Auth", Rendora.C.Headless.AuthToken)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintln("Unsuccessful result with code:  %d", resp.StatusCode))
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var ret HeadlessResponse
	err = json.Unmarshal(body, &ret)
	if err != nil {
		return nil, err
	}

	return &ret, nil
}

var targetURL string

func getHeadless(uri string) (*HeadlessResponse, error) {
	switch Rendora.HMode {
	case HeadlessModeExternal:
		return getHeadlessExternal(uri)
	case HeadlessModeInternal:
		fallthrough
	default:
		return Rendora.H.GoTo(Rendora.C.Target.URL + uri)
	}
}

func getResponse(uri string) (*HeadlessResponse, error) {
	cKey := Rendora.C.Cache.Redis.KeyPrefix + ":" + uri
	resp, exists, err := Rendora.Cache.Get(cKey)

	if err != nil {
		log.Println(err)
	}

	if exists {
		return resp, nil
	}

	dt, err := getHeadless(uri)
	if err != nil {
		return nil, err
	}
	if Rendora.C.Output.Minify {
		m := minify.New()
		m.AddFunc("text/html", html.Minify)
		m.AddFunc("text/css", css.Minify)
		dt.Content, err = m.String("text/html", dt.Content)
		if err != nil {
			return nil, err
		}
	}

	defer Rendora.Cache.Set(cKey, dt)
	return dt, nil
}

func getSSR(c *gin.Context) {

	resp, err := getResponse(c.Request.RequestURI)
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	contentHdr, ok := resp.Headers["Content-Type"]
	if ok == false {
		contentHdr = "text/html; charset=utf-8"
	}

	c.Header("Content-Type", contentHdr)
	c.String(resp.Status, resp.Content)

	if Rendora.C.Server.Enable {
		Rendora.M.CountSSR.Inc()
	}
}
