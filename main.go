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

	"github.com/rendora/rendora/cmd/rendora"
	"github.com/rendora/rendora/cmd/version"

	"github.com/spf13/cobra"
)

var gitVersion string

func main() {
	rootCmd := &cobra.Command{
		Use:  "rendora",
		Long: "dynamic server-side rendering using headless Chrome to effortlessly solve the SEO problem for modern javascript websites",
	}

	rootCmd.AddCommand(rendora.RunCommand())
	rootCmd.AddCommand(version.RunCommand(gitVersion))

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err.Error())
	}
}
