package util

import (
	"fmt"
	"strconv"
	"strings"
)

var minVip, maxVip int = 168493058, 168558590 // 10.11.0.2 10.11.255.253

var fieldMaps = []int{24, 16, 8, 0}

func addr2oct(addr string) int {
	/*将点分十进制IP地址转换成十进制整数
	 */
	fields := strings.Split(addr, ".")
	var sum int
	for index, fieldStr := range fields {
		fieldInt, err := strconv.Atoi(fieldStr)
		if err == nil {
			var fieldIntOct int = fieldInt << fieldMaps[index]
			sum += fieldIntOct
		} else {
			fmt.Printf("错误: %v\n", err)
		}
	}
	return sum
}

func oct2addr(dec int) string {
	/*将十进制整数IP转换成点分十进制的字符串IP地址
	 */
	var vips []string
	for _, value := range fieldMaps {
		field := strconv.Itoa(dec >> value & 0xff)
		vips = append(vips, field)
	}
	return strings.Join(vips, ".")
}

func assignVip() {
	/*分配虚拟IP redis先调事务一致性锁 然后从redis读取可分配的ip十进制值 vipOct
	 */
	var vipOct int = 168493058
	var vip = oct2addr(vipOct)
	fmt.Print(vip)
}

// func main() {
// 	assignVip()

// 	fmt.Print("\n")

// 	var res = addr2oct("10.11.56.39")
// 	fmt.Print(res + 1)

// }
