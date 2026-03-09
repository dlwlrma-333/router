package auth

import (
	"fmt"
	"strconv"
	"strings"
)

func sessionIDToString(value interface{}) (string, error) {
	switch v := value.(type) {
	case string:
		id := strings.TrimSpace(v)
		if id == "" {
			return "", fmt.Errorf("empty session user id")
		}
		return id, nil
	case int:
		if v <= 0 {
			return "", fmt.Errorf("invalid session user id")
		}
		return strconv.Itoa(v), nil
	case int64:
		if v <= 0 {
			return "", fmt.Errorf("invalid session user id")
		}
		return strconv.FormatInt(v, 10), nil
	case float64:
		if v <= 0 {
			return "", fmt.Errorf("invalid session user id")
		}
		return strconv.FormatInt(int64(v), 10), nil
	default:
		return "", fmt.Errorf("invalid session user id type: %T", value)
	}
}
