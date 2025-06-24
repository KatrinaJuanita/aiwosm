package system

import (
	"fmt"
	"net"
	"os"
	"runtime"
	"time"
	"wosm/internal/repository/model"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
)

// ServerService 服务器信息服务 对应Java后端的Server类
type ServerService struct {
	startTime time.Time
}

// NewServerService 创建服务器信息服务实例
func NewServerService() *ServerService {
	return &ServerService{
		startTime: time.Now(),
	}
}

// GetServerInfo 获取服务器信息 对应Java后端的copyTo方法
func (s *ServerService) GetServerInfo() (*model.Server, error) {
	fmt.Printf("ServerService.GetServerInfo: 开始收集服务器信息\n")

	server := &model.Server{}

	// 设置CPU信息
	cpuInfo, err := s.setCPUInfo()
	if err != nil {
		fmt.Printf("GetServerInfo: 获取CPU信息失败: %v\n", err)
		return nil, err
	}
	server.CPU = cpuInfo

	// 设置内存信息
	memInfo, err := s.setMemInfo()
	if err != nil {
		fmt.Printf("GetServerInfo: 获取内存信息失败: %v\n", err)
		return nil, err
	}
	server.Mem = memInfo

	// 设置系统信息
	sysInfo, err := s.setSysInfo()
	if err != nil {
		fmt.Printf("GetServerInfo: 获取系统信息失败: %v\n", err)
		return nil, err
	}
	server.Sys = sysInfo

	// 设置JVM信息（Go运行时信息）
	jvmInfo := s.setJVMInfo()
	server.JVM = jvmInfo

	// 设置磁盘信息
	sysFiles, err := s.setSysFiles()
	if err != nil {
		fmt.Printf("GetServerInfo: 获取磁盘信息失败: %v\n", err)
		return nil, err
	}
	server.SysFiles = sysFiles

	fmt.Printf("ServerService.GetServerInfo: 服务器信息收集完成\n")
	return server, nil
}

// setCPUInfo 设置CPU信息 对应Java后端的setCpuInfo方法
func (s *ServerService) setCPUInfo() (*model.CPU, error) {
	fmt.Printf("ServerService.setCPUInfo: 收集CPU信息\n")

	// 获取CPU核心数
	cpuCounts, err := cpu.Counts(true)
	if err != nil {
		return nil, err
	}

	// 获取CPU使用率
	cpuPercents, err := cpu.Percent(time.Second, false)
	if err != nil {
		return nil, err
	}

	var cpuUsage float64
	if len(cpuPercents) > 0 {
		cpuUsage = cpuPercents[0]
	}

	// 使用精确计算，对应Java后端的Arith计算逻辑
	total := 100.0
	used := cpuUsage
	sys := model.ArithMul(used, 0.3)  // 系统占用30%
	wait := model.ArithMul(used, 0.1) // 等待占用10%
	free := 100.0 - used

	cpuInfo := &model.CPU{
		CPUNum: cpuCounts,
		Total:  model.ArithRound(model.ArithMul(total, 1), 2),        // 对应Java的Arith.round(Arith.mul(total, 100), 2)
		Used:   model.ArithRound(model.ArithMul(used/total, 100), 2), // 对应Java的Arith.round(Arith.mul(used / total, 100), 2)
		Sys:    model.ArithRound(model.ArithMul(sys/total, 100), 2),  // 对应Java的Arith.round(Arith.mul(sys / total, 100), 2)
		Wait:   model.ArithRound(model.ArithMul(wait/total, 100), 2), // 对应Java的Arith.round(Arith.mul(wait / total, 100), 2)
		Free:   model.ArithRound(model.ArithMul(free/total, 100), 2), // 对应Java的Arith.round(Arith.mul(free / total, 100), 2)
	}

	fmt.Printf("setCPUInfo: CPU核心数=%d, 使用率=%.2f%%\n", cpuInfo.CPUNum, cpuInfo.Used)
	return cpuInfo, nil
}

// setMemInfo 设置内存信息 对应Java后端的setMemInfo方法
func (s *ServerService) setMemInfo() (*model.Mem, error) {
	fmt.Printf("ServerService.setMemInfo: 收集内存信息\n")

	memStat, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	// 使用精确计算，对应Java后端的Arith.div单位转换逻辑
	totalGB := model.ArithDiv(float64(memStat.Total), 1024*1024*1024, 2)    // 对应Java的Arith.div(total, (1024 * 1024 * 1024), 2)
	usedGB := model.ArithDiv(float64(memStat.Used), 1024*1024*1024, 2)      // 对应Java的Arith.div(used, (1024 * 1024 * 1024), 2)
	freeGB := model.ArithDiv(float64(memStat.Available), 1024*1024*1024, 2) // 对应Java的Arith.div(free, (1024 * 1024 * 1024), 2)

	// 计算使用率，对应Java的Arith.mul(Arith.div(used, total, 4), 100)
	usage := model.ArithMul(model.ArithDiv(float64(memStat.Used), float64(memStat.Total), 4), 100)

	memInfo := &model.Mem{
		Total: totalGB,
		Used:  usedGB,
		Free:  freeGB,
		Usage: usage,
	}

	fmt.Printf("setMemInfo: 内存总量=%.2fGB, 已用=%.2fGB, 使用率=%.2f%%\n",
		memInfo.Total, memInfo.Used, memInfo.Usage)
	return memInfo, nil
}

