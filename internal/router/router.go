package router

import (
	"fmt"
	"os"
	"strings"

	"gitee.com/RandolphCYG/akita/bootstrap"
	"gitee.com/RandolphCYG/akita/internal/conf"
	"gitee.com/RandolphCYG/akita/internal/handler"
	"gitee.com/RandolphCYG/akita/internal/model"
	"gitee.com/RandolphCYG/akita/internal/service/task"
	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/email"
	"gitee.com/RandolphCYG/akita/pkg/hr"
	"gitee.com/RandolphCYG/akita/pkg/log"

	"github.com/gin-gonic/gin"
)

var r *gin.Engine

func init() {
	gin.SetMode(gin.ReleaseMode) //  先设置为生产模式 保证日志静默
	r = gin.New()                // 初始化gin引擎
	r.Use(log.LogerMiddleware()) // 日志中间件
}

func Init(cfg string) {
	// 初始化logo
	bootstrap.InitApplication()
	// 初始化系统配置
	Cfg, err := conf.Init(cfg)
	if err != nil {
		panic(err)
	}

	// 根据路由模式执行操作
	initRouterMode()

	// 初始化 db
	model.Init(&Cfg.Database)
	// 执行数据迁移
	log.Log.Info("Data migration begin ...")
	model.DB.AutoMigrate(&model.LdapCfg{}, &model.LdapField{}, &hr.HrDataConn{}, &model.WeworkCfg{}, &model.WeworkOrder{}, &model.LdapUserDepartRecord{}, &model.WeworkUserSyncRecord{})
	if result := model.DB.Limit(1).Find(&model.LdapCfg{}); result.RowsAffected == 0 {
		model.DB.Create(&Cfg.LdapCfg)
	}
	log.Log.Info("Data migration successful ...")

	// 初始化 redis
	cache.Init(&Cfg.Redis)

	// 初始化 email
	email.Init(&Cfg.Email)

	// 初始化ldap连接
	model.LdapCfgs, _ = model.GetAllLdapConn() // 直接查询
	// 初始化ldap字段配置
	model.LdapFields, _ = model.GetLdapFieldByConnUrl(model.LdapCfgs.ConnUrl)

	// 初始化企业微信配置信息
	err = model.GetWeworkOrderCfg()
	if err != nil {
		log.Log.Error("初始化企业微信审批配置信息错误, err: ", err)
	}
	err = model.GetWeworkUuapCfg()
	if err != nil {
		log.Log.Error("初始化企业微信UUAP公告应用配置信息错误, err: ", err)
	}
	err = model.GetWeworkUserManageCfg()
	if err != nil {
		log.Log.Error("初始化企业微信通讯录管理配置信息错误, err: ", err)
	}
	log.Log.Info("UUAP server init successful ...")
}

// initRouterMode 根据路由模式执行操作
func initRouterMode() {
	switch conf.Conf.System.Mode {
	case "debug":
		gin.SetMode(gin.DebugMode)
		log.Log.Warn("Debug mode will not start crontab tasks !!!")
	case "release":
		gin.SetMode(gin.ReleaseMode)
		task.StartAll() // 启动所有定时任务
	case "test":
		gin.SetMode(gin.TestMode)
	default:
		panic("gin mode unknown: " + conf.Conf.System.Mode + " (available mode: debug release test)")
	}
}

// router 路由
func InitRouter() *gin.Engine {
	v1 := r.Group("/api/v1")
	{
		// 全局设置
		site := v1.Group("site")
		site.GET("ping", handler.Ping) // 就绪探针

		// ldap 连接
		ldapConns := v1.Group("ldap/conns")
		ldapConns.GET("fetch", handler.FetchLdapConn)
		ldapConns.POST("create", handler.CreateLdapConn)
		ldapConns.POST("update", handler.UpdateLdapConn)
		ldapConns.DELETE("delete", handler.DeleteLdapConn)
		ldapConns.POST("test", handler.TestLdapConn)
		// ldap 连接的字段明细配置
		ldapFields := v1.Group("ldap/fields")
		ldapFields.GET("fetch", handler.FetchLdapField)
		ldapFields.POST("create", handler.CreateLdapField)
		ldapFields.POST("update", handler.UpdateLdapField)
		ldapFields.DELETE("delete", handler.DeleteLdapField)
		ldapFields.POST("test", handler.TestLdapField)
		// ldap 用户
		ldapUsers := v1.Group("ldap/users")
		ldapUsers.GET("manual/cache/hr", handler.CacheHrUsersManual)            // 手动触发缓存HR用户
		ldapUsers.GET("manual/sync", handler.SyncLdapUsersManual)               // 手动触发更新ldap用户
		ldapUsers.GET("manual/scan/expire", handler.ScanExpiredLdapUsersManual) // 手动触发扫描过期ldap用户

		// wework 工单
		weworkOrders := v1.Group("wework/orders")
		weworkOrders.POST("handle", handler.HandleOrders)
		// wework 用户
		weworkUsers := v1.Group("wework/users")
		weworkUsers.GET("manual/cache", handler.CacheUsersManual)                   // 手动触发缓存企业微信用户
		weworkUsers.GET("manual/scan/expire", handler.ScanExpiredWeworkUsersManual) // 手动触发扫描企业微信过期用户
		weworkUsers.GET("manual/scan/new", handler.ScanNewHrUsersManual)            // 手动触发扫描HR缓存数据并为新员工创建企业微信账号

		// c7n 项目
		c7nProjects := v1.Group("c7n/projects")
		c7nProjects.GET("manual/cache", handler.CacheC7nProjectsManual) // 手动触发缓存C7N项目
		// c7n 用户
		c7nUsers := v1.Group("c7n/users")
		c7nUsers.GET("manual/sync", handler.UpdateC7nUsersManual) // 手动触发LDAP用户同步到C7N

		// tasks 定时任务
		tasks := v1.Group("tasks")
		tasks.GET("fetchall", handler.FetchAll)  // 停止定时任务
		tasks.POST("start", handler.TaskStart)   // 启动定时任务
		tasks.POST("remove", handler.TaskRemove) // 移除定时任务
		tasks.GET("stop", handler.TaskStop)      // 停止所有定时任务
	}

	// 生产模式打印路由
	if gin.Mode() == gin.ReleaseMode {
		for _, p := range r.Routes() {
			nuHandlers := len(r.Handlers)
			routePrint("%-6s %-25s --> %s (%d handlers)\n", p.Method, p.Path, p.Handler, nuHandlers)
			// "+"http://"+conf.Conf.System.Addr+"  // 不提供完整地址防止用户点击太快导致触发不该执行的功能
		}
	}

	return r
}

// 重写路由打印方法
func routePrint(format string, values ...interface{}) {
	if !strings.HasSuffix(format, "\n") {
		format += "\n"
	}
	fmt.Fprintf(os.Stdout, "[GIN-release] "+format, values...)
}
