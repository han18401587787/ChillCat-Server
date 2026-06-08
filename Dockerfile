# 构建阶段
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=sum.golang.google.cn
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/seed ./cmd/seed

# 运行阶段
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/server .
COPY --from=builder /app/seed .
COPY --from=builder /app/configs ./configs

EXPOSE 8080

CMD ["./server"]
