# syntax=docker/dockerfile:1
# The MUD server in a container (lesson 42). Three stages: Rust builds the
# browser client, Go builds the server with the web client baked in, and the
# last image has nothing but the server – no shell, no compiler, no root.
#
#   docker build -t mud .
#   docker run -p 4000:4000 -p 8080:8080 -v mud-data:/data mud

# 1. The browser client (lesson 41).
FROM rust:1-slim AS web
RUN rustup target add wasm32-unknown-unknown
WORKDIR /src/client
COPY client/ ./
RUN cargo build --release --target wasm32-unknown-unknown -p mud-web

# 2. The server (go:embed takes web/, mud_web.wasm included).
FROM golang:1.24 AS server
WORKDIR /src/server
COPY server/go.mod server/go.sum ./
RUN go mod download
COPY server/ ./
COPY --from=web /src/client/target/wasm32-unknown-unknown/release/mud_web.wasm web/
# Pure Go (the SQLite driver too), so the program needs no C libraries at all.
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/mudserver . && mkdir /out/data

# 3. The image that runs: "distroless" has no shell and runs as "nonroot".
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/mudserver /mudserver
COPY --from=server --chown=nonroot:nonroot /out/data /data
ENV MUD_ADDR=:4000 \
    MUD_WEB=:8080 \
    MUD_DB=/data/mud.db \
    MUD_LOG=json
VOLUME /data
EXPOSE 4000 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s CMD ["/mudserver", "-healthcheck"]
ENTRYPOINT ["/mudserver"]
