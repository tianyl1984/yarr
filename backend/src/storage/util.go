package storage

func truncateStr(s string, length int) string {
	if s == "" || length <= 0 {
		return s
	}

	// 注意：len(s) 返回的是字节数，不是字符数（对于中文要特殊处理）
	runes := []rune(s) // 转换为 rune 切片以支持中文等多字节字符

	if len(runes) <= length {
		return s
	}

	return string(runes[:length])
}
