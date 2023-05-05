## 简介
基于docker的NAS系统，特点是简单、免费开源、跨平台

文档地址：http://doc.dockernas.com

可以使用docker部署
```sh
docker run -d --name dockernas --restart always --add-host=host.docker.internal:host-gateway -p 8080:8080 -v /var/run/docker.sock:/var/run/docker.sock -v /root/docker/data/nas:/home/dockernas/data gwbc/dockernas
```

目前主要在windows上测试，Linux下问题可能相对多些，可以提issue反馈

## 编译方法
代码编译方式如下所示
```sh
cd frontend 
npm install
npm run build
cd ..
go build ./dockernas.go
```
docker镜像构建方式如下
```sh
#注意需要先在本地构建好前端代码
docker build . -t dockernas
#多平台构建，构建后直接push到dockerhub
docker buildx build --platform linux/arm64,linux/amd64 -t gwbc/dockernas:latest . --push
```

## docker开启远程
```sh
#ubuntu
vi /lib/systemd/system/docker.service

#[Service] -> ExecStart 添加
-H tcp://0.0.0.0:2375

systemctl daemon-reload 
systemctl restart docker
```

## 配置文件
```json
{
    "basePath": "/home/data",               #docker启动应用的根路径
    "bindAddr": "0.0.0.0:8080",             #web监听端口
    "dockerSvrIP": "172.16.100.226:2345",   #docker远程服务IP
    "passwd": "zhang",                      #web密码
    "user": "admin"                         #web用户名
}
```
