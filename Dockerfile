# Build a static Go binary in one stage.
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /vision-products-api ./cmd/api

# Run only the compiled binary in a small, non-root image.
FROM alpine:3.23
RUN addgroup -S app && adduser -S app -G app && mkdir /data && chown app:app /data
COPY --from=builder /vision-products-api /usr/local/bin/vision-products-api
USER app
EXPOSE 8080
ENV HTTP_ADDRESS=:8080 DATABASE_PATH=/data/products.db
ENTRYPOINT ["vision-products-api"]
