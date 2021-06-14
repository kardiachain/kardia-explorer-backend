package utils

import (
	"encoding/base64"
	"regexp"
	"strings"
)

func CheckBase64Logo(logo string) bool {
	if strings.Contains(logo, "data:image/jpeg;base64,") || strings.Contains(logo, "data:image/png;base64,") || strings.Contains(logo, "data:image/webp;base64,") {
		if _, err := base64.StdEncoding.DecodeString(strings.Split(logo, ",")[1]); err == nil {
			return true
		}
	} else if _, err := base64.StdEncoding.DecodeString(logo); err == nil {
		return true
	}
	return false
}

func IsValidAddress(v string) bool {
	re := regexp.MustCompile("^0x[0-9a-fA-F]{40}$")
	return re.MatchString(v)
}
