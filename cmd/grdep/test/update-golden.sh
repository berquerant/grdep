#!/bin/bash

set -e

d=$(cd $(dirname $0); pwd)

cd "${d}/../../.."
make

grdep="${d}/../../../dist/grdep"
target="test/target"
cfg="$(mktemp).yml"

cd "${d}/.."
"$grdep" skeleton > "$cfg"
echo "$target" | "$grdep" run "$cfg" | sort > "${d}/golden.json"
