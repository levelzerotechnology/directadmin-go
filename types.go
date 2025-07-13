package directadmin

import (
	"strconv"
	"strings"
)

const Unlimited int = -1

func parseNum(value string) int {
	if value == "" || value == "unlimited" {
		return -1
	}

	result, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return result
}

func parseOnOff(value string) bool {
	value = strings.ToUpper(value)

	switch value {
	case "NO", "OFF":
		return false
	case "ON", "YES":
		return true
	}

	return false
}

func reverseParseNum(value int, empty bool) string {
	if value == -1 {
		if empty {
			return ""
		}
		return "unlimited"
	}

	return strconv.Itoa(value)
}

func reverseParseOnOff(value bool) string {
	var result string

	if value {
		result = "ON"
	} else {
		result = "OFF"
	}

	return result
}

func reverseParseYesNo(value bool) string {
	var result string

	if value {
		result = "YES"
	} else {
		result = "NO"
	}

	return result
}
