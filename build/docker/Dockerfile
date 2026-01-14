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

# Base image for dependencies
ARG BASE_INSTALLER_IMAGE
FROM ${BASE_INSTALLER_IMAGE} AS base_image
RUN rm /usr/lib/habanalabs/libtpc_kernels.so && \
    rm /usr/lib/habanalabs/libComplexGuid.so && \
    rm /usr/lib/habanalabs/libTPCFuser.so && \
    rm /usr/lib/habanalabs/libSynapse.so && \
    rm /usr/lib/habanalabs/libtpcOpsValidation.so

################################################################################
# Build rhlml binary
FROM rust:1.90.0-bookworm AS rhlml
RUN rustup component add rustfmt clippy

COPY --from=base_image /usr/lib/habanalabs /usr/lib/habanalabs
COPY --from=base_image /usr/include/habanalabs /usr/include/habanalabs
RUN mkdir -p /usr/lib/habanalabs && echo "/usr/lib/habanalabs/" >> /etc/ld.so.conf.d/habanalabs.conf
RUN ldconfig

COPY habana-metric-exporter /usr/local/go/src/metric-exporter

RUN apt-get clean && apt update && apt install -y --no-install-recommends libclang-dev
WORKDIR /usr/local/go/src/metric-exporter/rhlml
RUN cargo fmt --check --all
RUN cargo clippy --all-targets --all-features -- -D warnings
RUN cargo build --release --bin rhlml && cp target/release/rhlml /bin/rhlml

################################################################################
# Build metric-exporter
FROM golang:1.25.5 AS exporter_builder
RUN apt-get update -y && apt-get install -y --no-install-recommends git

COPY --from=base_image /usr/lib/habanalabs /usr/lib/habanalabs
COPY --from=base_image /usr/include/habanalabs /usr/include/habanalabs
RUN echo "/usr/lib/habanalabs/" >> /etc/ld.so.conf.d/habanalabs.conf
RUN ldconfig

COPY habana-metric-exporter /usr/local/go/src/metric-exporter

WORKDIR /usr/local/go/src/metric-exporter
ENV GCFLAGS="all=-spectre=all -N -l"
ENV ASMFLAGS="all=-spectre=all"
ENV LDFLAGS="all=-s -w"
RUN CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build ${GOFLAGS} -trimpath -mod=readonly -gcflags="${GCFLAGS}" -asmflags="${ASMFLAGS}" -ldflags="${LDFLAGS}" -a -o /bin/metric-exporter
RUN go get github.com/google/go-licenses && \
    go run github.com/google/go-licenses save ./... --save_path=licenses || true

# Verify binary build specs with checksec
ENV CHECKSEC_REF="No RELRO,No Canary found,NX enabled,No PIE,N/A,N/A,No Symbols,N/A,0,0"
RUN apt-get update -y && apt-get --no-install-recommends -y install file && \
    wget -q https://raw.githubusercontent.com/slimm609/checksec/refs/heads/main/checksec.bash -O checksec && \
    chmod +x checksec
RUN ./checksec --file=/bin/metric-exporter --output=csv | grep -q "${CHECKSEC_REF}"

################################################################################
# Build release image
FROM ubuntu:24.04
ARG VERSION=""
ARG BUILD_DATE=""

COPY habana-metric-exporter /usr/local/go/src/metric-exporter
COPY --from=base_image /usr/lib/habanalabs /usr/lib/habanalabs
COPY --from=base_image /usr/include/habanalabs /usr/include/habanalabs
COPY --from=rhlml /bin/rhlml /usr/bin/rhlml
COPY --from=exporter_builder /bin/metric-exporter /bin/metric-exporter
RUN echo "/usr/lib/habanalabs/" >> /etc/ld.so.conf.d/habanalabs.conf
RUN ldconfig

USER 65534
ENTRYPOINT  ["/bin/metric-exporter"]

LABEL org.opencontainers.image.vendor="Intel Corp."
LABEL org.opencontainers.image.title="Habana Gaudi Metric Exporter"
LABEL org.opencontainers.image.description="Prometheus exporter for Habana Gaudi metrics. This container image is intended for demo purposes only and not intended for production use. To receive expanded security maintenance from Canonical on the Ubuntu base layer, you may follow the [how-to](https://documentation.ubuntu.com/pro-client/en/docs/howtoguides/enable_in_dockerfile/) guide to enable Ubuntu Pro in a Dockerfile which will require the image to be rebuilt."
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"
LABEL org.opencontainers.image.source=https://github.com/HabanaAI/gaudi-metric-exporter
LABEL org.opencontainers.image.licenses="GPL-2.0-or-later"
