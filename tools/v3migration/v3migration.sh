#!/usr/bin/env bash

set -eu

# this scripts is used when migrating v2 to v3.
# usage: cd ${GOPATH}/src/github.com/shirou/gopsutil && bash tools/v3migration/v3migration.sh



DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
ROOT=$(cd ${DIR}/../.. && pwd)


## 1. refresh
cd ${ROOT}

/bin/rm -rf v3

## 2. copy directories
# docker is removed, #464 will be fixed
mkdir -p v3
cp -rp cpu disk docker host internal load mem net process winservices v3

# build migartion tool
go build -o v3/v3migration ${DIR}/v3migration.go


V3DIR=$(cd ${ROOT}/v3 && pwd)
cd ${V3DIR}

## 3. mod
go mod init

###  change import path
find . -name "*.go" | xargs -I@ sed -i 's|"github.com/shirou/gopsutil/|"github.com/shirou/gopsutil/v3/|g' @

############ Issues

# #429 process.NetIOCounters is pointless on Linux
./v3migration `pwd` 429


# #464 CgroupMem : fix typo and wrong file names
sed -i 's|memoryLimitInBbytes|memoryLimitInBytes|g' docker/docker.go
sed -i 's|memoryLimitInBbytes|memory.limit_in_bytes|g' docker/docker_linux.go
sed -i 's|memoryFailcnt|"memory.failcnt|g' docker/docker_linux.go


# fix #346
sed -i 's/Soft     int32/Soft     uint64/' process/process.go
sed -i 's/Hard     int32/Hard     uint64/' process/process.go
sed -i 's| //TODO too small. needs to be uint64||' process/process.go
sed -i 's|limitToInt(val string) (int32, error)|limitToUint(val string) (uint64, error)|' process/process_*.go
sed -i 's|limitToInt|limitToUint|' process/process_*.go
sed -i 's|return int32(res), nil|return uint64(res), nil|' process/process_*.go
sed -i 's|math.MaxInt32|math.MaxUint64|' process/process_*.go



############ SHOULD BE FIXED BY HAND



