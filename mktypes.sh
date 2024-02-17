#!/bin/sh

PKGS="cpu disk docker host load mem net process sensors winservices"

GOOS=$(go env GOOS)
GOARCH=$(go env GOARCH)

for PKG in $PKGS
do
        if [ -e "${PKG}/types_${GOOS}.go" ]; then
                (echo "// +build $GOOS"
                echo "// +build $GOARCH"
                go tool cgo -godefs "${PKG}/types_${GOOS}.go") | gofmt > "${PKG}/${PKG}_${GOOS}_${GOARCH}.go"
        fi
done
