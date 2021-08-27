package bootstrap

import (
	"fmt"
)

// InitApplication 初始化应用常量
func InitApplication() {
	fmt.Print(Logo + BackendVersion + `
================================================

`)
}
