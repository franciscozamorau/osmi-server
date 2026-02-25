package utils

import "strings"

// SafeStringForLog oculta datos sensibles
func SafeStringForLog(s string) string {
	if s == "" {
		return ""
	}
	if len(s) <= 2 {
		return "***"
	}
	return s[:2] + "***"
}

// SafeEmailForLog oculta email
func SafeEmailForLog(email string) string {
	if email == "" {
		return ""
	}
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***@***"
	}
	if len(parts[0]) <= 2 {
		return "***@" + parts[1]
	}
	return parts[0][:2] + "***@" + parts[1]
}

// Join es un wrapper para strings.Join
func Join(elems []string, sep string) string {
	return strings.Join(elems, sep)
}
