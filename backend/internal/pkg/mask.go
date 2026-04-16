package pkg

import "unicode/utf8"

// MaskPhone 对手机号脱敏：保留前3后4，中间用 **** 替换
// 输入为空或长度不足7位时原样返回
func MaskPhone(phone string) string {
	length := utf8.RuneCountInString(phone)
	if length < 7 || phone == "" {
		return phone
	}
	runes := []rune(phone)
	return string(runes[:3]) + "****" + string(runes[length-4:])
}
