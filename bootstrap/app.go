package bootstrap

import (
	"fmt"

	"gitee.com/RandolphCYG/akita/pkg/conf"
)

// InitApplication 初始化应用常量
func InitApplication() {
	fmt.Print(conf.Logo + conf.BackendVersion + `  Commit #` + conf.LastCommit + `  Pro=` + conf.IsPro + `
================================================

`)
}
