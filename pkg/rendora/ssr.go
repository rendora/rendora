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

/*
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
		return nil, fmt.Errorf("unsuccessful result with code:  %d", resp.StatusCode)
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

*/

var targetURL string

func (R *Rendora) getHeadless(uri string) (*HeadlessResponse, error) {
	return R.h.getResponse(R.c.Target.URL + uri)
}

func (R *Rendora) getResponse(uri string) (*HeadlessResponse, error) {
	cKey := R.c.Cache.Redis.KeyPrefix + ":" + uri
	resp, exists, err := R.cache.get(cKey)

	if err != nil && R.c.LogsMode != "NONE" {
		log.Println(err)
	}

	if exists {
		return resp, nil
	}

	dt, err := R.getHeadless(uri)
	if err != nil {
		return nil, err
	}
	if R.c.Output.Minify {
		m := minify.New()
		m.AddFunc("text/html", html.Minify)
		m.AddFunc("text/css", css.Minify)
		dt.Content, err = m.String("text/html", dt.Content)
		if err != nil {
			return nil, err
		}
	}

	defer R.cache.set(cKey, dt)
	return dt, nil
}

func (R *Rendora) getSSR(c *gin.Context) {

	resp, err := R.getResponse(c.Request.RequestURI)
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

	if R.c.Server.Enable {
		R.metrics.CountSSR.Inc()
	}

}
