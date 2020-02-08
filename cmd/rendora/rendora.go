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
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rendora/rendora/cmd/rendora/middleware"
	"github.com/rendora/rendora/cmd/rendora/middleware/browser"
	"github.com/rendora/rendora/config"
	"github.com/rendora/rendora/service"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	timeout      = 60
	defaultIndex = "index.html"
)

const (
	nodeStatic = "static"
	nodeRender = "render"
	nodeAll    = "all"
)

const (
	renderNodeRendora    = "rendora"
	renderNodeRendertron = "rendertron"
)

var (
	g errgroup.Group
)

//Rendora contains the main structure instance
type rendora struct {
	c       *config.RendoraConfig
	cache   *service.Store
	metrics *service.Metrics
	hc      *service.HeadlessClient
	cfgFile string
}

//new creates a new Rendora instance
func new(cfgFile string) (*rendora, error) {
	c, err := config.New(cfgFile)
	if err != nil {
		return nil, err
	}

	log.Println("Configuration loaded")

	rendora := &rendora{
		c:       c,
		metrics: &service.Metrics{},
		cfgFile: cfgFile,
	}

	return rendora, nil
}

//Run starts Rendora and API (if enabled) servers
func (r *rendora) run() error {
	if r.c.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	if r.c.Node == nodeStatic || r.c.Node == nodeAll {
		g.Go(func() error {
			return r.initStaticServer().ListenAndServe()
		})
	}

	if r.c.Node == nodeRender || r.c.Node == nodeAll {
		r.cache = service.InitCacheStore(&service.StoreConfig{
			Type:    r.c.Cache.Type,
			Timeout: r.c.Cache.Timeout,
			Redis: struct {
				Address   string
				Password  string
				DB        int
				KeyPrefix string
			}{
				Address:   r.c.Cache.Redis.Address,
				Password:  r.c.Cache.Redis.Password,
				DB:        r.c.Cache.Redis.DB,
				KeyPrefix: r.c.Cache.Redis.KeyPrefix,
			},
		})

		hc, err := service.NewHeadlessClient(&service.HeadlessConfig{
			UserAgent:     r.c.Headless.UserAgent,
			Mode:          r.c.Headless.Mode,
			URL:           r.c.Headless.URL,
			AuthToken:     r.c.Headless.AuthToken,
			BlockedURLs:   r.c.Headless.BlockedURLs,
			Timeout:       r.c.Headless.Timeout,
			InternalURL:   r.c.Headless.Internal.URL,
			WaitReadyNode: r.c.Headless.WaitReadyNode,
			WaitTimeout:   r.c.Headless.WaitTimeout,
		})

		if err != nil {
			return err
		}

		r.hc = hc

		log.Println("Connected to headless Chrome")

		if r.c.Server.Metrics {
			r.metrics = service.InitPrometheus()
		}

		g.Go(func() error {
			return r.initRendoraServer().ListenAndServe()
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}

func (r *rendora) initStaticServer() *http.Server {
	router := gin.Default()
	router.Use(r.CheckBrowser(), r.middleware(), middleware.ReplaceHTML())
	router.Use(static.Serve("/", static.LocalFile(r.c.StaticConfig.StaticDir, false)))
	router.NoRoute(middleware.Index(strings.Join([]string{r.c.StaticConfig.StaticDir, defaultIndex}, string(os.PathSeparator))))
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", r.c.StaticConfig.Listen.Address, r.c.StaticConfig.Listen.Port),
		Handler:      router,
		ReadTimeout:  timeout * time.Second,
		WriteTimeout: timeout * time.Second,
	}

	return srv
}

func (r *rendora) initRendoraServer() *http.Server {
	router := gin.Default()
	router.Use(func(c *gin.Context) {
		if r.c.Server.Auth.Enable {
			if c.Request.Header.Get(r.c.Server.Auth.Name) != r.c.Server.Auth.Value {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "wrong authentication key",
				})
			}
		}

	})

	router.HEAD("/ping", r.ping)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	router.POST("/render", r.apiRender)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", r.c.Server.Listen.Address, r.c.Server.Listen.Port),
		Handler:      router,
		ReadTimeout:  timeout * time.Second,
		WriteTimeout: timeout * time.Second,
	}

	return srv
}

func (r *rendora) middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			return
		}

		if r.isWhitelisted(c) {
			ext := filepath.Ext(c.Request.RequestURI)
			if ext != "" && ext != filepath.Ext(defaultIndex) {
				return
			}

			agent := c.Request.Header.Get("User-Agent")
			if strings.Contains(strings.ToLower(agent), "mobile") {
				c.Set("mobile", true)
			}

			if r.c.Node == nodeAll {
				r.getSSR(c)
			} else {
				switch r.c.StaticConfig.Proxy.Node {
				case renderNodeRendora:
					r.getSSRFromProxy(c)
				case renderNodeRendertron:
					r.getSSRFromRendertron(c)
				default:
					c.String(http.StatusInternalServerError, "unsupported node type")
				}
			}
		}

		if r.c.Server.Metrics {
			r.metrics.CountTotal.Inc()
		}
	}
}

func (r *rendora) CheckBrowser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !r.isWhitelisted(ctx) {
			browser.Check(ctx)
		}
	}
}

// RunCommand start run rendora service
func RunCommand() *cobra.Command {
	var cfgFile string

	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start run rendora service",
		Aliases: []string{"s"},
		Run: func(cmd *cobra.Command, args []string) {
			rendora, err := new(cfgFile)
			if err != nil {
				log.Fatal(err)
			}

			err = rendora.run()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	return cmd
}
