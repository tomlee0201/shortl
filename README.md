简洁的短链程序，可以在[http://shortl.cc](http://shortl.cc)体验。

# Table of Contents

- [特点](#特点)
- [环境准备](#环境准备)
- [配置说明](#配置说明)
- [执行](#执行)
- [测试](#测试)
- [API接口](#API接口)
- [引用源码](#引用源码)
- [特别说明](#特别说明)
- [License](#License)

# 特点
* 生成短链可添加有效期和密码，更加安全
* 记录访问记录，可进行二次开发，实现访问统计数据和对短链进行安全掌控
* 提供API，方面程序调用

# 环境准备
短链服务需要SQL来做持久化存储，在mysql命令行下执行下面语句创建数据库和表
```
sql> source initdb.sql
```

正确执行后，会创建一个shortlink的数据库，数据库里创建一张短链对应表和一张访问记录表

# 配置说明
正确修改shortl.yaml配置文件
```
version: '1'
services:
  shortlink:
    #短链服务的域名
    domain: 'localhost'
    #短链服务的端口
    port: '8080'
    #短链使用的LRU缓存大小，假设每条记录是1K，缓存一百万条需要1G内存，不要超过服务器的可用内存的80%。
    #不需要全部放入到内存缓存中，LRU淘汰的条目如果再访问会从数据库中加载。
    lru_cache_size: 100000

    #user:password@tcp(dbhost:dbport)/dbname?charset=utf8
    #mysql的设置，请正确设置好user password dbhost dbport dbname这几个参数
    db: 'root:123456@tcp(localhost:3306)/shortlink'

```
# 执行
在创建好数据库和修改完配置文件后，把配置文件及resource目录放到程序的相同目录，执行命令
```
./shortl
```

# 测试
使用浏览器打开地址http://host:port 测试验证。作者提供了一个测试环境，点[这里](http://shortl.cc)。

# API接口
服务还提供了API接口供应用调用

| Path | Method | 内容 |  结果 |
| ------ | ------ | ------ | ------ |
| /api/create | POST | Form格式，key "url"为必选参数,值为待转链接；key "duration"可选，短链超时时间，单位为秒; key "password"可选，当使用此参数时，短链必须加上pwd参数  | JSON格式，参考下面示例 | |

示例：
```
curl -d "url=http://www.baidu.com&duration=86400&password=shortlcc" http://shortl.cc/api/create

{"orignal":"http://www.baidu.com","key":"gBr750Kig","domain":"shortl.cc","port":"80"}
```

* 如果没有传password值，短链结果为http://${domain}:${post}/${key}，否则短链为http://${domain}:${post}/${key}?pwd=${password}，示例中的短链为http://shortl.cc/gBr750Kig?pwd=shortlcc ，有效期为1天。

# 引用源码
* LRUCache [xcltapestry/xclpkg](github.com/xcltapestry/xclpkg/)
* UUID [teris-io/shortid](https://github.com/teris-io/shortid)


# 特别说明
体验环境十分脆弱（单核512M内存），请勿用于商业用途。如果您觉得对您所有帮助，请帮忙点star，十分感谢～～



# License
Under the MIT License
