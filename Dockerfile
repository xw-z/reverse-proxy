### build back-end
FROM golang:1.22-alpine as go-builder
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

ENV APP_MODE=prod
ENV APP_LOG=info
ENV APP_ACCESS_LOG=true

WORKDIR /app
COPY ./ ./

RUN go build -v -o bin/ -mod vendor ./ \
    && rm -fr app vendor libs

### production
FROM alpine:3
LABEL AUTHOR="xwzhou@yeah.net"
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache ca-certificates curl bash
ENV TZ=Asia/Shanghai
RUN apk --no-cache add tzdata  && \
    ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && \
    echo $TZ > /etc/timezone
ENV APP_MODE=prod
ENV APP_LOG=info
ENV APP_ACCESS_LOG=true

WORKDIR /app
#COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go-builder /app/bin /bin

ENTRYPOINT ["reverse-proxy"]

EXPOSE 2375
EXPOSE 2376