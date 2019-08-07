package parse

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
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
func ReadFile(filePath string) error {
	var ext ExtType
	switch filepath.Ext(filePath) {
	case ".json":
		ext = ExtTypeJSON
	case ".yml":
		ext = ExtTypeYAML
	case ".yaml":
		ext = ExtTypeYAML
	default:
		return fmt.Errorf("ext is not support")
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	fatherStructName := gostyle.FormatToCamelCase(filepath.Base(filepathplus.NoExt(filePath)))
	return Parse(data, fatherStructName, ext)
}

// Parse 解析数据
func Parse(data []byte, fatherStructName string, ext ExtType, addTags ...*gostyle.StructFieldTag) error {
	var mapData map[string]interface{}
	var err error

	switch ext {
	case ExtTypeJSON:
		mapData, err = JSONUnmarshal(data)
	case ExtTypeYAML:
		mapData, err = YAMLUnmarshal(data)
	default:
		err = fmt.Errorf("ext is not support:%s", ext)
	}
	if err != nil {
		return err
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
	return err
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

func (p *parseData) parseMapAsStruct(fatherStructName string, mapData map[string]interface{}) (*gostyle.StructDataBaseType, error) {
	// fields := []*gostyle.StructField{}
	// for key, value := range mapData {
	// 	camelName := gostyle.FormatToCamelCase(key)
	// 	tags := []*gostyle.StructFieldTag{
	// 		&gostyle.StructFieldTag{
	// 			Key:      string(p.ext),
	// 			Template: "%s,omitempty",
	// 		},
	// 	}

	// 	// 添加额外的tag
	// 	for _, tag1 := range p.addTags {
	// 		var isMatch bool
	// 		for _, tag2 := range tags {
	// 			if tag1.Key == tag1.Key {
	// 				isMatch = true
	// 				continue
	// 			}
	// 		}
	// 		if !isMatch {
	// 			tags = append(tags, &gostyle.StructFieldTag{
	// 				Key:      tag1.Key,
	// 				Template: tag1.Template,
	// 			})
	// 		}
	// 	}

	// 	// 开始初始化field
	// 	field := &gostyle.StructField{
	// 		Name: camelName,
	// 		Tags: gostyle.StructFieldTagList(tags),
	// 	}
	// 	sort.Sort(field.Tags)

	// 	// 处理dataType
	// 	var dataType string
	// 	isSlice := false
	// 	isMap := false

	// 	valueType := reflect.TypeOf(value)
	// 	if valueType.Kind() == reflect.Slice {
	// 		isSlice = true
	// 		value = value.([]interface{})[0]
	// 		valueType = reflect.TypeOf(value)
	// 	}

	// 	if valueType.Kind() == reflect.String {
	// 		dataType = "string"
	// 	} else if valueType.Kind() == reflect.Bool {
	// 		dataType = "bool"
	// 	} else if valueType.Kind() == reflect.Float64 || valueType.Kind() == reflect.Int {
	// 		dataType = "int"
	// 	} else if valueType.Kind() == reflect.Map {
	// 		isMap = true
	// 		formatMapData, ok := FormatMapData(value)
	// 		if !ok {
	// 			return fmt.Errorf("valueType is not support:map[interface{}]interface{}")
	// 		}
	// 		dataType = genStructName(structName, field.Name)
	// 		err = o.parseMap(format, dataType, formatMapData)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 	} else if valueType.Kind() == reflect.Slice {
	// 		return fmt.Errorf("valueType is not support:[][]interface{}")
	// 	} else {
	// 		return fmt.Errorf("valueType is not support:%v", valueType)
	// 	}

	// 	if isSlice && isMap {
	// 		field.DataType = fmt.Sprintf("[]*%s", dataType)
	// 	} else if isSlice {
	// 		field.DataType = fmt.Sprintf("[]%s", dataType)
	// 	} else if isMap {
	// 		field.DataType = fmt.Sprintf("*%s", dataType)
	// 	} else {
	// 		field.DataType = dataType
	// 	}
	// 	fields = append(fields, field)
	// }

	// sort.Sort(SortObjField(fields))

	// o.Structs = append(o.Structs, &ObjStruct{Name: structName, Fields: fields})
	return nil, nil
}

func (p *parseData) parseValue(fatherStructName string, value interface{}) (gostyle.DataType, error) {
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
		dataBaseType, err = p.parseMapAsStruct(fatherStructName, value.(map[string]interface{}))
		if err != nil {
			return nil, err
		}
	case reflect.String:
		dataBaseType, _ = gostyle.GetGeneralDataBaseType(reflect.String)
	case reflect.Bool:
		dataBaseType, _ = gostyle.GetGeneralDataBaseType(reflect.Bool)
	case reflect.Int:
		dataBaseType, _ = gostyle.GetGeneralDataBaseType(reflect.Int)
	case reflect.Float64:
		if p.ext == ExtTypeJSON {
			dataBaseType, _ = gostyle.GetGeneralDataBaseType(reflect.Int)
		} else {
			dataBaseType, _ = gostyle.GetGeneralDataBaseType(reflect.Float64)
		}
	default:
		return nil, fmt.Errorf("valueType is not support:%v", valueType.Kind())
	}

	var dataType gostyle.DataType
	switch structKind {
	case reflect.Slice:
		dataType, err = gostyle.GetSliceDataType(dataBaseType, 0)
	case reflect.Map:
		k, _ := gostyle.GetGeneralDataBaseType(reflect.String)
		dataType, err = gostyle.GetMapDataType(k, dataBaseType)
	default:
		dataType, err = gostyle.GetGeneralDataType(dataBaseType)
	}
	if err != nil {
		return nil, err
	}
	return dataType, nil
}

// // NewObjConfigFile 配置文件转objConfigFile
// func NewObjConfigFile(configName, packageName, structName, configPath string) (*ObjConfigFile, error) {
// 	format, mapData, err := parse(configPath)
// 	if err != nil {
// 		return nil, err
// 	}

// 	objConfigFile := &ObjConfigFile{
// 		Name:        configName,
// 		StructName:  structName,
// 		PackageName: packageName,
// 		Format:      format,
// 	}
// 	err = objConfigFile.parseMap(format, structName, mapData)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Structs 排序
// 	structLength := len(objConfigFile.Structs)
// 	newStructs := []*ObjStruct{objConfigFile.Structs[structLength-1]}
// 	tmpStructs := objConfigFile.Structs[0 : structLength-1]
// 	sort.Sort(SortObjStruct(tmpStructs))
// 	newStructs = append(newStructs, tmpStructs...)
// 	objConfigFile.Structs = newStructs
// 	return objConfigFile, nil
// }
