package kit

import (
	"strings"
	"unicode/utf8"

	"google.golang.org/protobuf/types/known/structpb"
)

// SensitiveFields 敏感字段列表
var SensitiveFields = map[string]bool{
	"api_key":               true,
	"personal_access_token": true,
	"access_token":          true,
	"token":                 true,
	"secret":                true,
	"access_key_secret":     true,
	"secret_key":            true,
}

// IsSensitiveField 检查字段是否为敏感字段
func IsSensitiveField(fieldName string) bool {
	if fieldName == "" {
		return false
	}
	return SensitiveFields[strings.ToLower(fieldName)]
}

// MaskMiddle 隐藏字符串中间部分
func MaskMiddle(value string) string {
	if value == "" {
		return value
	}

	// 确保字符串是有效的 UTF-8
	if !utf8.ValidString(value) {
		// 如果包含无效的 UTF-8 字符，清理它
		value = strings.ToValidUTF8(value, "")
		if value == "" {
			return "****"
		}
	}

	// 使用 rune 切片来正确处理 UTF-8 字符
	runes := []rune(value)
	runeLen := len(runes)

	if runeLen <= 1 {
		return "****"
	}

	if runeLen <= 8 {
		// 短字符串保留前2后2
		if runeLen <= 4 {
			return "****"
		}
		return string(runes[:2]) + "****" + string(runes[runeLen-2:])
	} else {
		// 长字符串保留前4后4
		maskLength := runeLen - 8
		mask := strings.Repeat("*", maskLength)
		return string(runes[:4]) + mask + string(runes[runeLen-4:])
	}
}

// IsMaskedValue 判断字符串是否是被掩码处理过的值
func IsMaskedValue(value string) bool {
	if value == "" {
		return false
	}
	// 掩码值至少包含4个连续的*
	return strings.Contains(value, "****")
}

// MaskSensitiveFields 处理Struct中的敏感字段
func MaskSensitiveFields(structValue *structpb.Struct) (*structpb.Struct, error) {
	if structValue == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	for key, value := range structValue.Fields {
		if IsSensitiveField(key) {
			// 如果是敏感字段且是字符串类型
			if strValue := value.GetStringValue(); strValue != "" {
				result[key] = MaskMiddle(strValue)
			} else {
				result[key] = value
			}
		} else if nestedStruct := value.GetStructValue(); nestedStruct != nil {
			// 递归处理嵌套的Struct
			maskedNested, err := MaskSensitiveFields(nestedStruct)
			if err != nil {
				return nil, err
			}
			if maskedNested != nil {
				result[key] = maskedNested
			} else {
				result[key] = value
			}
		} else {
			result[key] = value
		}
	}

	return structpb.NewStruct(result)
}

// sanitizeString 清理字符串中的无效 UTF-8 字符
func sanitizeString(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	// 替换无效的 UTF-8 字符
	return strings.ToValidUTF8(s, "")
}

// MaskSensitiveFieldsInMap 处理map中的敏感字段
func MaskSensitiveFieldsInMap(data map[string]interface{}) (map[string]interface{}, error) {
	if data == nil {
		return nil, nil
	}

	result := make(map[string]interface{})

	for key, value := range data {
		// 确保 key 是有效的 UTF-8
		key = sanitizeString(key)
		keyLower := strings.ToLower(key)

		if IsSensitiveField(keyLower) {
			// 如果是敏感字段且是字符串类型
			if strValue, ok := value.(string); ok && strValue != "" {
				result[key] = MaskMiddle(strValue)
			} else {
				// 对于非字符串值，也需要清理字符串字段
				result[key] = sanitizeValue(value)
			}
		} else if nestedMap, ok := value.(map[string]interface{}); ok {
			// 递归处理嵌套的map
			maskedNested, err := MaskSensitiveFieldsInMap(nestedMap)
			if err != nil {
				return nil, err
			}
			if maskedNested != nil {
				result[key] = maskedNested
			} else {
				result[key] = sanitizeValue(value)
			}
		} else {
			result[key] = sanitizeValue(value)
		}
	}

	return result, nil
}

// sanitizeValue 清理值中的无效 UTF-8 字符
func sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return sanitizeString(v)
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = sanitizeValue(item)
		}
		return result
	case map[string]interface{}:
		result := make(map[string]interface{})
		for k, val := range v {
			result[sanitizeString(k)] = sanitizeValue(val)
		}
		return result
	default:
		return value
	}
}
