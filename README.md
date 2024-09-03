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
## 在域名后拼接 GitHub 链接即可，协议头可以选择性添加

```
https://gh.jiasu.in/github.com/0-RTT/ghproxy-go/blob/main/main.go

https://gh.jiasu.in/https://github.com/0-RTT/ghproxy-go/blob/main/main.go

https://gh.jiasu.in/raw.githubusercontent.com/0-RTT/ghproxy-go/main/main.go

https://gh.jiasu.in/https://raw.githubusercontent.com/0-RTT/ghproxy-go/main/main.go

git clone https://gh.jiasu.in/github.com/0-RTT/ghproxy-go.git

git clone https://gh.jiasu.in/https://github.com/0-RTT/ghproxy-go.git

```
