package main

import (
	"fmt"
	"log"
	"net/url"
	"runtime"

	"github.com/asaskevich/govalidator"
	"github.com/spf13/cobra"
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

var rootCmd *cobra.Command
var cfgFile string

func initCobra() {

	rootCmd = &cobra.Command{
		Use:  "rendora",
		Long: `dynamic server-side rendering using headless Chrome to effortlessly solve the SEO problem for modern javascript websites`,
		Run: func(cmd *cobra.Command, args []string) {
			InitConfig()
			execMain()
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	var versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the version number of rendora",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Rendora Version: ", VERSION)
			fmt.Println("Go Version: ", runtime.Version())
		},
	}

	rootCmd.AddCommand(versionCmd)

}

// InitConfig initializes the application configuration
func InitConfig() {

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
		log.Fatal("error: " + err.Error())
	}

	err = viper.Unmarshal(Rendora.C)

	if err != nil {
		log.Fatal("error: " + err.Error())
	}

	_, err = govalidator.ValidateStruct(Rendora.C)
	if err != nil {
		log.Fatal("error: " + err.Error())
	}

	Rendora.Cache = InitCacheStore()

	defaultBlockedURLs = Rendora.C.Headless.BlockedURLs

	Rendora.BackendURL, err = url.Parse(Rendora.C.Backend.URL)
	if err != nil {
		log.Fatal("error: " + err.Error())
	}

	log.Println("Configuration loaded")

	Rendora.H, err = NewHeadlessClient()

	if err != nil {
		log.Fatal("error: " + err.Error())
	}

	log.Println("Connected to headless Chrome")

	switch Rendora.C.Headless.Mode {
	case "external":
		Rendora.HMode = HeadlessModeExternal
	case "internal":
		Rendora.HMode = HeadlessModeInternal
	default:
		Rendora.HMode = HeadlessModeInternal
	}

	if Rendora.C.Server.Enable {
		Rendora.M = initPrometheus()
	}

}

type rendora struct {
	C          *RendoraConfig
	Cache      *CacheStore
	BackendURL *url.URL
	H          *HeadlessClient
	HMode      HeadlessMode
	M          *Metrics
}

// VERSION shows Rendora's version
var VERSION string

//Rendora contains global information, most importantly the configuration
var Rendora = rendora{
	C: &RendoraConfig{},
	M: &Metrics{},
}
