FROM hub.zentao.net/ci/golang:1.25-alpine3.22 AS build

ARG TZ="Asia/Shanghai"

ENV TZ=${TZ}

ENV GOPROXY=https://goproxy.cn,direct

ENV GO111MODULE=on
ENV CGO_ENABLED=1

# 安装 C 编译依赖（gcc、musl-dev、sqlite-dev），用于编译 go-sqlite3
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add git bash musl tzdata curl \
    # 安装 C 依赖（关键：gcc 是编译 CGO 必需，sqlite-dev 是 SQLite 库文件）
    && apk --no-cache add gcc musl-dev sqlite-dev

WORKDIR /app

COPY . .

RUN go install github.com/go-task/task/v3/cmd/task@latest \
  && task build

### Pull CA Certs
FROM hub.zentao.net/ci/alpine:3.19 AS cert-image

RUN apk --update add ca-certificates

FROM hub.zentao.net/ci/alpine:3.19

ARG TZ="Asia/Shanghai"

ENV TZ=${TZ}

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.tuna.tsinghua.edu.cn/g' /etc/apk/repositories \
    && apk --no-cache add git bash musl tzdata curl \
    && ln -sf /usr/share/zoneinfo/${TZ} /etc/localtime \
    && echo ${TZ} > /etc/timezone \
    && date \
    # && apk del --no-cache tzdata \
    && rm -rf /var/cache/apk/*

# setup app dir and its content
WORKDIR /app

VOLUME /data

# ENV GITFOX_CI_PLUGINS_ZIP_URL=https://pkg.zentao.net/gitfox/20260123/plugins.zip
ENV GITFOX_CI_PLUGINS_ZIP_URL=/app/ciPlugins/plugins.tar.gz

ENV GITFOX_LOG_DIR=/data/logs/gitfox
ENV GITFOX_GIT_ROOT_DIR=/data/gitfox/git
ENV GITFOX_SSH_HOST_KEYS_DIR=/data/gitfox/ssh
ENV GITFOX_XDG_CACHE_HOME=/data/gitfox/cache
ENV GITFOX_REGISTRY_FS_ROOT_DIR=/data/gitfox/registry
ENV GITFOX_SQLITE_FILE=/data/gitfox/sqlite/sqlite.db
# 日志输出的最低级别  -1~5：debug,info,warn,error,dpanic,panic,fatal
ENV GITFOX_LOG_LEVEL=-1
ENV GITFOX_DB_DRIVER=mysql
ENV GITFOX_DB_DATASOURCE=root:pass4GitFox@tcp(mysql:3306)/gitfox-community?parseTime=true&loc=Local
ENV GITFOX_DOCKER_HOST=unix:///var/run/docker.sock
ENV GITFOX_DOCKER_API_VERSION=1.40
ENV GITFOX_URL_BASE=http://gitfox.oop.cc:3052

COPY --from=build /app/bin/gitfox_linux /app/gitfox
COPY --from=build /app/configs/dev.env /app/dev.env
COPY --from=build /app/configs/dev.yaml /app/dev.yaml
COPY --from=build /app/resources/ciPlugins/plugins.tar.gz /app/ciPlugins/plugins.tar.gz



COPY --from=cert-image /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

EXPOSE 3000
EXPOSE 22
EXPOSE 3322

CMD [ "/app/gitfox", "server", "dev" ]
