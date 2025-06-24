package excel

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/xuri/excelize/v2"
)

// ExcelUtil Excel工具类 对应Java后端的ExcelUtil<T>
type ExcelUtil struct {
	file      *excelize.File
	sheetName string
	title     string
	excelType ExcelType
	list      interface{}
	fields    []*ExcelField
	rownum    int
	maxHeight float64
	styles    map[string]int
}

// NewExcelUtil 创建Excel工具实例 对应Java后端的new ExcelUtil(Class.class)
func NewExcelUtil() *ExcelUtil {
	return &ExcelUtil{
		file:      excelize.NewFile(),
		sheetName: "Sheet1",
		rownum:    0,
		maxHeight: 15,
		styles:    make(map[string]int),
	}
}

// ExportExcel 导出Excel 对应Java后端的exportExcel方法
func (e *ExcelUtil) ExportExcel(list interface{}, sheetName, title string) ([]byte, error) {
	return e.exportExcel(list, sheetName, title, TypeExport)
}

// exportExcel 内部导出方法
func (e *ExcelUtil) exportExcel(list interface{}, sheetName, title string, excelType ExcelType) ([]byte, error) {
	// 初始化
	if err := e.init(list, sheetName, title, excelType); err != nil {
		return nil, fmt.Errorf("初始化失败: %v", err)
	}

	// 创建表头
	if err := e.createHead(); err != nil {
		return nil, fmt.Errorf("创建表头失败: %v", err)
	}

	// 填充数据
	if err := e.fillExcelData(); err != nil {
		return nil, fmt.Errorf("填充数据失败: %v", err)
	}

	// 设置列宽
	e.setColumnWidth()

	// 生成文件
	buffer, err := e.file.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("生成Excel文件失败: %v", err)
	}

	return buffer.Bytes(), nil
}

// init 初始化 对应Java后端的init方法
func (e *ExcelUtil) init(list interface{}, sheetName, title string, excelType ExcelType) error {
	e.list = list
	e.sheetName = sheetName
	e.title = title
	e.excelType = excelType

	// 获取数据类型
	listValue := reflect.ValueOf(list)
	if listValue.Kind() != reflect.Slice {
		return fmt.Errorf("list必须是切片类型")
	}

	// 获取元素类型
	elemType := listValue.Type().Elem()
	if elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	// 创建Excel字段配置
	e.createExcelField(elemType)

	// 创建工作簿
	e.createWorkbook()

	// 创建标题
	e.createTitle()

	return nil
}

// createExcelField 创建Excel字段配置 对应Java后端的createExcelField
func (e *ExcelUtil) createExcelField(structType reflect.Type) {
	e.fields = getExcelFields(structType, e.excelType)
	e.maxHeight = e.getRowHeight()
}

// createWorkbook 创建工作簿 对应Java后端的createWorkbook
func (e *ExcelUtil) createWorkbook() {
	// 设置工作表名称
	e.file.SetSheetName("Sheet1", e.sheetName)
}

// createTitle 创建标题 对应Java后端的createTitle
func (e *ExcelUtil) createTitle() {
	if e.title == "" {
		return
	}

	// 创建标题行
	titleRow := e.rownum
	e.rownum++

	// 设置标题
	cell := fmt.Sprintf("A%d", titleRow+1)
	e.file.SetCellValue(e.sheetName, cell, e.title)

	// 合并单元格
	if len(e.fields) > 1 {
		endCell := fmt.Sprintf("%s%d", getColumnName(len(e.fields)-1), titleRow+1)
		e.file.MergeCell(e.sheetName, cell, endCell)
	}

	// 设置标题样式
	style, _ := e.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
			Size: 16,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
	})
	e.file.SetCellStyle(e.sheetName, cell, cell, style)
}

// createHead 创建表头 对应Java后端的createHead
func (e *ExcelUtil) createHead() error {
	headerRow := e.rownum
	e.rownum++

	// 创建表头样式
	headerStyle, err := e.file.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold: true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D3D3D3"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// 设置表头
	for i, field := range e.fields {
		cell := fmt.Sprintf("%s%d", getColumnName(i), headerRow+1)
		e.file.SetCellValue(e.sheetName, cell, field.Name)
		e.file.SetCellStyle(e.sheetName, cell, cell, headerStyle)
	}

	return nil
}

