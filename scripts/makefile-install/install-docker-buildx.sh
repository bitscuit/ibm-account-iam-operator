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
LOCAL_OS=$4
LOCAL_ARCH=$5

set -ex

# Download the binary file
curl -s -L -f "https://github.com/docker/buildx/releases/download/$DOCKER_BUILDX_VERSION/buildx-$DOCKER_BUILDX_VERSION.$LOCAL_OS-$LOCAL_ARCH" > "$DOCKER_BUILDX"
chmod a+x "$DOCKER_BUILDX"

# Make a symlink to the local file
ln -sf "$DOCKER_BUILDX" "$DOCKER_CLI_PLUGINS"/docker-buildx

# Install the plugin
docker buildx install