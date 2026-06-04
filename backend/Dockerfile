FROM golang:1.23-alpine AS build
WORKDIR /src
RUN apk add --no-cache build-base
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=build /out/server /app/server
USER nonroot:nonroot
EXPOSE 8080
ENV BITESENSE_ADDR=:8080 \
    BITESENSE_DATA_DIR=/data \
    BITESENSE_UPLOAD_DIR=/data/uploads \
    BITESENSE_DB_PATH=/data/bitesense.db
ENTRYPOINT ["/app/server"]
