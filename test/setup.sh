#!/usr/bin/env bash
# Copyright 2025 KDex Tech
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


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
