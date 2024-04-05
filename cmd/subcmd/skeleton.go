package subcmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(skeletonCmd)
}

var skeletonCmd = &cobra.Command{
	Use:   "skeleton",
	Short: "Generate config skeleton",
	RunE: func(_ *cobra.Command, _ []string) error {
		fmt.Println(skeletonYAML)
		return nil
	},
}

const skeletonYAML = `---
# Find matching dependencies in the following order:
#
# 1. Ignore directories and files according to 'ignore'.
# 2. Determine the file's category according to 'category'.
# 3. Normalize categories according to 'normalizer.category'.
# 4. Find nodes (dependencies) according to 'node'.
# 5. Normalize nodes according to 'normalizer.node'.
#
# List of regular expressions for files and directories to ignore.
ignore:
  - "ignore"
# Determine file categories from filename or text.
category:
  - name: if filename matches then the categories are bash and sh
    filename:
      regex: "\\.sh$"
      value:
        - "bash"
        - "sh"
  - filename: # 'name' is optional
      # https://pkg.go.dev/regexp#Regexp.ExpandString
      # extract extension as category
      regex: "\\.(?P<ext>\\w+)$"
      template: "$ext"
  - name: if file content matches then the category is bash
    text:
      regex: "#!/bin/bash"
      value:
        - "bash"
# Find dependencies.
node:
  - name: create bash node
    category: "bash"
    selector:
      regex: "^\\. (?P<v>.+)$"
      template: "$v"
  - name: create bin node
    category: ".*"
    selector:
      regex: "/usr/bin/\\w+"
# Normalize categories and nodes.
# If there is no match, the value remains as is.
normalizer:
  category:
    - name: sh to bash
      matcher:
        regex: "^sh$"
        value:
          - bash
  node:
    - name: extract binary name
      matcher:
        regex: "^/usr/bin/(?<v>\\w+)$"
        template: "$v"`
