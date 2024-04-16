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
# The following can be written in matchers:
#
# 'r' holds a regexp.
# If a line matches, then pass it to the next.
#
#   matchers:
#     - r: "REGEXP"
#
# 'tmpl' holds a template.
# If a line mathes, then replace variables in the 'tmpl' and pass it to the next.
#
#   matchers:
#     - r: "REGEXP"
#       tmpl: "TEMPLATE"
#
# 'val' holds constants.
# Pass the constants to the next.
#
#   matchers:
#     - val:
#         - "VALUE1"
#         - "VALUE2"
#
# If a line matches, then pass the constants to the next.
#
#   matchers:
#     - r: "REGEXP"
#     - val:
#         - "VALUE1"
#         - "VALUE2"
#
# 'not' holds a regexp.
# If a line does not match, then pass it to the next.
#
#   matchers:
#     - not: "REGEXP"
#
# If a line does not match, then pass the constants to the next.
#
#   matchers:
#     - not: "REGEXP"
#     - val:
#         - "VALUE1"
#         - "VALUE2"
#
# 'sh' holds a script.
# Invoke the shell script (bash).
# If the script is successful and outputs something other than whitespaces from stdout,
# pass it to the next.
#
#   matchers:
#     - sh: "BASH"
#
# 'g' holds a glob.
# If a line matches, then pass it to the next.
#
#   matchers:
#     - g: "GLOB"
#
# List of matchers for files and directories to ignore.
ignore:
  - r: "ignore"
# Determine file categories from filename or text.
category:
  - name: if filename matches then the categories are bash and sh
    filename:
      # matchers can be wrtten here
      - r: "\\.sh$"
      - val:
        - "bash"
        - "sh"
  - filename: # 'name' is optional
      # extract extension as category
      - r: "\\.(?P<ext>\\w+)$"
        tmpl: "$ext"
  - name: if file content matches then the category is bash
    text:
      # matchers can be wrtten here
      - r: "#!/bin/bash"
      - val:
          - "bash"
# Find dependencies.
node:
  - name: create bash node
    category: "bash"
    matcher:
      - r: "^\\. (?P<v>.+)$"
        tmpl: "$v"
  - name: create bin node
    category: ".*"
    matcher:
      - r: "^/usr/bin/"
  - name: local/src but not /usr/local/src
    category: "bash"
    matcher:
      - not: "/usr/local/src"
      - r: "local/src"
  - name: install
    category: "bash"
    matcher:
      - r: "^install (?P<v>.+)$"
        tmpl: "$v"
      - sh: "tr ' ' '\n'"
  - name: docker from
    category: "dockerfile"
    matcher:
      - g: "FROM*"
# Normalize categories and nodes.
# If there is no match, the value remains as is.
normalizer:
  category:
    - name: sh to bash
      matcher:
        - r: "^sh$"
        - val:
            - bash
  node:
    - name: extract binary name
      matcher:
        - r: "/usr/bin/(?P<v>\\w+)"
          tmpl: "$v"`