// fillExcelData 填充Excel数据 对应Java后端的fillExcelData
func (e *ExcelUtil) fillExcelData() error {
	listValue := reflect.ValueOf(e.list)
	if listValue.Kind() != reflect.Slice {
		return fmt.Errorf("list必须是切片类型")
	}

	// 创建数据样式
	dataStyle, err := e.file.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return err
	}

	// 填充数据行
	for i := 0; i < listValue.Len(); i++ {
		item := listValue.Index(i)
		if item.Kind() == reflect.Ptr {
			if item.IsNil() {
				continue
			}
			item = item.Elem()
		}

		dataRow := e.rownum
		e.rownum++

		// 填充每个字段
		for j, field := range e.fields {
			cell := fmt.Sprintf("%s%d", getColumnName(j), dataRow+1)
			value := e.getFieldValue(item, field)
			e.file.SetCellValue(e.sheetName, cell, value)
			e.file.SetCellStyle(e.sheetName, cell, cell, dataStyle)
		}
	}

	return nil
}

// getFieldValue 获取字段值 对应Java后端的getTargetValue
func (e *ExcelUtil) getFieldValue(item reflect.Value, field *ExcelField) interface{} {
	// 获取字段值
	fieldValue := item.FieldByName(field.Field.Name)
	if !fieldValue.IsValid() {
		return field.DefaultValue
	}

	// 处理指针类型
	if fieldValue.Kind() == reflect.Ptr {
		if fieldValue.IsNil() {
			return field.DefaultValue
		}
		fieldValue = fieldValue.Elem()
	}

	// 转换为字符串
	var strValue string
	switch fieldValue.Kind() {
	case reflect.String:
		strValue = fieldValue.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		strValue = strconv.FormatInt(fieldValue.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		strValue = strconv.FormatUint(fieldValue.Uint(), 10)
	case reflect.Float32, reflect.Float64:
		strValue = strconv.FormatFloat(fieldValue.Float(), 'f', -1, 64)
	case reflect.Bool:
		if fieldValue.Bool() {
			strValue = "1"
		} else {
			strValue = "0"
		}
	case reflect.Struct:
		if fieldValue.Type() == reflect.TypeOf(time.Time{}) {
			t := fieldValue.Interface().(time.Time)
			if field.DateFormat != "" {
				// 转换Java日期格式为Go格式
				goFormat := convertDateFormat(field.DateFormat)
				strValue = t.Format(goFormat)
			} else {
				strValue = t.Format("2006-01-02 15:04:05")
			}
		} else {
			strValue = fmt.Sprintf("%v", fieldValue.Interface())
		}
	default:
		strValue = fmt.Sprintf("%v", fieldValue.Interface())
	}

	// 应用转换表达式
	if field.ReadConverterExp != "" {
		strValue = convertByExp(strValue, field.ReadConverterExp, field.Separator)
	}

	// 添加后缀
	if field.Suffix != "" {
		strValue += field.Suffix
	}

	return strValue
}

// setColumnWidth 设置列宽
func (e *ExcelUtil) setColumnWidth() {
	for i, field := range e.fields {
		colName := getColumnName(i)
		width := field.Width
		if width <= 0 {
			width = 16 // 默认宽度
		}
		e.file.SetColWidth(e.sheetName, colName, colName, width)
	}
}

// getRowHeight 获取行高
func (e *ExcelUtil) getRowHeight() float64 {
	maxHeight := 15.0
	for _, field := range e.fields {
		if field.Height > maxHeight {
			maxHeight = field.Height
		}
	}
	return maxHeight
}

// getColumnName 获取列名 (A, B, C, ..., Z, AA, AB, ...)
func getColumnName(index int) string {
	result := ""
	for index >= 0 {
		result = string(rune('A'+index%26)) + result
		index = index/26 - 1
	}
	return result
}

// convertDateFormat 转换Java日期格式为Go格式
func convertDateFormat(javaFormat string) string {
	// Java -> Go 日期格式转换
	goFormat := javaFormat
	goFormat = strings.ReplaceAll(goFormat, "yyyy", "2006")
	goFormat = strings.ReplaceAll(goFormat, "MM", "01")
	goFormat = strings.ReplaceAll(goFormat, "dd", "02")
	goFormat = strings.ReplaceAll(goFormat, "HH", "15")
	goFormat = strings.ReplaceAll(goFormat, "mm", "04")
	goFormat = strings.ReplaceAll(goFormat, "ss", "05")
	return goFormat
}
