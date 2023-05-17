package executors

import (
	"fmt"
	"github.com/samber/lo"
	"marlinraker/src/util"
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

func (params Params) RequireString(name string) (string, error) {
	value, exists := params.GetString(name)
	if !exists {
		return "", util.NewError(name+" param is required", 400)
	}
	return value, nil
}

func (params Params) RequirePath(name string) (string, error) {
	value, err := params.RequireString(name)
	if err != nil {
		return "", err
	}
	return util.SanitizePath(value), err
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

func (params Params) RequireInt64(name string) (int64, error) {
	value, exists := params.GetInt64(name)
	if !exists {
		return 0, util.NewError(name+" param is required", 400)
	}
	return value, nil
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

func (params Params) RequireBool(name string) (bool, error) {
	value, exists := params.GetBool(name)
	if !exists {
		return false, util.NewError(name+" param is required", 400)
	}
	return value, nil
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

func (params Params) RequireStringSlice(name string) ([]string, error) {
	value, exists := params.GetStringSlice(name)
	if !exists {
		return nil, util.NewError(name+" param is required", 400)
	}
	return value, nil
}

func (params Params) RequireAny(name string) (any, error) {
	value, exists := params[name]
	if !exists {
		return nil, util.NewError(name+" param is required", 400)
	}
	return value, nil
}
