# 构建阶段
FROM golang:1.20-alpine AS builder

WORKDIR /app

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o devops-api ./cmd/api

# 运行阶段
FROM alpine:3.17

WORKDIR /app

# 复制构建产物
COPY --from=builder /app/devops-api .
COPY --from=builder /app/configs ./configs
COPY --from=builder /app/api ./api

# 安装CA证书(用于HTTPS)
RUN apk --no-cache add ca-certificates

# 暴露端口
EXPOSE 8080

# 运行应用
ENTRYPOINT ["./devops-api", "--config", "configs/default.yaml"]
