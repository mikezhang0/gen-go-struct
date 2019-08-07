package parse

// FormatMapData 格式化mapdata
func FormatMapData(mapData interface{}) (map[string]interface{}, bool) {
	if _, ok := mapData.(map[string]interface{}); ok {
		return mapData.(map[string]interface{}), true
	}
	if _, ok := mapData.(map[interface{}]interface{}); ok {
		data := make(map[string]interface{})
		for k, v := range mapData.(map[interface{}]interface{}) {
			data[k.(string)] = v
		}
		return data, true
	}
	return nil, false
}
