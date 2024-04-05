package subcmd

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func init() {
	configCheckCmd.Flags().Bool("json", false, "output as json")
	rootCmd.AddCommand(configCheckCmd)
}

var configCheckCmd = &cobra.Command{
	Use:   "configcheck FILE_OR_TEXT [FILE_OR_TEXT] [--json]",
	Short: "Test configurations",
	RunE: func(cmd *cobra.Command, args []string) error {
		config, err := parseConfigs(args)
		if err != nil {
			return err
		}

		var b []byte
		if asJSON, _ := cmd.Flags().GetBool("json"); asJSON {
			b, err = json.Marshal(config)
		} else {
			b, err = yaml.Marshal(config)
		}
		if err != nil {
			return err
		}
		fmt.Print(string(b))
		return nil
	},
}
