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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

func (R *Rendora) getProxy(c *gin.Context) {
	director := func(req *http.Request) {
		req.Host = R.backendURL.Host
		req.URL.Scheme = R.backendURL.Scheme
		req.URL.Host = R.backendURL.Host
		req.RequestURI = c.Request.RequestURI
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func (R *Rendora) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			R.getProxy(c)
			return
		}

		if c.Request.Header.Get("X-Rendora-Type") == "RENDER" {
			R.getProxy(c)
			return
		}

		if R.isWhitelisted(c) {
			R.getSSR(c)
		} else {
			R.getProxy(c)
		}

		if R.c.Server.Enable {
			R.metrics.CountTotal.Inc()
		}
	}
}

func (R *Rendora) initProxyServer() *http.Server {
	r := gin.New()
	r.Use(R.middleware())

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", R.c.Listen.Address, R.c.Listen.Port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv
}

func (R *Rendora) initRendoraServer() *http.Server {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if R.c.Server.Auth.Enable {
			if c.Request.Header.Get(R.c.Server.Auth.Name) != R.c.Server.Auth.Value {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "wrong authentication key",
				})
			}
		}

	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/render", R.apiRender)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", R.c.Server.Listen.Address, R.c.Server.Listen.Port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv
}

var (
	g errgroup.Group
)

//Run starts Rendora proxy nd API (if enabled) servers
func (R *Rendora) Run() error {

	if R.c.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	g.Go(func() error {
		return R.initProxyServer().ListenAndServe()
	})

	if R.c.Server.Enable {
		g.Go(func() error {
			return R.initRendoraServer().ListenAndServe()
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}
	return nil
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
