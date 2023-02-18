# 本地编译镜像时，可以先执行 go mod vendor，从而避免编译中的 go mod download 消耗过多时间。

FROM golang:1.18 as builder

# Golang 镜像加速
ENV CGO_ENABLED=0
ENV GOPRIVATE=""
ENV GOPROXY="https://goproxy.cn,direct"
ENV GOSUMDB="sum.golang.google.cn"

WORKDIR /build/

ADD ../.. .
RUN [ ! -d "vendor" ] && go mod download all || echo "go mod download skipped..."
RUN go build -o dashboard cmd/dashboard/main.go

FROM alpine

# Alpine 系统镜像加速
RUN sed -e 's/dl-cdn[.]alpinelinux.org/mirrors.aliyun.com/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache busybox-extras

# 使用国内时区
ENV TZ Asia/Shanghai
RUN apk add tzdata alpine-conf && \
    /sbin/setup-timezone -z Asia/Shanghai && \
    apk del alpine-conf

WORKDIR /root/

COPY --from=builder /build/dashboard dashboard
RUN chmod +x dashboard

ENTRYPOINT ["/root/dashboard"]