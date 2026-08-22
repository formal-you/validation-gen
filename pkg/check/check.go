// Package check 提供生成代码依赖的纯校验 helper。
//
// 这些函数无副作用、无反射，语义与固定版本 go-playground/validator/v10（v10.30.1）
// 的内置规则保持一致，保证静态生成代码与 runtime 校验结果一致。
package check

import (
	"net/mail"
	"regexp"
	"strconv"
	"unicode/utf8"
)

// Runes 返回字符串的 UTF-8 字符数。
// validator/v10 的 min/max/len/gt/gte/lt/lte 对 string 均按字符数比较，
// 而不是字节数，因此生成代码统一通过该函数计算长度。
func Runes(s string) int {
	return utf8.RuneCountInString(s)
}

// emailRegex 与 validator/v10 v10.30.1 regexes.go 中的 emailRegexString 完全一致。
// 注意：该正则要求域名部分必须包含点分段（TLD），因此 "a@b" 这类地址不通过。
var emailRegex = regexp.MustCompile("^(?:(?:(?:(?:[a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+(?:\\.([a-zA-Z]|\\d|[!#\\$%&'\\*\\+\\-\\/=\\?\\^_`{\\|}~]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])+)*)|(?:(?:\\x22)(?:(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(?:\\x20|\\x09)+)?(?:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x7f]|\\x21|[\\x23-\\x5b]|[\\x5d-\\x7e]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[\\x01-\\x09\\x0b\\x0c\\x0d-\\x7f]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}]))))*(?:(?:(?:\\x20|\\x09)*(?:\\x0d\\x0a))?(\\x20|\\x09)+)?(?:\\x22))))@(?:(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|\\d|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.)+(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])|(?:(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])(?:[a-zA-Z]|\\d|-|\\.|~|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])*(?:[a-zA-Z]|[\\x{00A0}-\\x{D7FF}\\x{F900}-\\x{FDCF}\\x{FDF0}-\\x{FFEF}])))\\.?$")

// Email 判断字符串是否为合法 email 地址。
// 与 validator/v10 的 isEmail 一致：必须同时通过 net/mail.ParseAddress 与 emailRegex。
func Email(s string) bool {
	if _, err := mail.ParseAddress(s); err != nil {
		return false
	}
	return emailRegex.MatchString(s)
}

// OneOfString 判断字符串是否命中候选项之一（精确比较）。
// 与 validator/v10 的 isOneOf 对 string 的语义一致。
func OneOfString(s string, vals ...string) bool {
	for _, v := range vals {
		if v == s {
			return true
		}
	}
	return false
}

// OneOfInt 判断有符号整数的十进制表示是否命中候选项之一。
// 与 validator/v10 的 isOneOf 一致：按十进制字符串比较，而不是数值比较，
// 因此候选项 "01" 与数值 1 不匹配。
func OneOfInt(v int64, vals ...string) bool {
	s := strconv.FormatInt(v, 10)
	for _, val := range vals {
		if val == s {
			return true
		}
	}
	return false
}

// OneOfUint 判断无符号整数的十进制表示是否命中候选项之一，语义同 OneOfInt。
func OneOfUint(v uint64, vals ...string) bool {
	s := strconv.FormatUint(v, 10)
	for _, val := range vals {
		if val == s {
			return true
		}
	}
	return false
}
