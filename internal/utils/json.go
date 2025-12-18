package utils

import (
	"encoding/json"
	"fmt"
	"net/url"
)

func JSONToURLValues(jsonStr string) (url.Values, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &params); err != nil {
		return nil, err
	}

	return MapToURLValues(params), nil
}

func MapToURLValues(params map[string]interface{}) url.Values {
	values := url.Values{}
	for key, value := range params {
		switch v := value.(type) {
		case string:
			values.Add(key, v)
		case int, int64, float64, bool:
			values.Add(key, fmt.Sprintf("%v", v))
		case []interface{}:
			for _, item := range v {
				values.Add(key, fmt.Sprintf("%v", item))
			}
		case nil:
			continue
		default:
			values.Add(key, fmt.Sprintf("%v", v))
		}
	}
	return values
}
