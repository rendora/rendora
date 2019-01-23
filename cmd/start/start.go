package start

import (
	"log"

	"github.com/rendora/rendora/service/rendora"
	"github.com/spf13/cobra"
)

// RunCommand start run rendora service
func RunCommand() *cobra.Command {
	var cfgFile string

	cmd := &cobra.Command{
		Use:     "start",
		Short:   "Start run rendora service",
		Aliases: []string{"s"},
		Run: func(cmd *cobra.Command, args []string) {
			rendora, err := rendora.New(cfgFile)
			if err != nil {
				log.Fatal(err)
			}

			err = rendora.Run()
			if err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file")

	return cmd
}
