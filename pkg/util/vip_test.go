package util

import (
	"log"
	"strconv"
	"strings"
	"testing"
)

// VipToOct
func TestVipToOct(t *testing.T) {
	VipToOct("10.11.0.2")
}

// OctToVip
func TestOctToVip(t *testing.T) {
	OctToVip(168493058)
}

// OctToVipArray 数组
func OctToVipArray(vipOct uint64) (addr string) {
	if vipOct < VIPMIN && vipOct > VIPMAX {
		log.Fatal("传入错误十进制整数，检查是否所有虚拟IP都分配出去")
		return
	} else {
		var vips [4]string
		for index, value := range fieldMaps {
			vips[index] = strconv.FormatUint(vipOct>>value&0xff, 10)
		}
		return strings.Join(vips[:], ".")
	}
}

// octToVipSlice 切片
func OctToVipSlice(vipOct uint64) (addr string) {
	if vipOct < VIPMIN && vipOct > VIPMAX {
		log.Fatal("传入错误十进制整数，检查是否所有虚拟IP都分配出去")
		return
	} else {
		var vips []string
		for _, value := range fieldMaps {
			vips = append(vips, strconv.FormatUint(vipOct>>value&0xff, 10))
		}
		return strings.Join(vips, ".")
	}
}

// 基准测试
func BenchmarkA(b *testing.B) {
	b.ResetTimer()
	for index := uint64(168507432); index <= uint64(168558590); index++ {
		OctToVipArray(168493058)
	}
}

func BenchmarkB(b *testing.B) {
	b.ResetTimer()
	for index := uint64(168507432); index <= uint64(168558590); index++ {
		OctToVipSlice(168493058)
	}
}
