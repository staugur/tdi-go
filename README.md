# tdi-go
🍒花瓣网、堆糖网下载油猴脚本的远程下载服务（Tdi for Go）.

此程序相当于`CrawlHuaban`(中心端)的成员，用户选择远端下载后会由中心端选择一个成员提供给用户，减少中心端压力。

Python版本的仓库地址是：https://github.com/staugur/tdi

PHP版本的仓库地址是：https://github.com/staugur/tdi-php

Node.js版本的仓库地址是：https://github.com/staugur/tdi-node

## 安装

仅支持 Linux 系统！

### 使用 golang 安装

go1.15+, go module(on)

```go
go get -u tcw.im/tdi
mv ~/go/bin/tdi /bin/tdi
tdi -i
```

### 使用 docker 安装

```bash
docker pull staugur/tdi-go
```

[点击查看详细文档](https://docs.saintic.com/tdi-go/)

## Nginx参考

```nginx
server {
    listen 80;
    server_name 域名;
    charset utf-8;
    client_max_body_size 10M;
    client_body_buffer_size 128k;
    location /downloads {
        #下载程序目录
        alias /path/to/downloads/;
        default_type application/octet-stream;
        proxy_max_temp_file_size 0;
        if ($request_filename ~* ^.*?\.(zip|tgz)$){
            add_header Content-Disposition 'attachment;';
        }
    }
    location / {
       #13145是默认端口
       proxy_pass http://127.0.0.1:13145;
       proxy_set_header Host $host;
       proxy_set_header X-Real-IP $remote_addr;
       proxy_set_header X-Forwarded-Proto $scheme;
       proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```
