package util

import (
	"fmt"
	"log"
	"strconv"
	"strings"
)

// 对应的虚拟IP段是 10.11.0.2 10.11.255.253 最小和最大可以分配的虚拟IP对应的uint64
const VIPMIN, VIPMAX uint64 = 168493058, 168558590

var fieldMaps = [4]uint64{24, 16, 8, 0}

// VipToOct 将IP地址转换成十进制整数
func VipToOct(addr string) (vipOct uint64) {
	fields := strings.Split(addr, ".")
	for index, fieldStr := range fields {
		fieldInt, err := strconv.ParseUint(fieldStr, 10, 0)
		if err == nil {
			fieldInt <<= fieldMaps[index]
			vipOct += fieldInt
		} else {
			fmt.Printf("错误: %v\n", err)
		}
	}
	return vipOct
}

// OctToVip 将十进制整数IP转换成IP地址字符串
func OctToVip(vipOct uint64) (addr string) {
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
