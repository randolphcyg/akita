package util

import (
	"time"
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
