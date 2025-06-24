package excel

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
)

// ExcelType 导出类型 对应Java后端的Type枚举
type ExcelType int

const (
	TypeAll    ExcelType = 0 // 导出导入
	TypeExport ExcelType = 1 // 仅导出
	TypeImport ExcelType = 2 // 仅导入
)

// ColumnType 列类型 对应Java后端的ColumnType枚举
type ColumnType int

const (
	ColumnTypeNumeric ColumnType = 0 // 数字
	ColumnTypeString  ColumnType = 1 // 字符串
	ColumnTypeImage   ColumnType = 2 // 图片
	ColumnTypeText    ColumnType = 3 // 文本
)

// ExcelField Excel字段信息 对应Java后端的Field+Excel注解组合
type ExcelField struct {
	Field            reflect.StructField // 字段信息
	Sort             int                 // 排序
	Name             string              // 列名
	DateFormat       string              // 日期格式
	DictType         string              // 字典类型
	ReadConverterExp string              // 读取转换表达式
	Separator        string              // 分隔符
	Scale            int                 // BigDecimal精度
	RoundingMode     int                 // BigDecimal舍入规则
	Height           float64             // 行高
	Width            float64             // 列宽
	Suffix           string              // 后缀
	DefaultValue     string              // 默认值
	Prompt           string              // 提示信息
	WrapText         bool                // 是否换行
	Combo            []string            // 下拉选择
	ComboReadDict    bool                // 是否从字典读取下拉
	NeedMerge        bool                // 是否需要合并单元格
	IsExport         bool                // 是否导出数据
	TargetAttr       string              // 目标属性
	IsStatistics     bool                // 是否统计
	CellType         ColumnType          // 列类型
	Type             ExcelType           // 字段类型
}

// parseExcelTag 解析Excel标签 对应Java后端的@Excel注解解析
func parseExcelTag(field reflect.StructField) *ExcelField {
	tag := field.Tag.Get("excel")
	if tag == "" || tag == "-" {
		return nil
	}

	excelField := &ExcelField{
		Field:        field,
		Sort:         999999, // 默认排序值
		Name:         field.Name,
		Separator:    ",",
		Scale:        -1,
		Height:       14,
		Width:        16,
		IsExport:     true,
		CellType:     ColumnTypeString,
		Type:         TypeAll,
	}

	// 解析标签参数
	parts := strings.Split(tag, ";")
	for _, part := range parts {
		kv := strings.SplitN(part, ":", 2)
		if len(kv) != 2 {
			continue
		}
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])

		switch key {
		case "name":
			excelField.Name = value
		case "sort":
			if sort, err := strconv.Atoi(value); err == nil {
				excelField.Sort = sort
			}
		case "dateFormat":
			excelField.DateFormat = value
		case "dictType":
			excelField.DictType = value
		case "readConverterExp":
			excelField.ReadConverterExp = value
		case "separator":
			excelField.Separator = value
		case "scale":
			if scale, err := strconv.Atoi(value); err == nil {
				excelField.Scale = scale
			}
		case "height":
			if height, err := strconv.ParseFloat(value, 64); err == nil {
				excelField.Height = height
			}
		case "width":
			if width, err := strconv.ParseFloat(value, 64); err == nil {
				excelField.Width = width
			}
		case "suffix":
			excelField.Suffix = value
		case "defaultValue":
			excelField.DefaultValue = value
		case "prompt":
			excelField.Prompt = value
		case "wrapText":
			excelField.WrapText = value == "true"
		case "combo":
			if value != "" {
				excelField.Combo = strings.Split(value, ",")
			}
		case "comboReadDict":
			excelField.ComboReadDict = value == "true"
		case "needMerge":
			excelField.NeedMerge = value == "true"
		case "isExport":
			excelField.IsExport = value == "true"
		case "targetAttr":
			excelField.TargetAttr = value
		case "isStatistics":
			excelField.IsStatistics = value == "true"
		case "cellType":
			switch value {
			case "numeric":
				excelField.CellType = ColumnTypeNumeric
			case "string":
				excelField.CellType = ColumnTypeString
			case "image":
				excelField.CellType = ColumnTypeImage
			case "text":
				excelField.CellType = ColumnTypeText
			}
		case "type":
			switch value {
			case "all":
				excelField.Type = TypeAll
			case "export":
				excelField.Type = TypeExport
			case "import":
				excelField.Type = TypeImport
			}
		}
	}

	return excelField
}

// getExcelFields 获取结构体的Excel字段 对应Java后端的getFields方法
func getExcelFields(structType reflect.Type, excelType ExcelType) []*ExcelField {
	var fields []*ExcelField

	// 遍历所有字段（包括嵌入字段）
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// 跳过非导出字段
		if !field.IsExported() {
			continue
		}

		// 解析Excel标签
		excelField := parseExcelTag(field)
		if excelField == nil {
			continue
		}

		// 检查字段类型是否匹配
		if excelField.Type != TypeAll && excelField.Type != excelType {
			continue
		}

		fields = append(fields, excelField)
	}

	// 按排序字段排序
	sort.Slice(fields, func(i, j int) bool {
		return fields[i].Sort < fields[j].Sort
	})

	return fields
}

// convertByExp 根据表达式转换值 对应Java后端的convertByExp
func convertByExp(value, readConverterExp, separator string) string {
	if readConverterExp == "" {
		return value
	}

	if separator == "" {
		separator = ","
	}

	// 解析表达式 如: "0=男,1=女,2=未知"
	pairs := strings.Split(readConverterExp, separator)
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 && strings.TrimSpace(kv[0]) == value {
			return strings.TrimSpace(kv[1])
		}
	}
	return value
}

// reverseConvertByExp 根据表达式反向转换值 对应Java后端的reverseByExp
func reverseConvertByExp(value, readConverterExp, separator string) string {
	if readConverterExp == "" {
		return value
	}

	if separator == "" {
		separator = ","
	}

	// 解析表达式 如: "0=男,1=女,2=未知"
	pairs := strings.Split(readConverterExp, separator)
	for _, pair := range pairs {
		kv := strings.Split(pair, "=")
		if len(kv) == 2 && strings.TrimSpace(kv[1]) == value {
			return strings.TrimSpace(kv[0])
		}
	}
	return value
}
