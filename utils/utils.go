package utils

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func GetEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func TrimAll(str string) string {
	str = strings.Replace(str, "\n", "", -1)
	str = strings.ReplaceAll(str, "\u200F", "") // 去除Unicode的右向左标记（RLM）
	str = strings.ReplaceAll(str, "\u200E", "") // 去除Unicode的左向右标记（LRM）
	str = strings.TrimSpace(str)
	return str
}

func TrimSpan(str string) string {
	str = strings.ReplaceAll(str, ":", "")
	str = strings.Replace(str, "\n", "", -1)
	str = strings.ReplaceAll(str, "\u200F", "") // 去除Unicode的右向左标记（RLM）
	str = strings.ReplaceAll(str, "\u200E", "") // 去除Unicode的左向右标记（LRM）
	str = strings.TrimSpace(str)

	return str
}

func ExtractNumberFromString(s string) (float64, error) {
	re := regexp.MustCompile(`([\d,\.]+)`)
	matches := re.FindStringSubmatch(s)
	if len(matches) < 2 {
		return 0, fmt.Errorf("no number found in string")
	}
	numString := matches[1]
	numString = strings.ReplaceAll(numString, ",", "")
	return strconv.ParseFloat(numString, 64)
}
