package subcmd

import (
	"fmt"
	"os"

	"github.com/berquerant/grdep"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reCmd)
}

var reCmd = &cobra.Command{
	Use:   "re REGEXP [TEMPLATE]",
	Short: "Test regexp",
	Long:  `Test regexp, useful to debug matcher.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := getLogger(cmd, os.Stdout)

		switch len(args) {
		case 1:
			r := grdep.NewRegexp(args[0])
			for x := range grdep.ReadLines(cmd.Context(), os.Stdin) {
				if err := x.Err; err != nil {
					return err
				}
				matches := r.Unwrap().FindAllStringSubmatch(x.Text, -1)
				if len(matches) == 0 {
					continue
				}
				logger.Info("matched", "linum", x.Linum, "text", x.Text, "matches", matches)
			}
			return nil
		case 2:
			r := grdep.NewRegexp(args[0])
			template := args[1]
			for x := range grdep.ReadLines(cmd.Context(), os.Stdin) {
				if err := x.Err; err != nil {
					return err
				}
				matches := r.Unwrap().FindAllStringSubmatchIndex(x.Text, -1)
				if len(matches) == 0 {
					continue
				}
				result := []byte{}
				for _, match := range matches {
					result = r.Unwrap().ExpandString(result, template, x.Text, match)
				}
				logger.Info("matched", "linum", x.Linum, "text", x.Text, "matches", string(result))
			}
			return nil
		default:
			return fmt.Errorf("%w: need REGEXP", errInvalidArgument)
		}
	},
}
