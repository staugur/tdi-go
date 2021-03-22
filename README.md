# tdi-go

🍒 花瓣网、堆糖网下载油猴脚本的远程下载服务（Tdi for Go）.

此程序相当于` [CrawlHuaban](https://open.sainitc.com/CrawlHuaban/Register)(中心端)的成员，用户选择远端下载后会由中心端选择一个成员提供给用户，减少中心端压力。

Python 版本的仓库地址是：https://github.com/staugur/tdi

PHP 版本的仓库地址是：https://github.com/staugur/tdi-php

Node.js 版本的仓库地址是：https://github.com/staugur/tdi-node

## 安装

仅支持 Linux 系统！

### 使用 golang 安装

go1.15+, go module(on)

```go
go get -u tcw.im/tdi
mv ~/go/bin/tdi /bin/tdi
tdi -i
```

**或者从源码安装**

```bash
git clone https://github.com/staugur/tdi-go && cd tdi-go
make build
mv bin/tdi /bin/tdi
tdi -i
```

### 使用 docker 安装

```bash
docker pull staugur/tdi-go
```

[点击查看详细文档](https://docs.saintic.com/tdi-go/)
