package parse

import (
	"encoding/json"
)

// JSONUnmarshal JSON解析
func JSONUnmarshal(data []byte) (map[string]interface{}, error) {
	mapData := make(map[string]interface{})
	err := json.Unmarshal(data, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}
