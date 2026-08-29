package validator

import (
	"regexp"

	"github.com/go-playground/validator/v10"
)

// IsIp4 验证是否为有效的 IPv4 地址
func IsIp4(field validator.FieldLevel) bool {
	addr := field.Field().String()
	regStr := `^(([1-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.)(([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\.){2}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])$`
	if match, _ := regexp.MatchString(regStr, addr); match {
		return true
	}
	return false
}

// IsYMD 验证是否为 YYYY-MM-DD 格式
func IsYMD(field validator.FieldLevel) bool {
	//(?!0000)  闰年:2016-02-29
	str := field.Field().String()
	regular := `^((((1[6-9]|[2-9]\d)\d{2})-(0?[13578]|1[02])-(0?[1-9]|[12]\d|3[01]))|(((1[6-9]|[2-9]\d)\d{2})-(0?[13456789]|1[012])-(0?[1-9]|[12]\d|30))|(((1[6-9]|[2-9]\d)\d{2})-0?2-(0?[1-9]|1\d|2[0-8]))|(((1[6-9]|[2-9]\d)(0[48]|[2468][048]|[13579][26])|((16|[2468][048]|[3579][26])00))-0?2-29-))$`
	return regexp.MustCompile(regular).MatchString(str)
}

// IsYMDHM 验证是否为 YYYY-MM-DD HH:mm 格式
func IsYMDHM(field validator.FieldLevel) bool {
	//(?!0000)  闰年:2016-02-29  15:04:00
	str := field.Field().String()
	regular := `^((((1[6-9]|[2-9]\d)\d{2})-(0?[13578]|1[02])-(0?[1-9]|[12]\d|3[01]))|(((1[6-9]|[2-9]\d)\d{2})-(0?[13456789]|1[012])-(0?[1-9]|[12]\d|30))|(((1[6-9]|[2-9]\d)\d{2})-0?2-(0?[1-9]|1\d|2[0-8]))|(((1[6-9]|[2-9]\d)(0[48]|[2468][048]|[13579][26])|((16|[2468][048]|[3579][26])00))-0?2-29-)) (20|21|22|23|[0-1]?\d):[0-5]?\d$`
	return regexp.MustCompile(regular).MatchString(str)
}

// IsYMDHMS 验证是否为 YYYY-MM-DD HH:mm:ss 格式
func IsYMDHMS(field validator.FieldLevel) bool {
	//(?!0000)  闰年:2016-02-29  15:04:00
	str := field.Field().String()
	regular := `^((((1[6-9]|[2-9]\d)\d{2})-(0?[13578]|1[02])-(0?[1-9]|[12]\d|3[01]))|(((1[6-9]|[2-9]\d)\d{2})-(0?[13456789]|1[012])-(0?[1-9]|[12]\d|30))|(((1[6-9]|[2-9]\d)\d{2})-0?2-(0?[1-9]|1\d|2[0-8]))|(((1[6-9]|[2-9]\d)(0[48]|[2468][048]|[13579][26])|((16|[2468][048]|[3579][26])00))-0?2-29-)) (20|21|22|23|[0-1]?\d):[0-5]?\d:[0-5]?\d$`
	return regexp.MustCompile(regular).MatchString(str)
}

// IsTest 通用正则验证
func IsTest(str string, reg string) bool {
	return regexp.MustCompile(reg).MatchString(str)
}
