package jsontime

import (
	"database/sql/driver"
	"fmt"
	"strings"
	"time"
)

// JSONTime 自定义时间类型，支持JSON格式化 对应Java后端@JsonFormat注解
type JSONTime struct {
	time.Time
}

// 时间格式常量 对应Java后端@JsonFormat(pattern = "yyyy-MM-dd HH:mm:ss")
const (
	TimeFormat = "2006-01-02 15:04:05"
	DateFormat = "2006-01-02"
)

// MarshalJSON 实现JSON序列化 对应Java后端@JsonFormat注解的功能
func (jt JSONTime) MarshalJSON() ([]byte, error) {
	if jt.Time.IsZero() {
		return []byte("null"), nil
	}
	formatted := fmt.Sprintf(`"%s"`, jt.Time.Format(TimeFormat))
	return []byte(formatted), nil
}

// UnmarshalJSON 实现JSON反序列化 对应Java后端@JsonFormat注解的功能
func (jt *JSONTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	if str == "null" || str == "" {
		jt.Time = time.Time{}
		return nil
	}

	// 尝试多种时间格式
	formats := []string{
		TimeFormat,
		DateFormat,
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05.000Z",
		"2006-01-02 15:04:05.000",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			jt.Time = t
			return nil
		}
	}

	return fmt.Errorf("无法解析时间格式: %s", str)
}

// Value 实现driver.Valuer接口，用于数据库存储
func (jt JSONTime) Value() (driver.Value, error) {
	if jt.Time.IsZero() {
		return nil, nil
	}
	return jt.Time, nil
}

// Scan 实现sql.Scanner接口，用于数据库读取
func (jt *JSONTime) Scan(value interface{}) error {
	if value == nil {
		jt.Time = time.Time{}
		return nil
	}

	switch v := value.(type) {
	case time.Time:
		jt.Time = v
	case string:
		t, err := time.Parse(TimeFormat, v)
		if err != nil {
			return err
		}
		jt.Time = t
	default:
		return fmt.Errorf("无法将 %T 转换为 JSONTime", value)
	}

	return nil
}

// String 实现Stringer接口
func (jt JSONTime) String() string {
	if jt.Time.IsZero() {
		return ""
	}
	return jt.Time.Format(TimeFormat)
}

// IsZero 检查时间是否为零值
func (jt JSONTime) IsZero() bool {
	return jt.Time.IsZero()
}

// Now 获取当前时间的JSONTime
func Now() JSONTime {
	return JSONTime{Time: time.Now()}
}

// NewJSONTime 创建新的JSONTime实例
func NewJSONTime(t time.Time) JSONTime {
	return JSONTime{Time: t}
}

// ParseJSONTime 解析字符串为JSONTime
func ParseJSONTime(str string) (JSONTime, error) {
	if str == "" {
		return JSONTime{}, nil
	}

	t, err := time.Parse(TimeFormat, str)
	if err != nil {
		return JSONTime{}, err
	}

	return JSONTime{Time: t}, nil
}

// FormatTime 格式化time.Time为字符串 对应Java后端@JsonFormat的格式化功能
func FormatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(TimeFormat)
}

// FormatDate 格式化time.Time为日期字符串
func FormatDate(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(DateFormat)
}

// ParseTime 解析字符串为time.Time指针
func ParseTime(str string) (*time.Time, error) {
	if str == "" {
		return nil, nil
	}

	t, err := time.Parse(TimeFormat, str)
	if err != nil {
		return nil, err
	}

	return &t, nil
}

// TimeToJSONTime 将*time.Time转换为JSONTime
func TimeToJSONTime(t *time.Time) JSONTime {
	if t == nil {
		return JSONTime{}
	}
	return JSONTime{Time: *t}
}

// JSONTimeToTime 将JSONTime转换为*time.Time
func JSONTimeToTime(jt JSONTime) *time.Time {
	if jt.IsZero() {
		return nil
	}
	return &jt.Time
}
