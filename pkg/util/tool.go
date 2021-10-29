package util

import (
	"encoding/json"
	"math/rand"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/kirinlabs/HttpRequest"
	"github.com/nosixtools/solarlunar/festival"

	"gitee.com/RandolphCYG/akita/pkg/cache"
	"gitee.com/RandolphCYG/akita/pkg/log"
)

const (
	characterBytes = "!@#$%^&*?"
	digitBytes     = "1234567890"
	letterBytes    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits  = 8                    // 6 bits to represent a letter index
	letterIdxMask  = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax   = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

var (
	src = rand.NewSource(time.Now().UnixNano())
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// PwdGenerator 复杂密码生成器 TODO 复杂密码的排序过于固定
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

// SamplePwdGenerator 最普通的复杂密码生成 不使用版本
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

// Judge 密码复杂度判断
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
	complexity := len(RemoveRepeatedElement(flag[:]))
	if complexity >= 3 {
		return true
	} else {
		return false
	}
}

// RemoveRepeatedElement 数组去重 通过map键的唯一性去重
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

// ExpireStr 根据过期天数计算过期日期的字符串
func ExpireStr(expireDay int) string {
	return time.Now().AddDate(0, 0, expireDay).Format("2006/01/02")
}

// ExpireStrToTime 根据过期日期的字符串转换为日期
func ExpireStrToTime(expireDateStr string) time.Time {
	expireDate, _ := time.Parse("2006/01/02", expireDateStr)
	return expireDate
}

// IsExpire 根据过期日期的字符串计算是否过期
func IsExpire(expireDateStr string) bool {
	expireDate, _ := time.Parse("2006/01/02", expireDateStr)
	return time.Now().After(expireDate)
}

// SendRobotMsg 发送机器人信息
func SendRobotMsg(msg string) {
	// 从缓存取url
	weworkRobotStaffChangesNotifier, err := cache.HGet("third_party_cfgs", "wework_robot_staff_changes_notifier")
	if err != nil {
		log.Log.Error("读取三方系统-c7n配置错误: ", err)
		return
	}
	req := HttpRequest.NewRequest()
	data := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": msg,
		},
	}

	msgPkg, _ := json.Marshal(data)
	res, err := req.Post(weworkRobotStaffChangesNotifier, msgPkg)
	if err != nil {
		// 抛错
		log.Log.Error("Fail to fetch token, err: ", err)
		return
	}
	log.Log.Info(res.Content())
}

// TruncateMsg 裁剪企业微信机器人消息 将长消息按行判断切分，返回消息切片
func TruncateMsg(originalMsg, sep string) (resMsgSegments []string) {
	if len([]byte(originalMsg)) < 4096 {
		resMsgSegments = append(resMsgSegments, originalMsg)
	} else {
		// 按行做切割处理
		msgSegments := strings.Split(originalMsg, sep)

		var segment string
		for _, s := range msgSegments {
			countLen := len([]byte(segment + s))
			if countLen > 4096 {
				resMsgSegments = append(resMsgSegments, segment)
				segment = s + sep
			} else {
				segment += s
				segment += sep
			}
		}
		resMsgSegments = append(resMsgSegments, segment) // 将最后一段消息加上
	}
	return
}

// DnToDepart 将DN地址转换为部门架构
func DnToDepart(dn string) (depart string) {
	rawDn := strings.Split(dn, ",")
	rawDn = Reverse(rawDn[:len(rawDn)-2]) // 去掉DC 逆序
	// 元素拼接，用.替换所有的OU=，去掉开始的.
	depart = strings.Replace(strings.Join(rawDn, ""), "OU=", ".", -1)[1:]
	return
}

// DnToDeparts 将DN地址转换为多级部门切片
func DnToDeparts(dn string) (departs string) {
	rawDn := strings.Split(dn, ",")
	rawDn = Reverse(rawDn[:len(rawDn)-2]) // 去掉DC 逆序
	for i, d := range rawDn {
		rawDn[i] = strings.Trim(strings.ToUpper(d), "OU=")
	}
	departs = strings.Join(rawDn, ".")
	return
}

// Reverse 切片逆序
func Reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// ExpireTime 用户过期期限处理 天数为-1 则过期时间为永久;否则 当前时间往后推迟 expireDays 天
func ExpireTime(expireDays int64) (expireTimestamp int64) {
	expireTimestamp = 9223372036854775807
	if expireDays != -1 {
		expireTimestamp = UnixToNt(time.Now().AddDate(0, 0, int(expireDays)))
	}
	return
}

// Find 判断切片是否有某元素
func Find(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// IsFestival 判断是否为节日
func IsFestival(t time.Time) (isFestival bool, result []string) {
	isFestival = false
	fest := festival.NewFestival("./festival.json")
	result = fest.GetFestivals(t.Format("2006-01-02"))
	if len(result) >= 1 {
		isFestival = true
	}
	return
}

// IsWeekend 判断是否为周末
func IsWeekend(t time.Time) (isWeekend bool) {
	if int(t.Weekday()) == 6 || int(t.Weekday()) == 0 {
		return true
	}
	return false
}

// IsMonday 判断是否为周一
func IsMonday(t time.Time) (isMonday bool) {
	return int(t.Weekday()) == 1
}

// IsHolidaySilentMode 假期静默模式
func IsHolidaySilentMode(t time.Time) (isHolidaySilentMode bool, festival string) {
	isFestival, festivals := IsFestival(t)
	targetFestivals := []string{"除夕", "春节", "国庆节", "中秋节"}
	for _, tf := range targetFestivals {
		_, find := Find(festivals, tf)
		if isFestival && find {
			return true, tf
		}
	}
	// 周末判断
	if IsWeekend(t) {
		return true, ""
	} else {
		return false, ""
	}
}

// UnixToNt Unix 时间转换为 Window NT 时间
func UnixToNt(expireTime time.Time) (ntTimestamp int64) {
	ntTimestamp = expireTime.Unix()*int64(1e+7) + int64(1.1644473600125e+17)
	return
}

// NtToUnix Window NT 时间转换为 Unix 时间
func NtToUnix(ntTime int64) (unixTime time.Time) {
	ntTime = (ntTime - 1.1644473600125e+17) / 1e+7
	return time.Unix(ntTime, 0)
}

// SubDays 计算日期相差多少天 返回值day>0, t1晚于t2; day<0, t1早于t2
func SubDays(t1, t2 time.Time) (day int) {
	swap := false
	if t1.Unix() < t2.Unix() {
		t_ := t1
		t1 = t2
		t2 = t_
		swap = true
	}

	day = int(t1.Sub(t2).Hours() / 24)

	// 计算在被24整除外的时间是否存在跨自然日
	if int(t1.Sub(t2).Milliseconds())%86400000 > int(86400000-t2.Unix()%86400000) {
		day += 1
	}

	if swap {
		day = -day
	}

	return
}

// FormatLdapExpireDays 将原始过期日期规范化为正常设计范围内的过期时间，若未永久不过期，则返回 106752 天
func FormatLdapExpireDays(rawDays int) (validDays int) {
	if rawDays > -100 && rawDays < 200 {
		return rawDays
	} else {
		return 106752
	}
}
