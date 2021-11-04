# Akita

## 1. 项目介绍

此项目负责管理Active Diretory Domain(ldap)用户、企微用户和处理企微工单，做到用户从入职到离职整个生命周期的账号自治，用户主要用企业微信工单与管理员、后端系统交互。发信可以配置成企业微信和邮件通知。

## 2. 开发

该项目主要在go1.17环境开发，有开发需求时，请注意go版本。

```shell
git clone git@gitee.com:RandolphCYG/akita.git
# 进入项目根目录
cd akita
# 安装go依赖
go mod tidy
# 修改配置文件 build\package\akita\app\config\config.yaml 本地开发的话可以将`system.Mode`改成`debug`、`system.Debug`改成true；此时本地开发时候不会自动启动所有定时任务；
# 注意本地需要配置`mysql`、`redis`、并将生产数据库和缓存库的数据克隆到开发机器，将配置文件改成本地数据库，尽量不要影响生产数据。
go run main.go
```

main.go为程序主入口，其init函数作用是初始化系统配置`bootstrap.Init(*cfgFile)` ，此步骤将主配置文件读取后启动`mysql`、`redis`，并将`ldap`和`企业微信`、`邮件服务器`配置初始化到环境变量中；
main函数起了个主线程，`engine := router.InitRouter()`作用是初始化路由，根据系统模式决定是否启动全部定时任务；


## 3. 目录介绍

`./main.go`项目启动入口

### `/bootstrap`

服务初始化作用

### `/pkg`

- 其他项目也可以引用的公共包，对数据库、缓存、邮件、外部系统、通用组件、企微接口等的封装；
- pkg中的每个包尽量独立，pkg包之前尽量不互相依赖，internal依赖于pkg，绝对不要写反向依赖。

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
  Host: smtp.exmail.qq.com             # SMTP地址
  Port: 25                             # 端口
  Username: sys@xxxx.com               # 用户名
  Password: xxx                        # 密码
  NickName: system                     # 发送者名称
  Address: devops@xxxxx.com    		   # 发送者邮箱
  ReplyTo: NULL                        # 回复地址
  KeepAlive: 30                        # 连接保持时长
```

### `/internal`

服务主要的业务内容，因为功能并不复杂，所以采用了简单清晰的四层架构，尽量减少反向依赖；

1. router 路由，因为路由并不多，因此没有将路由分到各功能
2. handler 处理程序，该层接受请求，调用存储库层并满足业务流程并发送响应；
3. service 服务，主要的业务处理逻辑都放在这里；
4. model 模型，将映射数据库表的结构体和orm操作都放在这里；
5. middleware 中间件 日志中间件和超时中间件

## 4. 特殊点简介

1. 企业微信工单解析

企业微信依赖在`/pkg/wework/api`中，企业微信原始工单的通用解析在`/pkg/wework/order/wework.go`的结构体`WeworkOrder`，这是第一层解析；

类似于`AccountsRegister`是对应具体工单的解析，此步骤将工单解析为对应结构体，值得注意的是mapstructure映射要和工单中的字段名称相同，`spName`、`userid`、`remark`等是工单通用的默认字段。

2. 定时任务

在`/internal/service/task/task.go`中的`init`方法完成对定时任务的注册，其后在初始化gin的路由时调用`InitTasks`函数将所有注册了的定时任务`AllTasks`加到Schedule中。

3. ldap

基于ldap模块，和企业人员信息特点，封装了根据`真实姓名`+`工号`的查询方式，组合字段对应ldap中用户的cn即`commonName`,外部公司人员也遵循这一规则；

例如外部的人员，工号是`9527`，则其`sam`账号为`OD9527`，但是提交工单时候依然正常填写姓名工号即可。所有人的工号都可以在企业微信自己的工号字段查询。

ldap_fields表的company_type字段：

```
{"本公司":{"is_outer":false,"prefix":""},"其他公司":{"is_outer":true,"prefix":"OD"}}
```

如果有其他公司的，需要管理员更新下这个字段(和其他公司相同规则就行)————`is_outer: true`、`prefix: "大写尽量简短的英文"`，企业微信工单的`公司`字段加上对应的公司，原先计划这些维护项目是做一个前端方便管理的。


4. 缓存

因为核心业务太小众化，因此无需放出来，功能也是无效的，该项目结构简单，可以用来写较小的后端项目；

`Init`函数可以用来初始化缓存连接池，另外也封装了对其他字段的操作方法。

