FROM golang:1.26-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    GOPROXY=direct \
    GONOSUMDB=*

WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -ldflags="-s -w" -o app .

FROM alpine:3.23
WORKDIR /apps
ENV LANG=en_US.UTF-8
ENV APP_ENV=COS

RUN echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/v3.23/main" > /etc/apk/repositories && \
    echo "http://mirrors.tuna.tsinghua.edu.cn/alpine/v3.23/community" >> /etc/apk/repositories && \
    apk add --no-cache tzdata ca-certificates && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    update-ca-certificates

COPY --from=builder /build/app .

EXPOSE 80
ENTRYPOINT ["./app"]
