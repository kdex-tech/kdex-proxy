#!/usr/bin/env bash

set -e

rm -rf node_modules kdex-ui.tgz

(
    cd ../../kdex-ui
    npm i
    npm pack
    cp -f *.tgz ../kdex-proxy/test/kdex-ui.tgz
)

npm i

(
    cd node_modules
    npx esbuild **/*.js --allow-overwrite --outdir=. --define:process.env.NODE_ENV=\"production\"
)
