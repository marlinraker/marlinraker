package executors

import (
	"fmt"
	"github.com/samber/lo"
	"strconv"
)

type Params map[string]any

func (params Params) GetString(name string) (string, bool) {
	value, exists := params[name]
	if !exists {
		return "", false
	}
	switch value := value.(type) {
	case string:
		return value, true
	}
	return fmt.Sprintf("%v", value), true
}

func (params Params) GetInt64(name string) (int64, bool) {
	value, exists := params[name]
	if !exists {
		return 0, false
	}
	switch value := value.(type) {
	case int64:
		return value, true
	case float64:
		return int64(value), true
	default:
		str := fmt.Sprintf("%v", value)
		if value, err := strconv.ParseInt(str, 10, 64); err == nil {
			return value, true
		}
	}
	return 0, false
}

func (params Params) GetBool(name string) (bool, bool) {
	value, exists := params[name]
	if !exists {
		return false, false
	}
	switch value := value.(type) {
	case bool:
		return value, true
	case string:
		return value == "true", true
	}
	return false, false
}

func (params Params) GetStringSlice(name string) ([]string, bool) {
	value, exists := params[name]
	if !exists {
		return nil, false
	}
	switch value := value.(type) {
	case []string:
		return value, true
	case []any:
		strings := lo.Map(value, func(item any, _ int) string { return fmt.Sprintf("%v", item) })
		return strings, true
	}
	return nil, false
}
