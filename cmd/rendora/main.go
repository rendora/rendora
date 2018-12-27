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
			err = Rendora.Run()

			if err != nil {
				log.Fatal(err)
			}

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
