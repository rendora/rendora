package rendora

import (
	"log"
	"net/url"

	"github.com/asaskevich/govalidator"
	"github.com/spf13/viper"
)

// HeadlessMode represents headless modes, currently only internal is supported
type HeadlessMode uint16

//HeadlessMode types
const (
	HeadlessModeInternal HeadlessMode = 0
	HeadlessModeExternal HeadlessMode = 1
)

type backend struct {
	URL string
}

//RendoraConfig represents the global configuration of Rendora
type RendoraConfig struct {
	HeadlessMode string `mapstructure:"headlessMode" valid:"in(default|internal|external)"`
	Debug        bool   `mapstructure:"debug"`
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
func (R *Rendora) initConfig(cfgFile string) error {

	if cfgFile == "" {
		viper.SetConfigName("config")
		viper.AddConfigPath(".")
		viper.AddConfigPath("/etc/rendora")
	} else {
		viper.SetConfigFile(cfgFile)
	}

	viper.SetDefault("debug", false)
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

	R.BackendURL, err = url.Parse(R.c.Backend.URL)
	if err != nil {
		return err
	}

	log.Println("Configuration loaded")

	R.H, err = R.NewHeadlessClient()

	if err != nil {
		return err
	}

	log.Println("Connected to headless Chrome")

	switch R.c.Headless.Mode {
	case "external":
		R.HMode = HeadlessModeExternal
	case "internal":
		R.HMode = HeadlessModeInternal
	default:
		R.HMode = HeadlessModeInternal
	}

	if R.c.Server.Enable {
		R.M = initPrometheus()
	}

	return nil

}

type Rendora struct {
	c          *RendoraConfig
	Cache      *CacheStore
	BackendURL *url.URL
	H          *HeadlessClient
	HMode      HeadlessMode
	M          *Metrics
}

// VERSION shows Rendora's version
var VERSION string

//Rendora contains global information, most importantly the configuration
