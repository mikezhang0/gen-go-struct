package parse

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"strings"

	"github.com/mikezhang666/goplus/gostyle"
	filepathplus "github.com/mikezhang666/goplus/path/filepath"
)

// ExtType 扩展类型
type ExtType string

// ExtType常量
const (
	ExtTypeJSON = "json"
	ExtTypeYAML = "yaml"
)

// ReadFile 解析文件
func ReadFile(filePath string, addTags ...*gostyle.StructFieldTag) (string, error) {
	var ext ExtType
	switch filepath.Ext(filePath) {
	case ".json":
		ext = ExtTypeJSON
	case ".yml":
		ext = ExtTypeYAML
	case ".yaml":
		ext = ExtTypeYAML
	default:
		panic(fmt.Sprintf("ext unmarshal is not support:%s", ext))
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	name := gostyle.FormatToCamelCase(filepath.Base(filepathplus.NoExt(filePath)))
	return Parse(data, name, ext, addTags...)
}

// Parse 解析数据
func Parse(data []byte, fatherStructName string, ext ExtType, addTags ...*gostyle.StructFieldTag) (string, error) {
	var mapData map[string]interface{}
	var err error

	switch ext {
	case ExtTypeJSON:
		mapData, err = JSONUnmarshal(data)
	case ExtTypeYAML:
		mapData, err = YAMLUnmarshal(data)
	default:
		panic(fmt.Sprintf("ext unmarshal is not support:%s", ext))
	}
	if err != nil {
		return "", err
	}

	parseData := &parseData{
		ext:     ext,
		addTags: []*gostyle.StructFieldTag{},
		structs: []*gostyle.Struct{},
	}
	if addTags != nil {
		parseData.addTags = addTags
	}
	_, err = parseData.parseMapAsStruct(fatherStructName, mapData)
	return parseData.string(), err
}

func getTags(value string, ext ExtType, addTags []*gostyle.StructFieldTag) []*gostyle.StructFieldTag {
	var extDefaultTag *gostyle.StructFieldTag
	switch ext {
	case ExtTypeJSON:
		extDefaultTag = gostyle.GetStructFieldTag("json", "%s,omitempty")
	case ExtTypeYAML:
		extDefaultTag = gostyle.GetStructFieldTag("yaml", "%s,omitempty")
	default:
		panic(fmt.Sprintf("ext default tag is not support:%s", ext))
	}

	extDefaultTag.SetValue(value)

	tags := []*gostyle.StructFieldTag{extDefaultTag}
	for _, t := range addTags {
		t.SetValue(value)
		if t.GetKey() != extDefaultTag.GetKey() {
			tags = append(tags, t)
		}
	}
	return tags
}

// 获得子结构体的名称
func genStructName(fatherStructName, fieldName string) string {
	return fmt.Sprintf("%s%s", fatherStructName, fieldName)
}

// map 是否作为 struct
func isMapAsStruct(mapData map[string]interface{}) bool {
	for k := range mapData {
		if !strings.HasPrefix(k, "uqi_") { // 语法糖 包含该前缀 作为 map 结构
			return false
		}
	}
	return true
}

// parseData 解析结构
type parseData struct {
	ext     ExtType
	addTags []*gostyle.StructFieldTag // 额外的tag
	structs []*gostyle.Struct
}

func (p *parseData) parseMapAsStruct(name string, mapData map[string]interface{}) (gostyle.DataBaseType, error) {
	fields := []*gostyle.StructField{}
	for fieldTagKey, fieldData := range mapData {
		fieldName := gostyle.FormatToCamelCase(fieldTagKey)
		fieldDataType, err := p.parseValue(genStructName(name, fieldName), fieldData)
		if err != nil {
			return nil, err
		}
		fieldTags := getTags(fieldTagKey, p.ext, p.addTags)
		fields = append(fields, gostyle.GetStructField(fieldName, fieldDataType, fieldTags...))
	}

	sort.Sort(gostyle.StructFieldList(fields))
	p.structs = append(p.structs, gostyle.GetStruct(name, name, fields...))
	return gostyle.GetStructDataBaseType(name, true), nil
}

func (p *parseData) parseValue(name string, value interface{}) (gostyle.DataType, error) {
	var err error

	valueType := reflect.TypeOf(value)

	// 判断是 slice, map 还是 无结构
	var structKind reflect.Kind
	if valueType.Kind() == reflect.Slice {
		structKind = reflect.Slice
		value = value.([]interface{})[0] // slice 以第一项为准
		valueType = reflect.TypeOf(value)
	} else if valueType.Kind() == reflect.Map {
		if _, ok := value.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("valueType is not support:map[interface{}]interface{}")
		}
		if !isMapAsStruct(value.(map[string]interface{})) {
			structKind = reflect.Map
			for _, v := range value.(map[string]interface{}) {
				value = v // map 以任意项为准
			}
			valueType = reflect.TypeOf(value)
		}
	}

	// 判断基础数据结构
	var dataBaseType gostyle.DataBaseType
	switch valueType.Kind() {
	case reflect.Slice: // 不支持 [][]interface{}
		return nil, fmt.Errorf("valueType is not support:[][]interface{}")
	case reflect.Map:
		if _, ok := value.(map[string]interface{}); !ok {
			return nil, fmt.Errorf("valueType is not support:map[interface{}]interface{}")
		}
		// map 中的元素 一律作为 struct
		dataBaseType, err = p.parseMapAsStruct(name, value.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	case reflect.String:
		dataBaseType = gostyle.GetGeneralDataBaseType(reflect.String)
	case reflect.Bool:
		dataBaseType = gostyle.GetGeneralDataBaseType(reflect.Bool)
	case reflect.Int:
		dataBaseType = gostyle.GetGeneralDataBaseType(reflect.Int)
	case reflect.Float64:
		if p.ext == ExtTypeJSON {
			dataBaseType = gostyle.GetGeneralDataBaseType(reflect.Int)
		} else {
			dataBaseType = gostyle.GetGeneralDataBaseType(reflect.Float64)
		}
	default:
		return nil, fmt.Errorf("valueType is not support:%v", valueType.Kind())
	}

	var dataType gostyle.DataType
	switch structKind {
	case reflect.Slice:
		dataType = gostyle.GetSliceDataType(dataBaseType, 0)
	case reflect.Map:
		dataType = gostyle.GetMapDataType(gostyle.GetGeneralDataBaseType(reflect.String), dataBaseType)
	default:
		dataType = gostyle.GetGeneralDataType(dataBaseType)
	}
	return dataType, nil
}

// String 格式化
func (p *parseData) string() string {
	// Structs 排序
	structLength := len(p.structs)
	newStructs := []*gostyle.Struct{p.structs[structLength-1]}
	tempStructs := p.structs[:structLength-1]
	sort.Sort(gostyle.StructList(tempStructs))
	newStructs = append(newStructs, tempStructs...)
	// 格式化
	splits := []string{}
	for _, s := range newStructs {
		splits = append(splits, s.String())
	}
	return strings.Join(splits, "\n")
}
