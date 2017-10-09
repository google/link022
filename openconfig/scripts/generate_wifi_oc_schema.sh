# Copyright 2017 Google Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

#!/bin/bash
#
# Generates the latest version of the OpenConfig textproto schema descriptor and
# model visualization from the source YANG modules.
# To use this, run thr scripts directory:
#    ./generate_wifi_oc_schema.sh

export GOPATH=$HOME/go

# Tools
YANG_CONVERTER=$GOPATH/src/github.com/openconfig/ygot/generator/generator.go

# OpenConfig modules
OC_DIR=$(pwd)../../../../../../../openconfig
YANG_MODELS=$OC_DIR/public/release/models
IETF_MODELS=$OC_DIR/yang/standard/ietf/RFC
WIFI_MODELS=$OC_DIR/wifi
OWCA_TOP_MODULE=../models/wifi-office.yang
IGNORED_MODULES=openconfig-wifi-phy,openconfig-wifi-mac

# Output path
OUTPUT_PACKAGE_NAME=ocstruct
OUTPUT_FILE_PATH=../../generated/$OUTPUT_PACKAGE_NAME/$OUTPUT_PACKAGE_NAME.go

go run $YANG_CONVERTER \
-path=$YANG_MODELS,$IETF_MODELS,$WIFI_MODELS \
-generate_fakeroot -fakeroot_name=office \
-package_name=ocstruct -compress_paths=false \
-exclude_modules=$IGNORED_MODULES \
-output_file=$OUTPUT_FILE_PATH \
$OWCA_TOP_MODULE
