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
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/spf13/viper"
)

type backend struct {
	URL string
}

//rendoraConfig represents the global configuration of Rendora
type rendoraConfig struct {
	HeadlessMode string `mapstructure:"headlessMode" valid:"in(default|internal|external)"`
	Debug        bool   `mapstructure:"debug"`
	LogsMode     string `valid:"in(ERROR|INFO|DEBUG|NONE)"`
	Listen       struct {
		Address string `valid:"ip"`
		Port    uint16 `valid:"range(1|65535)"`
	}
	Backend struct {
		URL string `valid:"required,requrl"`
	} `mapstructure:"backend"`

	Target struct {
		URL string `valid:"required,requrl"`
	} `mapstructure:"target"`

	Headless struct {
		Mode        string   `valid:"in(default|internal|external)"`
		URL         string   `valid:"requrl"`
		AuthToken   string   `mapstructure:"authToken"`
		BlockedURLs []string `mapstructure:"blockedURLs"`
		Timeout     uint16   `valid:"range(5|30)"`
		Internal    struct {
			URL string `valid:"url"`
		}
		CacheDisabled bool `mapstructure:"cacheDisabled"`

		WaitAfterDOMLoad uint16 `mapstructure:"waitAfterDOMLoad" valid:"range(0|5000)"`
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
		Enable bool
		Auth   struct {
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

// InitConfig initializes the application configuration
func (R *Rendora) initConfig() error {

	if R.cfgFile == "" {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/rendora")
	} else {
		viper.SetConfigFile(R.cfgFile)
	}

	viper.SetDefault("debug", false)
	viper.SetDefault("logsMode", "NONE")
	viper.SetDefault("listen.port", 3001)
	viper.SetDefault("listen.address", "0.0.0.0")
	viper.SetDefault("cache.type", "local")
	viper.SetDefault("cache.timeout", 60*60)
	viper.SetDefault("cache.redis.keyprefix", "__:::rendora:")
	viper.SetDefault("cache.redis.password", "")
	viper.SetDefault("cache.redis.db", 0)
	viper.SetDefault("output.minify", false)
	viper.SetDefault("headless.mode", "default")
	viper.SetDefault("headless.waitAfterDOMLoad", 0)
	viper.SetDefault("headless.timeout", 15)
	viper.SetDefault("headless.internal.url", "http://localhost:9222")
	viper.SetDefault("headless.cacheDisabled", false)
	viper.SetDefault("filters.useragent.defaultPolicy", "blacklist")
	viper.SetDefault("filters.paths.defaultPolicy", "whitelist")
	viper.SetDefault("server.enable", "false")
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
		return err
	}

	err = viper.Unmarshal(R.c)

	if err != nil {
		return err
	}

	_, err = govalidator.ValidateStruct(R.c)
	if err != nil {
		return err
	}

	R.initCacheStore()

	defaultBlockedURLs = R.c.Headless.BlockedURLs

	R.backendURL, err = url.Parse(R.c.Backend.URL)
	if err != nil {
		return err
	}

	logsMode := viper.Get("logsMode")
	if logsMode != "NONE" {
		log.Println("Configuration loaded")
	}

	err = R.newHeadlessClient()

	if err != nil {
		return err
	}

	if logsMode != "NONE" {
		log.Println("Connected to headless Chrome")
	}

	if R.c.Server.Enable {
		R.initPrometheus()
	}

	return nil

}

//Rendora contains the main structure instance
type Rendora struct {
	c          *rendoraConfig
	cache      *cacheStore
	backendURL *url.URL
	h          *headlessClient
	metrics    *metrics
	cfgFile    string
}
