# 构建阶段
FROM golang:1.21-alpine AS builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache gcc musl-dev

# 复制依赖文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 编译
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o tunehub-music .

# 运行阶段
FROM alpine:latest

WORKDIR /app

# 安装 ca-certificates 用于 HTTPS 请求
RUN apk add --no-cache ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 从构建阶段复制可执行文件
COPY --from=builder /app/tunehub-music .

# 创建数据目录
RUN mkdir -p /app/data /app/downloads

# 暴露端口
EXPOSE 8080

# 数据卷
VOLUME ["/app/data", "/app/downloads"]

# 启动
CMD ["./tunehub-music"]
