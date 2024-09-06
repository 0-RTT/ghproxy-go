# ghproxy-go
## 一键安装

```
bash -c "$(curl -L https://raw.githubusercontent.com/0-RTT/index/main/ghproxy-go.sh)" @ install
```

## 一键卸载

```
bash -c "$(curl -L https://raw.githubusercontent.com/0-RTT/index/main/ghproxy-go.sh)" @ remove
```

## 配置nginx反代

```    
    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
  ```

## GitHub加速使用方法：直接在GitHub链接后拼接即可

```
git clone https://gh.jiasu.in/github.com/0-RTT/ghproxy-go.git

git clone https://gh.jiasu.in/https://github.com/0-RTT/ghproxy-go.git

https://gh.jiasu.in/github.com/0-RTT/ghproxy-go/blob/main/main.go

https://gh.jiasu.in/https://github.com/0-RTT/ghproxy-go/blob/main/main.go

```

## Docker Hub加速使用方法（由nginx反代实现和本项目无关）：

### 第一步：设置registry mirror
```
sudo tee /etc/docker/daemon.json <<EOF
{
    "registry-mirrors": [
        "https://gh.jiasu.in"
    ]
}
EOF
```
### 第二步：重新加载 systemd
```
sudo systemctl daemon-reload
```
### 第三步：重启 Docker
```
sudo systemctl restart docker
```

# 下面是同个域名实现GitHub和Docker Hub加速的nginx反代配置，仅供参考！
```
location /v2 {
    proxy_pass https://registry-1.docker.io;  
    proxy_set_header Host registry-1.docker.io;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;  
    proxy_buffering off;
    proxy_set_header Authorization $http_authorization;
    proxy_pass_header Authorization;
    proxy_intercept_errors on;
    recursive_error_pages on;
    error_page 301 302 307 = @handle_redirect;
}

location / {
    proxy_pass http://127.0.0.1:8080;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}

location @handle_redirect {
    resolver 1.1.1.1;
    set $saved_redirect_location '$upstream_http_location';
    proxy_pass $saved_redirect_location;
}
```
