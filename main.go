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

	"github.com/spf13/pflag"
)

var cfgFile = pflag.StringP("config", "c", "", "指定akita配置文件地址")

func init() {
	bootstrap.Init(*cfgFile) // 初始化系统配置
}

func main() {
	engine := router.InitRouter() // 初始化路由
	if err := engine.Run(conf.Conf.System.Addr); err == nil {
		panic("Fail to listen server on " + "http://" + conf.Conf.System.Addr + " ERR: " + err.Error())
	}
}
