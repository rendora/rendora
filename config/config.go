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

package config

import (
	"github.com/asaskevich/govalidator"
	"github.com/spf13/viper"
)

//RendoraConfig represents the global configuration of Rendora
type RendoraConfig struct {
	Debug        bool   `mapstructure:"debug"`
	Node         string `mapstructure:"node" valid:"in(static|render|all)"`
	StaticConfig struct {
		StaticDir string `mapstructure:"staticDir"`
		Listen    struct {
			Address string `valid:"ip"`
			Port    uint16 `valid:"range(1|65535)"`
		}
		Proxy struct {
			Node    string `mapstructure:"node" valid:"in(rendora|rendertron)"`
			Schema  string
			Address string
			Port    uint16 `valid:"range(1|65535)"`
		}
	} `mapstructure:"staticConfig"`

	HeadlessMode string `mapstructure:"headlessMode" valid:"in(default|internal|external)"`
	Target       struct {
		URL string `valid:"required,requrl"`
	} `mapstructure:"target"`

	Headless struct {
		UserAgent   string   `mapstructure:"userAgent"`
		Mode        string   `valid:"in(default|internal|external)"`
		URL         string   `valid:"requrl"`
		AuthToken   string   `mapstructure:"authToken"`
		BlockedURLs []string `mapstructure:"blockedURLs"`
		Timeout     int64    `valid:"range(5|60)"`
		Internal    struct {
			URL string `valid:"url"`
		}
		WaitReadyNode string `valid:"required" mapstructure:"waitReadyNode"`
		WaitTimeout   int64  `valid:"required" mapstructure:"waitTimeout"`
	} `mapstructure:"headless"`

	Cache struct {
		Type    string `valid:"in(local|redis|none)"`
		Timeout uint32 `valid:"range(1|4294967295)"`
		Redis   struct {
			Address   string `valid:"url"`
			Password  string
			DB        int    `valid:"range(0|15)"`
			KeyPrefix string `mapstructure:"keyPrefix"`
		} `mapstructure:"redis"`
	} `mapstructure:"cache"`

	Output struct {
		Minify bool
	} `mapstructure:"output"`

	Filters struct {
		// Preset    string `valid:"in(all|bots)"`
		UserAgent struct {
			Default    string `mapstructure:"defaultPolicy" valid:"in(whitelist|blacklist)"`
			Exceptions struct {
				Keywords []string `valid:"lowercase"`
				Exact    []string
			} `mapstructure:"exceptions"`
		} `mapstructure:"userAgent"`
		Paths struct {
			Default string `mapstructure:"defaultPolicy" valid:"in(whitelist|blacklist)"`
			Static  struct {
				Exact  []string
				Prefix []string
			} `mapstructure:"static"`
			Exceptions struct {
				Exact  []string
				Prefix []string
			} `mapstructure:"exceptions"`
		} `mapstructure:"paths"`
	} `mapstructure:"filters"`

	Server struct {
		Metrics bool
		Auth    struct {
			Enable bool
			Name   string
			Value  string
		}
		Listen struct {
			Address string `valid:"ip"`
			Port    uint16 `valid:"range(1|65535)"`
		}
	}
}

// New initializes the application configuration
func New(cfgFile string) (*RendoraConfig, error) {
	if cfgFile == "" {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/rendora")
	} else {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetDefault("debug", false)
	viper.SetDefault("node", "static")
	viper.SetDefault("staticConfig.staticDir", "./static")
	viper.SetDefault("staticConfig.listen.port", 3001)
	viper.SetDefault("staticConfig.listen.address", "0.0.0.0")
	viper.SetDefault("staticConfig.proxy.node", "rendora")
	viper.SetDefault("staticConfig.proxy.schema", "http")
	viper.SetDefault("staticConfig.proxy.port", 9242)
	viper.SetDefault("staticConfig.proxy.address", "127.0.0.1")
	viper.SetDefault("cache.type", "local")
	viper.SetDefault("cache.timeout", 60*60)
	viper.SetDefault("cache.redis.keyprefix", "__:::rendora:")
	viper.SetDefault("cache.redis.password", "")
	viper.SetDefault("cache.redis.db", 0)
	viper.SetDefault("output.minify", false)
	viper.SetDefault("headless.userAgent", "bilingo-ssr")
	viper.SetDefault("headless.mode", "default")
	viper.SetDefault("headless.timeout", 30)
	viper.SetDefault("headless.internal.url", "http://localhost:9222")
	viper.SetDefault("headless.waitReadyNode", "")
	viper.SetDefault("headless.waitTimeout", 2000)
	viper.SetDefault("filters.useragent.defaultPolicy", "blacklist")
	viper.SetDefault("filters.paths.defaultPolicy", "whitelist")
	viper.SetDefault("server.metrics", "false")
	viper.SetDefault("server.listen.address", "0.0.0.0")
	viper.SetDefault("server.listen.port", "9242")
	viper.SetDefault("server.auth.enable", false)
	viper.SetDefault("server.auth.name", "X-Auth-Rendora")
	viper.SetDefault("server.auth.value", "")
	viper.SetDefault("headless.blockedURLs", []string{
		"*.png", "*.jpg", "*.jpeg", "*.webp", "*.gif", "*.css", "*.woff2", "*.svg", "*.woff", "*.ttf", "*.ico",
		"https://www.youtube.com/*", "https://www.google-analytics.com/*",
		"https://fonts.googleapis.com/*",
	})

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	c := &RendoraConfig{}
	err = viper.Unmarshal(c)
	if err != nil {
		return nil, err
	}

	_, err = govalidator.ValidateStruct(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}
