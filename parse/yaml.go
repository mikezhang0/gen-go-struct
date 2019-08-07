package parse

import yaml "gopkg.in/yaml.v2"

// YAMLUnmarshal YAML解析
func YAMLUnmarshal(data []byte) (map[string]interface{}, error) {
	mapData := make(map[string]interface{})
	err := yaml.Unmarshal(data, &mapData)
	if err != nil {
		return nil, err
	}
	return mapData, nil
}
