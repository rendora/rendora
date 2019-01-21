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
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

//Rendora contains the main structure instance
type Rendora struct {
	c          *rendoraConfig
	cache      *cacheStore
	backendURL *url.URL
	h          *headlessClient
	metrics    *metrics
	cfgFile    string
}

//New creates a new Rendora instance
func New(cfgFile string) (*Rendora, error) {
	rendora := &Rendora{
		c:       &rendoraConfig{},
		metrics: &metrics{},
		cfgFile: cfgFile,
	}

	err := rendora.initConfig()
	if err != nil {
		return nil, err
	}

	return rendora, nil
}

//Run starts Rendora proxy nd API (if enabled) servers
func (r *Rendora) Run() error {
	if r.c.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	g.Go(func() error {
		return r.initProxyServer().ListenAndServe()
	})

	if r.c.Server.Enable {
		g.Go(func() error {
			return r.initRendoraServer().ListenAndServe()
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (r *Rendora) initProxyServer() *http.Server {
	router := gin.New()
	router.Use(r.middleware())

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", r.c.Listen.Address, r.c.Listen.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv
}

func (r *Rendora) initRendoraServer() *http.Server {
	router := gin.New()
	router.Use(func(c *gin.Context) {
		if r.c.Server.Auth.Enable {
			if c.Request.Header.Get(r.c.Server.Auth.Name) != r.c.Server.Auth.Value {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "wrong authentication key",
				})
			}
		}

	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.POST("/render", r.apiRender)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", r.c.Server.Listen.Address, r.c.Server.Listen.Port),
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv
}

func (r *Rendora) getProxy(c *gin.Context) {
	director := func(req *http.Request) {
		req.Host = r.backendURL.Host
		req.URL.Scheme = r.backendURL.Scheme
		req.URL.Host = r.backendURL.Host
		req.RequestURI = c.Request.RequestURI
	}

	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (r *Rendora) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			r.getProxy(c)
			return
		}

		if c.Request.Header.Get("X-Rendora-Type") == "RENDER" {
			r.getProxy(c)
			return
		}

		if r.isWhitelisted(c) {
			r.getSSR(c)
		} else {
			r.getProxy(c)
		}

		if r.c.Server.Enable {
			r.metrics.CountTotal.Inc()
		}
	}
}
