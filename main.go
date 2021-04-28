/**
 *
 *	  _    _    _  _
 * 	 /_\  | |__(_)| |_  __ _
 *  / _ \ | / /| ||  _|/ _` |
 * /_/ \_\|_\_\|_| \__|\__,_|
 *
 * generate by http://patorjk.com/software/taag/#p=display&h=1&f=Small&t=Akita
 */

package main

import (
	"gitee.com/RandolphCYG/akita/bootstrap"
	"gitee.com/RandolphCYG/akita/internal/conf"

	"gitee.com/RandolphCYG/akita/internal/router"
	"gitee.com/RandolphCYG/akita/pkg/crontab"
	"gitee.com/RandolphCYG/akita/pkg/log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/pflag"
)

var cfgFile = pflag.StringP("config", "c", "", "指定akita配置文件地址")

func init() {
	// 初始化
	bootstrap.Init(*cfgFile)
}

func main() {
	api := router.InitRouter()
	// 开始执行定时任务 TODO 这边需要确定下定时任务不影响系统启动
	crontab.Init()
	log.Log().Info("开始监听 %s", conf.Conf.System.Addr)
	if err := api.Run(conf.Conf.System.Addr); err != nil {
		log.Log().Error("无法监听[%s],%s", conf.Conf.System.Addr, err)
	}

}