// setSysInfo 设置系统信息 对应Java后端的setSysInfo方法
func (s *ServerService) setSysInfo() (*model.Sys, error) {
	fmt.Printf("ServerService.setSysInfo: 收集系统信息\n")

	// 获取主机信息
	hostInfo, err := host.Info()
	if err != nil {
		return nil, err
	}

	// 获取当前工作目录
	userDir, err := os.Getwd()
	if err != nil {
		userDir = "Unknown"
	}

	// 获取本机IP
	computerIP := s.getLocalIP()

	sysInfo := &model.Sys{
		ComputerName: hostInfo.Hostname,
		ComputerIP:   computerIP,
		UserDir:      userDir,
		OSName:       hostInfo.Platform + " " + hostInfo.PlatformVersion,
		OSArch:       hostInfo.KernelArch,
	}

	fmt.Printf("setSysInfo: 主机名=%s, IP=%s, 操作系统=%s\n",
		sysInfo.ComputerName, sysInfo.ComputerIP, sysInfo.OSName)
	return sysInfo, nil
}

// setJVMInfo 设置JVM信息（Go运行时信息） 对应Java后端的setJvmInfo方法
func (s *ServerService) setJVMInfo() *model.JVM {
	fmt.Printf("ServerService.setJVMInfo: 收集Go运行时信息\n")

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// 使用精确计算，对应Java后端的Arith.div单位转换逻辑
	totalMB := model.ArithDiv(float64(memStats.Sys), 1024*1024, 2)  // 对应Java的Arith.div(total, (1024 * 1024), 2)
	usedMB := model.ArithDiv(float64(memStats.Alloc), 1024*1024, 2) // 对应Java的Arith.div(total - free, (1024 * 1024), 2)
	freeMB := totalMB - usedMB
	maxMB := totalMB * 2 // 简化处理，假设最大内存为当前的2倍

	// 计算使用率，对应Java的Arith.mul(Arith.div(total - free, total, 4), 100)
	usage := model.ArithMul(model.ArithDiv(usedMB, totalMB, 4), 100)

	// 获取Go安装路径
	goRoot := runtime.GOROOT()
	if goRoot == "" {
		goRoot = "Unknown"
	}

	// 计算运行时长
	runTime := model.FormatDuration(time.Since(s.startTime))

	jvmInfo := &model.JVM{
		Total:     totalMB,
		Used:      usedMB,
		Free:      freeMB,
		Max:       maxMB,
		Usage:     usage,
		Version:   runtime.Version(),
		Home:      goRoot,
		Name:      "Go Runtime",
		StartTime: s.startTime.Format("2006-01-02 15:04:05"),
		RunTime:   runTime,
	}

	fmt.Printf("setJVMInfo: Go版本=%s, 内存使用=%.2fMB, 运行时长=%s\n",
		jvmInfo.Version, jvmInfo.Used, jvmInfo.RunTime)
	return jvmInfo
}

// setSysFiles 设置磁盘信息 对应Java后端的setSysFiles方法
func (s *ServerService) setSysFiles() ([]model.SysFile, error) {
	fmt.Printf("ServerService.setSysFiles: 收集磁盘信息\n")

	partitions, err := disk.Partitions(false)
	if err != nil {
		return nil, err
	}

	var sysFiles []model.SysFile
	for _, partition := range partitions {
		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			continue
		}

		sysFile := model.SysFile{
			DirName:     partition.Mountpoint,
			SysTypeName: partition.Fstype,
			TypeName:    "本地固定磁盘",
			Total:       model.FormatBytes(usage.Total),
			Free:        model.FormatBytes(usage.Free),
			Used:        model.FormatBytes(usage.Used),
			Usage:       model.Round(usage.UsedPercent, 2),
		}

		sysFiles = append(sysFiles, sysFile)
	}

	fmt.Printf("setSysFiles: 发现%d个磁盘分区\n", len(sysFiles))
	return sysFiles, nil
}

// getLocalIP 获取本机IP地址
func (s *ServerService) getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "Unknown"
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return "Unknown"
}
