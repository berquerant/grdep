# grdep

```
❯ grdep -h
Find dependencies by grep.

Usage:
  grdep [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  configcheck Test configurations
  help        Help about any command
  re          Test regexp
  run         Find dependencies
  skeleton    Generate config skeleton

Flags:
      --debug     Enable debug logs
  -h, --help      help for grdep
      --metrics   Show metrics

Use "grdep [command] --help" for more information about a command.
```

## Usage

```
❯ grdep skeleton | tee skeleton.yml
---
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
#   matcher:
#     - r: "REGEXP"
#
# 'tmpl' holds a template.
# If a line mathes, then replace variables in the 'tmpl' and pass it to the next.
#
#   matcher:
#     - r: "REGEXP"
#       tmpl: "TEMPLATE"
#
# 'val' holds constants.
# Pass the constants to the next.
#
#   matcher:
#     - val:
#         - "VALUE1"
#         - "VALUE2"
#
# If a line matches, then pass the constants to the next.
#
#   matcher:
#     - r: "REGEXP"
#     - val:
#         - "VALUE1"
#         - "VALUE2"
#
# 'not' holds a regexp.
# If a line does not match, then pass it to the next.
#
#   matcher:
#     - not: "REGEXP"
#
# If a line does not match, then pass the constants to the next.
#
#   matcher:
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
#   matcher:
#     - sh: "BASH"
#
# 'g' holds a glob.
# If a line matches, then pass it to the next.
#
#   matcher:
#     - g: "GLOB"
#
# 'lua' holds a lua script.
# 'lua_call' holds an entrypoint.
# LUA_SCRIPT should contain a function named LUA_ENTRYPOINT.
# The function should takes a string as an argument and returns a string.
# If the script is successful and returns something other than whitespaces, pass it to the next.
#
#   matcher:
#     - lua: LUA_SCRIPT
#       lua_call: "LUA_ENTRYPOINT"
#
# 'lua_file' holds a lua script file.
#
#   matcher:
#     - lua_file: "LUA_SCRIPT_FILE"
#       lua_call: "LUA_ENTRYPOINT"
#
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
  - name: docker entrypoint
    category: "dockerfile"
    matcher:
      - g: "ENTRY*"
      - sh: "cut -d '[' -f 2"
      - sh: "cut -d ']' -f 1"
      - lua: |
          function up(src)
            return string.upper(string.gsub(src, "\"", ""))
          end
        lua_call: up
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
          tmpl: "$v"

❯ echo 'some/path' | grdep run skeleton.yml
```
