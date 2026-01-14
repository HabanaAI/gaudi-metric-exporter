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

# syntax=docker/dockerfile:1

# Comments are provided throughout this file to help you get started.
# If you need more help, visit the Dockerfile reference guide at
# https://docs.docker.com/engine/reference/builder/

# HabanaLabs driver deps
ARG BASE_INSTALLER_IMAGE
FROM ${BASE_INSTALLER_IMAGE} as baseImage

# -- Build Stage 1 to build code in minimal image

# syntax=docker/dockerfile:experimental


################################################################################
# Create a stage for building the application.

ARG RUST_VERSION=1.88.0
ARG APP_NAME=rhlml
FROM rust:1.88.0-bookworm AS build
ARG APP_NAME
WORKDIR /app

COPY --from=baseImage /usr/lib/habanalabs /usr/lib/habanalabs
COPY --from=baseImage /usr/include/habanalabs /usr/include/habanalabs
RUN mkdir -p /usr/lib/habanalabs && echo "/usr/lib/habanalabs/" >> /etc/ld.so.conf.d/habanalabs.conf
RUN ldconfig
COPY . .
RUN apt-get clean && apt update && apt install -y --no-install-recommends libclang-dev 
RUN ls /usr/lib/habanalabs
RUN cargo build --release --bin rhlml && cp target/release/rhlml /bin/rhlml

################################################################################
# Create a new stage for running the application that contains the minimal
# runtime dependencies for the application. This often uses a different base
# image from the build stage where the necessary files are copied from the build
# stage.
#
# The example below uses the debian bullseye image as the foundation for    running the app.
# By specifying the "bullseye-slim" tag, it will also use whatever happens to    be the
# most recent version of that tag when you build your Dockerfile. If
# reproducability is important, consider using a digest
# (e.g.,    debian@sha256:ac707220fbd7b67fc19b112cee8170b41a9e97f703f588b2cdbbcdcecdd8af57).
FROM debian:bookworm-slim AS final

COPY --from=baseImage /usr/lib/habanalabs /usr/lib/habanalabs
COPY --from=baseImage /usr/include/habanalabs /usr/include/habanalabs
RUN mkdir -p /usr/lib/habanalabs && echo "/usr/lib/habanalabs/" >> /etc/ld.so.conf.d/habanalabs.conf
RUN ldconfig

# Create a non-privileged user that the app will run under.
# See https://docs.docker.com/develop/develop-images/dockerfile_best-practices/   #user
ARG UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    appuser
USER appuser

# Copy the executable from the "build" stage.
COPY --from=build /bin/rhlml /bin/

# What the container should run when it is started.
CMD ["/bin/bash", "-c", "sleep 1d"]
