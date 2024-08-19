# ghproxy-go
## 一键安装

```bash -c "$(curl -L https://jiasu.in/ghproxy-go.sh)" @ install```

## 一键卸载

```bash -c "$(curl -L https://jiasu.in/ghproxy-go.sh)" @ remove```

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
