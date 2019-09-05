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
	"time"

	"github.com/rendora/rendora/service"

	"github.com/gin-gonic/gin"
	"github.com/tdewolff/minify/v2"
	"github.com/tdewolff/minify/v2/css"
	"github.com/tdewolff/minify/v2/html"
)

type reqBody struct {
	URL string `json:"url"`
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

func (r *rendora) getHeadless(uri string) (*service.HeadlessResponse, error) {
	timeStart := time.Now()
	elapsed := float64(time.Since(timeStart)) / float64(time.Duration(1*time.Millisecond))

	if r.c.Server.Enable {
		r.metrics.Duration.Observe(elapsed)
	}

	headlessResponse, err := r.h.GetResponse(r.c.Target.URL + uri)
	if err != nil {
		return nil, err
	}

	headlessResponse.Latency = elapsed

	return headlessResponse, nil
}

func (r *rendora) getResponse(uri string) (*service.HeadlessResponse, error) {
	cKey := r.c.Cache.Redis.KeyPrefix + ":" + uri
	resp, exists, err := r.cache.Get(cKey)
	if err != nil {
		log.Println(err)
	}

	if exists {
		if r.c.Server.Enable {
			r.metrics.CountSSRCached.Inc()
		}

		return resp, nil
	}

	dt, err := r.getHeadless(uri)
	if err != nil {
		return nil, err
	}
	if r.c.Output.Minify {
		m := minify.New()
		m.AddFunc("text/html", html.Minify)
		m.AddFunc("text/css", css.Minify)
		dt.Content, err = m.String("text/html", dt.Content)
		if err != nil {
			return nil, err
		}
	}

	defer r.cache.Set(cKey, dt)
	return dt, nil
}

func (r *rendora) getSSR(c *gin.Context) {
	resp, err := r.getResponse(c.Request.RequestURI)
	if err != nil {
		c.AbortWithStatus(http.StatusServiceUnavailable)
		return
	}

	contentHdr, ok := resp.Headers["Content-Type"]
	if ok == false {
		contentHdr = "text/html; charset=utf-8"
	}

	c.Header("Content-Type", contentHdr.(string))
	c.String(int(resp.Status), resp.Content)

	if r.c.Server.Enable {
		r.metrics.CountSSR.Inc()
	}
}
