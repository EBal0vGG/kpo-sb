FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.work ./
COPY gateway/go.mod gateway/go.sum ./gateway/
COPY file-store/go.mod file-store/go.sum ./file-store/
COPY file-analysis/go.mod file-analysis/go.sum ./file-analysis/
COPY pkg/config/go.mod ./pkg/config/go.mod
COPY pkg/httpx/go.mod ./pkg/httpx/go.mod
COPY pkg/logger/go.mod ./pkg/logger/go.mod
COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    cd file-analysis && go build -o /out/server ./cmd/server

FROM gcr.io/distroless/base-debian12
COPY --from=builder /out/server /server
EXPOSE 8082
CMD ["/server"]



