#!/bin/bash

set -eo pipefail

VERSION=""
if [ -z "${VERSION}" ]; then
  VERSION=$(git tag | sort -r --version-sort | head -n1)
fi

if [ -z "${VERSION}" ]; then
  echo "VERSION can not be empty"
fi

VERSION=${VERSION/v/}

echo "Building version ${VERSION}"

mkdir -p bin

for arch in ${ALL_GOARCH}; do
  for platform in ${ALL_GOOS}; do
    file="bin/${NAME}-${VERSION}.${platform}.${arch}"
    echo "Building ${file}"
    CGO_ENABLED=0 GOOS=${platform} GOARCH=${arch} go build -ldflags="-s -w" -o ${file} ./cmd/${NAME}
    tar czf "${file}.tgz" "${file}"
    rm "${file}"
  done
done

cd bin && shasum -a 256 ${NAME}-${VERSION}.* > ${NAME}-${VERSION}.sha256sums
