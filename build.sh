#!/bin/sh

cd frontend
npm i
npm run build

cd ..
echo "Backend"

mkdir -p web/html
rm -fr web/html/*
cp -R frontend/dist/* web/html/

BUILD_TAGS="with_quic,with_grpc,with_utls,with_acme,with_gvisor,with_naive_outbound,badlinkname,tfogo_checklinkname0,with_tailscale"

if command -v clang >/dev/null 2>&1; then
  export CC="${CC:-clang}"
  export CXX="${CXX:-clang++}"
  go build -ldflags '-w -s -checklinkname=0 -extldflags "-fuse-ld=lld"' -tags "$BUILD_TAGS" -o sui main.go
else
  echo "clang is required to link sing-box v1.14 cronet archives on Linux." >&2
  echo "Install clang/lld first, then rerun ./build.sh." >&2
  exit 1
fi
