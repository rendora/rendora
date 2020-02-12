package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// RunCommand prints the git version of Rendora
func RunCommand(gitVersion string) *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print the version number of rendora",
		Aliases: []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Rendora Version: ", gitVersion)
			fmt.Println("Go Version: ", runtime.Version())
		},
	}
}
