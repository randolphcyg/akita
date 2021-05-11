package util

import (
	"math/rand"
	"time"
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

const letterBytes = "1234567890abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*?"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// 复杂密码生成器
func RandStringBytesMaskImprSrcUnsafe(n int) string {
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

	return *(*string)(unsafe.Pointer(&b))
}
