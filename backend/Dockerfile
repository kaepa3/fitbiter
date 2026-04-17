# backend/Dockerfile
FROM golang:1.26-alpine AS builder
RUN apk add --no-cache gcc musl-dev
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# backend フォルダ内のソースをビルド
RUN CGO_ENABLED=1 GOOS=linux go build -o fitbiter .

FROM alpine:latest
RUN apk add --no-cache ca-certificates sqlite
WORKDIR /app
# ビルド済みバイナリとテンプレートをコピー
COPY --from=builder /app/fitbiter .
COPY --from=builder /app/templates ./templates
RUN chmod +x ./fitbiter

EXPOSE 8080
CMD ["./fitbiter"]
