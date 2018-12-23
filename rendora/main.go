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

package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func getProxy(c *gin.Context) {
	director := func(req *http.Request) {
		req.Host = Rendora.BackendURL.Host
		req.URL.Scheme = Rendora.BackendURL.Scheme
		req.URL.Host = Rendora.BackendURL.Host
		req.RequestURI = c.Request.RequestURI
	}
	proxy := &httputil.ReverseProxy{Director: director}
	proxy.ServeHTTP(c.Writer, c.Request)
}

func middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			getProxy(c)
			return
		}

		if c.Request.Header.Get("X-Rendora-Type") == "RENDER" {
			getProxy(c)
			return
		}

		if IsWhitelisted(c) {
			getSSR(c)
		} else {
			getProxy(c)
		}

		if Rendora.C.Server.Enable {
			Rendora.M.CountTotal.Inc()
		}
	}
}

func initProxyServer() *http.Server {
	r := gin.New()
	r.Use(middleware())

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", Rendora.C.Listen.Address, Rendora.C.Listen.Port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv
}

func initRendoraServer() *http.Server {
	r := gin.New()
	r.Use(func(c *gin.Context) {
		if Rendora.C.Server.Auth.Enable {
			if c.Request.Header.Get(Rendora.C.Server.Auth.Name) != Rendora.C.Server.Auth.Value {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "wrong authentication key",
				})
			}
		}

	})
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	r.POST("/render", APIRender)

	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", Rendora.C.Server.Listen.Address, Rendora.C.Server.Listen.Port),
		Handler:      r,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return srv
}

var (
	g errgroup.Group
)

func init() {
	cobra.OnInitialize()
	initCobra()
}

func execMain() {

	if Rendora.C.Debug == false {
		gin.SetMode(gin.ReleaseMode)
	}

	g.Go(func() error {
		return initProxyServer().ListenAndServe()
	})

	if Rendora.C.Server.Enable {
		g.Go(func() error {
			return initRendoraServer().ListenAndServe()
		})
	}

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
