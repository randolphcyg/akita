package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPwd(t *testing.T) {
	pwd0, _ := NewPwd(8)
	fmt.Println(pwd0)
	// 测试密码复杂度
	assert.Equal(t, Judge(pwd0), true, "not satisfied!")

	//pwd1, _ := NewPwd(3)
	//fmt.Println(pwd1)
	//// 测试密码复杂度
	//assert.Equal(t, Judge(pwd1), true, "not satisfied!")

	pwd2, _ := NewPwd(16)
	fmt.Println(pwd2)
	// 测试密码复杂度
	assert.Equal(t, Judge(pwd2), true, "not satisfied!")

	//pwd3, _ := NewPwd(28)
	//fmt.Println(pwd3)
	//// 测试密码复杂度
	//assert.Equal(t, Judge(pwd3), true, "not satisfied!")
}

func TestSimpleNewPwd(t *testing.T) {
	pwd0, _ := SimpleNewPwd(8)
	fmt.Println(pwd0)
	// 测试密码复杂度
	assert.Equal(t, Judge(pwd0), true, "not satisfied!")

	//pwd1, _ := SimpleNewPwd(3)
	//fmt.Println(pwd1)
	//// 测试密码复杂度
	//assert.Equal(t, Judge(pwd1), true, "not satisfied!")

	pwd2, _ := SimpleNewPwd(16)
	fmt.Println(pwd2)
	// 测试密码复杂度
	assert.Equal(t, Judge(pwd2), true, "not satisfied!")

	//pwd3, _ := SimpleNewPwd(28)
	//fmt.Println(pwd3)
	//// 测试密码复杂度
	//assert.Equal(t, Judge(pwd3), true, "not satisfied!")
}

func TestMiddleNewPwd(t *testing.T) {
	pwd0, _ := MiddleNewPwd(8)
	fmt.Println(pwd0)
	// 测试密码复杂度
	assert.Equal(t, Judge(pwd0), true, "not satisfied!")

	//pwd1, _ := MiddleNewPwd(3)
	//fmt.Println(pwd1)
	//// 测试密码复杂度
	//assert.Equal(t, Judge(pwd1), true, "not satisfied!")

	pwd2, _ := MiddleNewPwd(16)
	fmt.Println(pwd2)
	// 测试密码复杂度
	assert.Equal(t, Judge(pwd2), true, "not satisfied!")

	//pwd3, _ := MiddleNewPwd(28)
	//fmt.Println(pwd3)
	//// 测试密码复杂度
	//assert.Equal(t, Judge(pwd3), true, "not satisfied!")
}
