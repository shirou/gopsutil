# SPDX-License-Identifier: BSD-3-Clause
FROM --platform=linux/amd64 debian:trixie

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates golang-go gcc \
    gcc-i686-linux-gnu \
    gcc-arm-linux-gnueabihf \
    gcc-aarch64-linux-gnu \
    gcc-loongarch64-linux-gnu \
    gcc-powerpc64-linux-gnu \
    gcc-powerpc64le-linux-gnu \
    gcc-riscv64-linux-gnu \
    gcc-s390x-linux-gnu \
    gcc-mips-linux-gnu \
    gcc-mipsel-linux-gnu \
    gcc-mips64-linux-gnuabi64 \
    gcc-mips64el-linux-gnuabi64 \
    libc6-dev \
    libc6-dev-i386-cross \
    libc6-dev-armhf-cross \
    libc6-dev-arm64-cross \
    libc6-dev-loong64-cross \
    libc6-dev-ppc64-cross \
    libc6-dev-ppc64el-cross \
    libc6-dev-riscv64-cross \
    libc6-dev-s390x-cross \
    libc6-dev-mips-cross \
    libc6-dev-mipsel-cross \
    libc6-dev-mips64-cross \
    libc6-dev-mips64el-cross \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /src/host
ENTRYPOINT ["bash", "linux/generate.sh"]
