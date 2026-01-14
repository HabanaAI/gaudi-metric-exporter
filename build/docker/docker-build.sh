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

# Set the environment variables
source ../../scripts/setup-env.sh


local () {
	clean
	# Build using the local version
	rsync -av ../../../habana-metric-exporter ./ --exclude-from ../../../habana-metric-exporter/.dockerignore
		DOCKER_BUILDKIT=1 docker build . \
		-t ${DOCKER_IMG_NAME}:${DOCKER_IMG_TAG} \
		--build-arg BASE_INSTALLER_IMAGE=${BASE_INSTALLER_IMAGE} \
		--build-arg VERSION="${DOCKER_IMG_TAG}" \
		--build-arg BUILD_DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
	clean
}

push () {
	docker push ${DOCKER_IMG_NAME}:${DOCKER_IMG_TAG}
}

clean () {
	rm -rf go-hlml
	rm -rf habana-metric-exporter
}

$1\
