#!/usr/bin/env bash

# Copyright (C) 2025 Intel Corporation

# This program is free software; you can redistribute it and/or modify it
# under the terms of the GNU General Public License version 2 or later, as published
# by the Free Software Foundation.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program; if not, see <http://www.gnu.org/licenses/>.

# SPDX-License-Identifier: GPL-2.0-or-later

set -eu

export SYNAPSE_VERSION="1.16.0"
export SYNAPSE_BUILD="523"
export SYNAPSE_DIST="ubuntu22.04"

export GITHUB_BRANCH="master"
export GITHUB_TOKEN="" # Add your key here
export GOHLML_BRANCH="master"

export DOCKER_IMG_NAME="habana-metric-exporter"
export DOCKER_IMG_TAG="0.6.1"

export GO_VERSION="1.21.0"

export BASE_INSTALLER_IMAGE="artifactory-kfs.habana-labs.com/docker-local/${SYNAPSE_VERSION}/${SYNAPSE_DIST}/habanalabs/base-installer:${SYNAPSE_VERSION}-${SYNAPSE_BUILD}"
