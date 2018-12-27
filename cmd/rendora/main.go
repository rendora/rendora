package main

import (
	"fmt"
	"log"
	"os"
	"runtime"

	"github.com/rendora/rendora/pkg/rendora"
	"github.com/spf13/cobra"
)

// VERSION prints the git version of Rendora
var VERSION string

func main() {
	cobra.OnInitialize()
	var cfgFile string

	rootCmd := &cobra.Command{
		Use:  "rendora",
		Long: `dynamic server-side rendering using headless Chrome to effortlessly solve the SEO problem for modern javascript websites`,
		Run: func(cmd *cobra.Command, args []string) {
			Rendora, err := rendora.New(cfgFile)
			if err != nil {
				log.Fatal(err)
			}
			Rendora.Run()
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

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
