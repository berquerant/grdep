#!/bin/bash

set -e

d=$(cd $(dirname $0); pwd)

grdep="${d}/../../dist/grdep"
target="test/target"
cfg="$(mktemp).yml"

cd "${d}/.."
"$grdep" skeleton > "$cfg"
echo "$target" | "$grdep" run "$cfg" > "${d}/golden.json"
