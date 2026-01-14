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

export SYNAPSE_VERSION="1.16.0"
export SYNAPSE_BUILD="476"
export SYNAPSE_DIST="ubuntu22.04"

export DOCKER_IMG_NAME="rhlml"
export DOCKER_IMG_TAG="0.0.1"

export BASE_INSTALLER_IMAGE="artifactory-kfs.habana-labs.com/docker-local/${SYNAPSE_VERSION}/${SYNAPSE_DIST}/habanalabs/base-installer:${SYNAPSE_VERSION}-${SYNAPSE_BUILD}"

DOCKER_BUILDKIT=1 docker build . \
	-t ${DOCKER_IMG_NAME}:${DOCKER_IMG_TAG} \
	--build-arg BASE_INSTALLER_IMAGE=${BASE_INSTALLER_IMAGE}
