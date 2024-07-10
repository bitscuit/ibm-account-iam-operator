#!/usr/bin/env bash

# Copyright 2024 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

DOCKER_BUILDX=$1
DOCKER_BUILDX_VERSION=$2
DOCKER_CLI_PLUGINS=$3

set -e

# Check if docker buildx works (in case e.g. provided with docker already), if so, always exit with an optional warning
if docker buildx version &> /dev/null; then
    version=$(docker buildx version | awk '{print $2}')
    if [ "$version" = "$DOCKER_BUILDX_VERSION" ]; then
        exit 0
    else
        echo "WARNING! Version mismatch: expected $DOCKER_BUILDX_VERSION, got $version"
        exit 0
    fi
fi

# docker buildx command didn't work, so check plugin exists and is linked to the CLI directory
test -s "$DOCKER_BUILDX"
test -s "$DOCKER_CLI_PLUGINS"/docker-buildx
