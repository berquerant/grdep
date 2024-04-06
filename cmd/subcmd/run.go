package subcmd

import (
	"os"

	"github.com/berquerant/grdep"
	"github.com/spf13/cobra"
)

func init() {
	runCmd.Flags().BoolP("category", "C", false, "Determine category and exit")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run FILE_OR_TEXT [FILE_OR_TEXT]",
	Short: "Find dependencies",
	Long:  "Find dependencies. Treat each line of standard input as a path to search.",
	RunE: func(cmd *cobra.Command, args []string) error {
		logger := getLogger(cmd, os.Stderr)
		config, err := parseConfigs(args)
		if err != nil {
			return err
		}
		var (
			categories          = newNamedCategorySelectors(config.Categories)
			nodes               = newNamedNodeSelectors(config.Nodes)
			categoryNormalizers = newNamedNormalizers(config.Normalizers.Categories)
			nodeNormalizers     = newNamedNormalizers(config.Normalizers.Nodes)
			isDebug             = getDebug(cmd)
			categoryOnly, _     = cmd.Flags().GetBool("category")
		)

		r := runner{
			config:     config,
			r:          os.Stdin,
			w:          os.Stdout,
			logger:     logger,
			isDebug:    isDebug,
			categories: grdep.CachedFunc(categories.Select),
			// Caching lines as keys is not very effective
			nodes:              nodes.Select,
			categoryNormalizer: grdep.CachedFunc(categoryNormalizers.Normalize),
			nodeNormalizer:     grdep.CachedFunc(nodeNormalizers.Normalize),
			categoryOnly:       categoryOnly,
		}
		return r.run(cmd.Context())
	},
}

func newCategorySelector(selector grdep.CSelector) grdep.CategorySelectorIface {
	if selector.Filename != nil {
		return grdep.NewFileCategorySelector(selector.Filename)
	}
	return grdep.NewTextCategorySelector(
		grdep.NewReaderCategorySelector(selector.Text),
	)
}

func newNamedCategorySelectors(categories []grdep.CSelector) grdep.NamedCategorySelectors {
	selectors := make([]*grdep.NamedCategorySelector, len(categories))
	for i, x := range categories {
		selectors[i] = grdep.NewNamedCategorySelector(x.Name, newCategorySelector(x))
	}
	return grdep.NamedCategorySelectors(selectors)
}

func newNamedNodeSelectors(nodes []grdep.NSelector) grdep.NamedNodeSelectors {
	selectors := make([]*grdep.NamedNodeSelector, len(nodes))
	for i, x := range nodes {
		selectors[i] = grdep.NewNamedNodeSelector(x.Name, grdep.NewNodeSelector(x.Category, x.Selector))
	}
	return grdep.NamedNodeSelectors(selectors)
}

func newNamedNormalizers(matchers []grdep.NamedMatcher) grdep.NamedNormalizers {
	return grdep.NamedNormalizers(matchers)
}
