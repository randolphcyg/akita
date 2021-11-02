package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
	"unicode"
	"unsafe"

	"github.com/kirinlabs/HttpRequest"
	"github.com/nosixtools/solarlunar/festival"

	"gitee.com/RandolphCYG/akita/pkg/cache"
)

const (
	characterBytes  = "!@#$%^&*?"
	digitBytes      = "1234567890"
	lowLetterBytes  = "abcdefghijklmnopqrstuvwxyz"
	highLetterBytes = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	letterIdxBits   = 6                    // 6 bits to represent a letter index
	letterIdxMask   = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax    = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

// NewPwd 复杂密码生成器
func NewPwd(length int) (string, error) {
	// check length
	if length < 4 || length > 25 {
		return "", errors.New("length is invalid")
	}
	// assign elements from 4 kinds of base elements
	r := rand.New(rand.NewSource(time.Now().Unix()))
	factor := length / 4
	b := make([]byte, length)
	b = assignElement(lowLetterBytes, length, 1, b)
	b = assignElement(highLetterBytes, length, factor*2, b)
	b = assignElement(characterBytes, length, factor*3, b)
	b = assignElement(digitBytes, length, factor*4, b)
	// shuffle
	rand.Shuffle(len(b), func(i, j int) {
		i = r.Intn(length)
		b[i], b[j] = b[j], b[i]
	})
	return *(*string)(unsafe.Pointer(&b)), nil
}

// assignElement 为密码分配元素
func assignElement(base string, length int, factor int, b []byte) []byte {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	for i, cacheValue, remain := length-factor, r.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cacheValue, remain = r.Int63(), letterIdxMax
		}
		if idx := int(cacheValue & letterIdxMask); idx < len(base) {
			b[i] = base[idx]
			i--
		}
		cacheValue >>= letterIdxBits
		remain--
	}
	return b
}

// SimpleNewPwd 复杂密码生成 简易实现方式
func SimpleNewPwd(length int) (pwd string, err error) {
	// check length
	if length < 4 || length > 25 {
		return "", errors.New("length is invalid")
	}
	// assign elements from 4 kinds of base elements
	r := rand.New(rand.NewSource(time.Now().Unix()))
	factor := length / 4
	characters := []rune(characterBytes)
	digits := []rune(digitBytes)
	highLetters := []rune(highLetterBytes)
	lowerLetters := []rune(lowLetterBytes)
	// get rand elements
	b := make([]rune, length)
	for i := range b[:factor] {
		b[i] = highLetters[r.Intn(len(highLetters))]
	}
	for i := range b[factor : factor*2] {
		b[i+factor] = lowerLetters[r.Intn(len(lowerLetters))]
	}
	for i := range b[factor*2 : factor*3] {
		b[i+factor*2] = characters[r.Intn(len(characters))]
	}
	for i := range b[factor*3:] {
		b[i+factor*3] = digits[r.Intn(len(digits))]
	}
	// shuffle
	rand.Shuffle(len(b), func(i, j int) {
		i = r.Intn(length)
		b[i], b[j] = b[j], b[i]
	})
	return string(b), nil
}

// MiddleNewPwd 复杂密码生成 中等实现方式
func MiddleNewPwd(length int) (pwd string, err error) {
	// check length
	if length < 4 || length > 25 {
		return "", errors.New("length is invalid")
	}
	// assign elements from 4 kinds of base elements
	var pwdBase string
	factor := length / 4
	lastFactor := length - 3*factor
	characters := getRandStr(characterBytes, factor)
	pwdBase += characters
	digits := getRandStr(digitBytes, factor)
	pwdBase += digits
	highLetters := getRandStr(highLetterBytes, factor)
	pwdBase += highLetters
	lowLetters := getRandStr(lowLetterBytes, lastFactor)
	pwdBase += lowLetters
	// shuffle
	var temp []string
	for _, s := range pwdBase {
		temp = append(temp, string(s))
	}
	pwd, err = shuffle(temp)
	if err != nil {
		return
	}
	return
}

// shuffle 打乱元素
func shuffle(slice []string) (str string, err error) {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	length := len(slice)
	// judge length
	if length < 1 {
		return "", errors.New("length is invalid")
	}
	// shuffle
	for i := 0; i < length; i++ {
		randIndex := r.Intn(length) // 随机数
		slice[length-1], slice[randIndex] = slice[randIndex], slice[length-1]
	}
	str = strings.Join(slice, "")
	return
}

// getRandStr 获取随机字符
func getRandStr(baseStr string, length int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano() + rand.Int63()))
	bytes := make([]byte, length)
	l := len(baseStr)
	for i := 0; i < length; i++ {
		bytes[i] = baseStr[r.Intn(l)]
	}
	return string(bytes)
}

// Judge 密码复杂度判断
func Judge(pwd string) bool {
	// 长度不满足
	if len(pwd) < 8 {
		return false
	}
	// 检查字符串元素复杂度
	var flag []int
	for i := 0; i < len(pwd); i++ {
		if unicode.IsLower(rune(pwd[i])) {
			flag = append(flag, 1)
		} else if unicode.IsDigit(rune(pwd[i])) {
			flag = append(flag, 2)
		} else if strings.Contains(characterBytes, string(pwd[i])) {
			flag = append(flag, 3)
		} else if unicode.IsUpper(rune(pwd[i])) {
			flag = append(flag, 4)
		}
	}
	// 复杂度标记切片去重
	complexity := len(removeRepeatedElement(flag[:]))
	if complexity >= 3 {
		return true
	} else {
		return false
	}
}

// removeRepeatedElement 数组去重 通过map键的唯一性去重
func removeRepeatedElement(s []int) []int {
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
		err = errors.New("读取三方系统-c7n配置错误: " + err.Error())
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
		err = errors.New("Fail to fetch token, err: " + err.Error())
		return
	}
	fmt.Println(res.Content())
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
