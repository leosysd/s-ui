FROM node:22-bookworm-slim AS front-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

FROM golang:1.25-bookworm AS backend-builder
WORKDIR /app
ENV CGO_ENABLED=1
ENV CC=clang
ENV CXX=clang++
ENV BUILD_TAGS="with_quic,with_grpc,with_utls,with_acme,with_gvisor,with_naive_outbound,badlinkname,tfogo_checklinkname0,with_tailscale"

RUN apt-get update && apt-get install -y --no-install-recommends \
    clang \
    lld \
    ca-certificates \
    file \
    git \
    && rm -rf /var/lib/apt/lists/*

COPY . .
COPY --from=front-builder /app/dist/ /app/web/html/

RUN go build -ldflags='-w -s -checklinkname=0 -extldflags "-fuse-ld=lld"' -tags "$BUILD_TAGS" -o sui main.go \
    && file sui

FROM debian:bookworm-slim
ENV TZ=Asia/Shanghai
WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    bash \
    ca-certificates \
    nftables \
    tzdata \
    && rm -rf /var/lib/apt/lists/*

COPY --from=backend-builder /app/sui /app/
COPY entrypoint.sh /app/
ENTRYPOINT [ "./entrypoint.sh" ]
