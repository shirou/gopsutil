# SPDX-License-Identifier: BSD-3-Clause
FROM --platform=linux/amd64 debian:trixie@sha256:9c6e381421d0f48d1b8133e573f33c185ee16cbf505eda9c4a8a4791b25ea8a6

ARG GO_VERSION=1.24.4-1
ARG GCC_VERSION=14.2.0-19
ARG GCC_CROSS_VERSION=14.2.0-19cross1
ARG GCC_MIPS_VERSION=14.2.0-13cross1
ARG LIBC_VERSION=2.41-12+deb13u4
ARG LIBC_CROSS_VERSION=2.41-11cross1
ARG LIBC_MIPS_VERSION=2.40-5cross1

RUN apt-get update && apt-get install -y --no-install-recommends \
    golang-go=2:1.24~2 \
    golang-1.24-go=${GO_VERSION} golang-1.24-src=${GO_VERSION} \
    gcc-14=${GCC_VERSION} \
    gcc-14-i686-linux-gnu=${GCC_CROSS_VERSION} \
    gcc-14-arm-linux-gnueabihf=${GCC_CROSS_VERSION} \
    gcc-14-aarch64-linux-gnu=${GCC_CROSS_VERSION} \
    gcc-14-loongarch64-linux-gnu=${GCC_CROSS_VERSION} \
    gcc-14-powerpc64-linux-gnu=${GCC_CROSS_VERSION} \
    gcc-14-powerpc64le-linux-gnu=${GCC_CROSS_VERSION} \
    gcc-14-riscv64-linux-gnu=${GCC_CROSS_VERSION} \
    gcc-14-s390x-linux-gnu=${GCC_CROSS_VERSION} \
    gcc-14-mips-linux-gnu=${GCC_MIPS_VERSION} \
    gcc-14-mipsel-linux-gnu=${GCC_MIPS_VERSION} \
    gcc-14-mips64-linux-gnuabi64=${GCC_MIPS_VERSION} \
    gcc-14-mips64el-linux-gnuabi64=${GCC_MIPS_VERSION} \
    libc6-dev=${LIBC_VERSION} \
    libc6-dev-i386-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-armhf-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-arm64-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-loong64-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-ppc64-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-ppc64el-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-riscv64-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-s390x-cross=${LIBC_CROSS_VERSION} \
    libc6-dev-mips-cross=${LIBC_MIPS_VERSION} \
    libc6-dev-mipsel-cross=${LIBC_MIPS_VERSION} \
    libc6-dev-mips64-cross=${LIBC_MIPS_VERSION} \
    libc6-dev-mips64el-cross=${LIBC_MIPS_VERSION} \
    && rm -rf /var/lib/apt/lists/*

# Keep compiler version selection here, not in the generation script.
RUN for cc in /usr/bin/*-gcc-14; do ln -sf "$cc" "${cc%-14}"; done

WORKDIR /src/host
ENTRYPOINT ["bash", "linux/generate.sh"]
