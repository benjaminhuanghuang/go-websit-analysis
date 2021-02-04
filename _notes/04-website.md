

![](./_images/website.png)
## 4-3 上报用户信息数据到打点服务器 (13:22)
tongji.js

## 4-4 Nginx打点服务器的搭建与配置 (25:33)
Download nginx source code


编译安装 Nginx
```
. /configure --prefix=/Useer/<user>/Public/nginx

make

make install
```

Modify Nginx config, 用empty_gif module来响应dig请求，这样可以尽量节省带宽
```
  location = /dig {
    empty_gif;
    eerror_page 405  =200 $request_url;
  }
```
用Nginx 的access.log来记录dig请求的信息
```
  access_log logs/dig.log main


  # define log format
  log_format main '.......'
```

Reload nginx and read dig.log
```
./sbin/nginx -s reload

tail -f logs/dig.log
```