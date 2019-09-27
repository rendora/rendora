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
	"github.com/rendora/rendora/config"
	"github.com/rendora/rendora/service"
	"github.com/silenceper/pool"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const (
	timeout      = 30
	defaultIndex = "index.html"
)

var (
	g errgroup.Group
)

//Rendora contains the main structure instance
type rendora struct {
	c       *config.RendoraConfig
	cache   *service.Store
	metrics *service.Metrics
	hp      pool.Pool
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

	headlessClientPool, err := service.NewHeadlessClientPool(&service.HeadlessConfig{
		UserAgent:     rendora.c.Headless.UserAgent,
		Mode:          rendora.c.Headless.Mode,
		URL:           rendora.c.Headless.URL,
		AuthToken:     rendora.c.Headless.AuthToken,
		BlockedURLs:   rendora.c.Headless.BlockedURLs,
		Timeout:       rendora.c.Headless.Timeout,
		InternalURL:   rendora.c.Headless.Internal.URL,
		WaitReadyNode: rendora.c.Headless.WaitReadyNode,
	})
	if err != nil {
		return nil, err
	}

	rendora.hp = headlessClientPool
	log.Println("Connected to headless Chrome")

	if rendora.c.Server.Enable {
		rendora.metrics = service.InitPrometheus()
	}

	return rendora, nil
}

//Run starts Rendora and API (if enabled) servers
func (r *rendora) run() error {
	if r.c.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	g.Go(func() error {
		return r.initStaticServer().ListenAndServe()
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

func (r *rendora) initStaticServer() *http.Server {
	router := gin.Default()
	router.Use(r.middleware())
	router.GET("/sse", r.sse)
	router.Use(static.Serve("/", static.LocalFile(r.c.StaticDir, false)))
	router.NoRoute(middleware.Index(strings.Join([]string{r.c.StaticDir, defaultIndex}, string(os.PathSeparator))))
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", r.c.Listen.Address, r.c.Listen.Port),
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

			r.getSSR(c)
		}

		if r.c.Server.Enable {
			r.metrics.CountTotal.Inc()
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
