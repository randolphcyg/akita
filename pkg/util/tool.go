package util

import (
	"math/rand"
	"strings"
	"time"
	"unicode"
	"unsafe"
)

// Unix 时间转换为 Window NT 时间
func UnixToNt(expireTime time.Time) (ntTimestamp int64) {
	ntTimestamp = expireTime.Unix()*int64(1e+7) + int64(1.1644473600125e+17)
	return
}

// Window NT 时间转换为 Unix 时间
func NtToUnix(ntTime int64) (unixTime time.Time) {
	ntTime = (ntTime - 1.1644473600125e+17) / 1e+7
	return time.Unix(int64(ntTime), 0)
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var src = rand.NewSource(time.Now().UnixNano())

const (
	characterBytes = "!@#$%^&*?"
	digitBytes     = "1234567890"
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits  = 8                    // 6 bits to represent a letter index
	letterIdxMask  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax   = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// 复杂密码生成器 TODO 复杂密码的排序过于固定
func PwdGenerator(n int) string {
	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	// 特殊符号
	for i, cache, remain := n-1-3, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(characterBytes) {
			b[i+3] = characterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}
	// 数字
	for i, cache, remain := n-1-5, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(digitBytes) {
			b[i+4] = digitBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// 最普通的复杂密码生成 不使用版本
func SamplePwdGenerator(n int) (pwd string) {
	characters := []rune("!@#$%^&*?")
	digits := []rune("1234567890")
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, n)
	for i := range b[:3] {
		b[i] = letters[rand.Intn(len(letters))]
	}
	for i := range b[3:5] {
		b[i+3] = characters[rand.Intn(len(characters))]
	}
	for i := range b[5:7] {
		b[i+5] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

// 密码复杂度判断
func Judge(pwd string) (isValid bool) {
	characters := "!@#$%^&*?"
	if len(pwd) < 8 {
		return false
	}
	var flag = [...]int{0, 0, 0, 0, 0, 0, 0, 0}
	for i := 0; i < len(pwd); i++ {
		// fmt.Printf("%c\n", pwd[i])
		if unicode.IsLower(rune(pwd[i])) {
			flag[i] = 1
		} else if unicode.IsDigit(rune(pwd[i])) {
			flag[i] = 2
		} else if strings.Contains(characters, string(pwd[i])) {
			flag[i] = 3
		} else if unicode.IsUpper(rune(pwd[i])) {
			flag[i] = 4
		}
	}
	complex := len(RemoveRepeatedElement(flag[:]))
	if complex >= 3 {
		return true
	} else {
		return false
	}
}

// 数组去重 通过map键的唯一性去重
func RemoveRepeatedElement(s []int) []int {
	result := make([]int, 0)
	m := make(map[int]bool) //map的值不重要
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
			m[v] = true
		}
	}
	return result
}
