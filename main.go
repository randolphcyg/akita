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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
)

var cfgFile = pflag.StringP("config", "c", "", "指定akita配置文件地址")

func init() {
	bootstrap.Init(*cfgFile) // 初始化系统配置
}

func main() {
	api := router.InitRouter()
	if err := api.Run(conf.Conf.System.Addr); err != nil {
		log.Error("Fail to listen server on "+"http://"+conf.Conf.System.Addr, err)
	}
	log.Info("Success to listen server on " + "http://" + conf.Conf.System.Addr)
}
