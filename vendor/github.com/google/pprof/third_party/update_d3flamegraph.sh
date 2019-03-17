#  Copyright 2018 Google Inc. All Rights Reserved.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

#!/usr/bin/env bash

set -eu
set -o pipefail

D3FLAMEGRAPH_REPO="https://raw.githubusercontent.com/spiermar/d3-flame-graph"
D3FLAMEGRAPH_VERSION="2.0.0-alpha4"
D3FLAMEGRAPH_JS="d3-flamegraph.js"
D3FLAMEGRAPH_CSS="d3-flamegraph.css"

cd $(dirname $0)

D3FLAMEGRAPH_DIR=d3flamegraph

generate_d3flamegraph_go() {
    local d3_js=$(curl -s "${D3FLAMEGRAPH_REPO}/${D3FLAMEGRAPH_VERSION}/dist/${D3FLAMEGRAPH_JS}" | sed 's/`/`+"`"+`/g')
    local d3_css=$(curl -s "${D3FLAMEGRAPH_REPO}/${D3FLAMEGRAPH_VERSION}/dist/${D3FLAMEGRAPH_CSS}")

    cat <<-EOF > $D3FLAMEGRAPH_DIR/d3_flame_graph.go
// A D3.js plugin that produces flame graphs from hierarchical data.
// https://github.com/spiermar/d3-flame-graph
// Version $D3FLAMEGRAPH_VERSION
// See LICENSE file for license details

package d3flamegraph

// JSSource returns the $D3FLAMEGRAPH_JS file
const JSSource = \`
$d3_js
\`

// CSSSource returns the $D3FLAMEGRAPH_CSS file
const CSSSource = \`
$d3_css
\`
EOF
    gofmt -w $D3FLAMEGRAPH_DIR/d3_flame_graph.go
}

get_license() {
    curl -s -o $D3FLAMEGRAPH_DIR/LICENSE "${D3FLAMEGRAPH_REPO}/${D3FLAMEGRAPH_VERSION}/LICENSE"
}

mkdir -p $D3FLAMEGRAPH_DIR
get_license
generate_d3flamegraph_go
