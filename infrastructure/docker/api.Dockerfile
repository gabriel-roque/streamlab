FROM golang:1.23-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY internal/ ./internal/
COPY apps/api/ ./apps/api/
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/streamlab-api ./apps/api

FROM alpine:3.20

RUN apk add --no-cache ca-certificates wget ffmpeg \
    && addgroup -S streamlab \
    && adduser -S -G streamlab streamlab \
    && mkdir -p /var/lib/streamlab/media \
    && chown -R streamlab:streamlab /var/lib/streamlab

COPY --from=build /out/streamlab-api /usr/local/bin/streamlab-api

USER streamlab
WORKDIR /var/lib/streamlab
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/streamlab-api"]
