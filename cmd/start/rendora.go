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

package start

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/rendora/rendora/config"
	"github.com/rendora/rendora/service"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/sync/errgroup"
)

var (
	g errgroup.Group
)

//Rendora contains the main structure instance
type Rendora struct {
	c          *config.RendoraConfig
	cache      *service.Store
	backendURL *url.URL
	h          *service.HeadlessClient
	metrics    *service.Metrics
	cfgFile    string
}

//New creates a new Rendora instance
func New(cfgFile string) (*Rendora, error) {
	c, err := config.New(cfgFile)
	if err != nil {
		return nil, err
	}

	log.Println("Configuration loaded")

	rendora := &Rendora{
		c:       c,
		metrics: &service.Metrics{},
		cfgFile: cfgFile,
	}

	rendora.cache = service.InitCacheStore(&service.StoreConfig{
		Type:    rendora.c.Cache.Type,
		Timeout: rendora.c.Cache.Timeout,
		Redis: struct {
			Address   string
			Password  string
			DB        int
			KeyPrefix string
		}{
			Address:   rendora.c.Cache.Redis.Address,
			Password:  rendora.c.Cache.Redis.Password,
			DB:        rendora.c.Cache.Redis.DB,
			KeyPrefix: rendora.c.Cache.Redis.KeyPrefix,
		},
	})

	rendora.backendURL, err = url.Parse(rendora.c.Backend.URL)
	if err != nil {
		return nil, err
	}

	headlessClient, err := service.NewHeadlessClient(&service.HeadlessConfig{
		Mode:             rendora.c.Headless.Mode,
		URL:              rendora.c.Headless.URL,
		AuthToken:        rendora.c.Headless.AuthToken,
		BlockedURLs:      rendora.c.Headless.BlockedURLs,
		Timeout:          rendora.c.Headless.Timeout,
		InternalURL:      rendora.c.Headless.Internal.URL,
		WaitAfterDOMLoad: rendora.c.Headless.WaitAfterDOMLoad,
	})
	if err != nil {
		return nil, err
	}

	rendora.h = headlessClient
	log.Println("Connected to headless Chrome")

	if rendora.c.Server.Enable {
		rendora.metrics = service.InitPrometheus()
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
