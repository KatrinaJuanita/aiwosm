package model

import (
	"fmt"
	"math"
	"math/big"
	"time"
)

// Server 服务器相关信息 对应Java后端的Server实体
type Server struct {
	CPU      *CPU      `json:"cpu"`      // CPU相关信息
	Mem      *Mem      `json:"mem"`      // 内存相关信息
	JVM      *JVM      `json:"jvm"`      // JVM相关信息（Go中为运行时信息）
	Sys      *Sys      `json:"sys"`      // 服务器相关信息
	SysFiles []SysFile `json:"sysFiles"` // 磁盘相关信息
}

// CPU CPU相关信息 对应Java后端的Cpu实体
type CPU struct {
	CPUNum int     `json:"cpuNum"` // 核心数
	Total  float64 `json:"total"`  // CPU总的使用率
	Sys    float64 `json:"sys"`    // CPU系统使用率
	Used   float64 `json:"used"`   // CPU用户使用率
	Wait   float64 `json:"wait"`   // CPU当前等待率
	Free   float64 `json:"free"`   // CPU当前空闲率
}

// Mem 内存相关信息 对应Java后端的Mem实体
type Mem struct {
	Total float64 `json:"total"` // 内存总量
	Used  float64 `json:"used"`  // 已用内存
	Free  float64 `json:"free"`  // 剩余内存
	Usage float64 `json:"usage"` // 使用率
}

// JVM JVM相关信息 对应Java后端的Jvm实体（Go中为运行时信息）
type JVM struct {
	Total     float64 `json:"total"`     // 当前运行时占用的内存总数(M)
	Max       float64 `json:"max"`       // 运行时最大可用内存总数(M)
	Free      float64 `json:"free"`      // 运行时空闲内存(M)
	Used      float64 `json:"used"`      // 运行时已用内存(M)
	Usage     float64 `json:"usage"`     // 使用率
	Version   string  `json:"version"`   // Go版本
	Home      string  `json:"home"`      // Go安装路径
	Name      string  `json:"name"`      // 运行时名称
	StartTime string  `json:"startTime"` // 启动时间
	RunTime   string  `json:"runTime"`   // 运行时长
}

// Sys 系统相关信息 对应Java后端的Sys实体
type Sys struct {
	ComputerName string `json:"computerName"` // 服务器名称
	ComputerIP   string `json:"computerIp"`   // 服务器IP
	UserDir      string `json:"userDir"`      // 项目路径
	OSName       string `json:"osName"`       // 操作系统
	OSArch       string `json:"osArch"`       // 系统架构
}

// SysFile 系统文件相关信息 对应Java后端的SysFile实体
type SysFile struct {
	DirName     string  `json:"dirName"`     // 盘符路径
	SysTypeName string  `json:"sysTypeName"` // 盘符类型
	TypeName    string  `json:"typeName"`    // 文件类型
	Total       string  `json:"total"`       // 总大小
	Free        string  `json:"free"`        // 剩余大小
	Used        string  `json:"used"`        // 已经使用量
	Usage       float64 `json:"usage"`       // 资源的使用率
}

// 工具函数

// Round 四舍五入保留指定小数位 对应Java后端的Arith.round方法
func Round(val float64, precision int) float64 {
	if precision < 0 {
		panic("The scale must be a positive integer or zero")
	}
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}

// ArithMul 精确乘法运算 对应Java后端的Arith.mul方法
func ArithMul(v1, v2 float64) float64 {
	b1 := big.NewFloat(v1)
	b2 := big.NewFloat(v2)
	result := new(big.Float).Mul(b1, b2)
	val, _ := result.Float64()
	return val
}

// ArithDiv 精确除法运算 对应Java后端的Arith.div方法
func ArithDiv(v1, v2 float64, scale int) float64 {
	if scale < 0 {
		panic("The scale must be a positive integer or zero")
	}
	if v2 == 0 {
		return 0
	}
	b1 := big.NewFloat(v1)
	b2 := big.NewFloat(v2)
	result := new(big.Float).Quo(b1, b2)
	val, _ := result.Float64()
	return Round(val, scale)
}

// ArithRound 精确四舍五入 对应Java后端的Arith.round方法
func ArithRound(val float64, scale int) float64 {
	if scale < 0 {
		panic("The scale must be a positive integer or zero")
	}
	return Round(val, scale)
}

// FormatBytes 格式化字节数为可读格式
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatBytesToGB 格式化字节数为GB
func FormatBytesToGB(bytes uint64) float64 {
	return Round(float64(bytes)/(1024*1024*1024), 2)
}

// FormatBytesToMB 格式化字节数为MB
func FormatBytesToMB(bytes uint64) float64 {
	return Round(float64(bytes)/(1024*1024), 2)
}

// CalculateUsage 计算使用率百分比
func CalculateUsage(used, total float64) float64 {
	if total == 0 {
		return 0
	}
	return Round((used/total)*100, 2)
}

// FormatDuration 格式化时间间隔
func FormatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60

	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else if minutes > 0 {
		return fmt.Sprintf("%d分钟%d秒", minutes, seconds)
	} else {
		return fmt.Sprintf("%d秒", seconds)
	}
}
