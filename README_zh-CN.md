# Akita

## 1. 项目介绍

此项目负责管理active diretory(ldap)用户和处理企业微信工单，做到用户从入职到离职整个生命周期的账号自治，用户主要用企业微信工单与管理员、后端系统交互。发信可以配置成企业微信和邮件通知。

## 2. 启动

```shell
go run main.go
# 触发调用定时任务(其中一个)
GET：http://127.0.0.1:8099/api/v1/ldap/user/start
```

暂无前端

## 3. 目录介绍

`./main.go`项目启动入口

### `/pkg`

其他项目也可以引用的公共包，redis、mysql、log、email、crontab、wework的封装

### `/web`

Web应用程序的目录——静态资源，未来若有前端项目打包放在这里一起发布。

### `/config`

通用应用程序的目录——配置文件模板或默认配置。

`config.yaml`

```yaml
system:
  Mode: debug
  Addr: 127.0.0.1:8099
  Debug: true
  
database:
  Type: mysql
  UserName: root
  Password: root
  Addr: 127.0.0.1:3306
  Name: akita
  ShowLog: true                   # 是否打印SQL日志
  MaxIdleConn: 10                 # 最大闲置的连接数,0意味着使用默认的大小2, 小于0表示不使用连接池
  MaxOpenConn: 60                 # 最大打开的连接数, 需要小于数据库配置中的max_connections数
  ConnMaxLifeTime: 60m            # 单个连接最大存活时间,建议设置比数据库超时时长(wait_timeout)稍小一些

redis:
  Addr: 127.0.0.1:6379
  Password: ""
  DB: 0
  PoolSize: 12000
  DialTimeout: 60s
  ReadTimeout: 500ms
  WriteTimeout: 500ms

ldapCfg:
  ConnUrl:       ldap://xxxx:636
  BaseDn:        DC=xxx,DC=com
  AdminAccount:  CN=Administrators,CN=Users,DC=xxx,DC=com
  Password:      xxxxxxxxxxxx
  SslEncryption: False
  Timeout:       5

email:
  Host: smtp.exmail.qq.com              # SMTP地址
  Port: 25                              # 端口
  Username: devops@xxxx.com             # 用户名
  Password: xxx                         # 密码
  NickName: SRE                         # 发送者名称
  Address: devops@xxxxx.com    		   # 发送者邮箱
  ReplyTo: NULL                         # 回复地址
  KeepAlive: 30                         # 连接保持时长
```
