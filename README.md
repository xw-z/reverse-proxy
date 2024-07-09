# reverse-proxy

```
docker pull xwzhou/reverse-proxy

docker pull registry.cn-hangzhou.aliyuncs.com/mirror/xwzhou_reverse-proxy
docker tag registry.cn-hangzhou.aliyuncs.com/mirror/xwzhou_reverse-proxy xwzhou/reverse-proxy
```


```
docker run -d --restart=always \
--network host \
--name docker-example \
--log-opt max-size=10m --log-opt max-file=3 \
xwzhou/reverse-proxy \
--http_addr=0.0.0.0:80 --target=http://example.com:80

docker run -d --restart=always \
--network host \
--name docker-proxy \
-v /var/run/docker.sock:/var/run/docker.sock:ro \
-v ~/.docker:/root/.docker \
--log-opt max-size=10m --log-opt max-file=3 \
xwzhou/reverse-proxy \
--http_addr=0.0.0.0:2375 --https_addr=0.0.0.0:2376 --target=unix:///var/run/docker.sock
```
