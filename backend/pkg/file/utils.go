package file

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
	"wosm/internal/config"
)

// FileUtils 文件操作工具类 对应Java后端的FileUtils
type FileUtils struct{}

// WriteBytes 写入字节数据到文件 对应Java后端的writeBytes方法
func WriteBytes(data []byte, uploadDir string) (string, error) {
	// 根据文件头判断文件类型
	extension := GetFileTypeByBytes(data)
	if extension == "unknown" {
		extension = "dat"
	}

	// 生成文件名
	now := time.Now()
	datePath := now.Format("2006/01/02")
	timestamp := now.UnixNano() / 1000000
	fileName := fmt.Sprintf("%s/%d.%s", datePath, timestamp, extension)

	// 获取完整路径
	fullPath := filepath.Join(uploadDir, fileName)

	// 确保目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	// 写入文件
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %v", err)
	}

	// 返回相对路径
	return strings.ReplaceAll(fileName, "\\", "/"), nil
}

// WriteToResponse 将文件内容写入响应流 对应Java后端的writeBytes方法
func WriteToResponse(filePath string, writer io.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("打开文件失败: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return fmt.Errorf("写入响应失败: %v", err)
	}

	return nil
}

// DeleteFile 删除文件 对应Java后端的deleteFile方法
func DeleteFile(filePath string) error {
	if !FileExists(filePath) {
		return nil // 文件不存在，认为删除成功
	}

	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// FileExists 检查文件是否存在
func FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetFileName 获取文件名（不含路径） 对应Java后端的getName方法
func GetFileName(filePath string) string {
	return filepath.Base(filePath)
}

// GetFileNameWithoutExtension 获取文件名（不含扩展名）
func GetFileNameWithoutExtension(filename string) string {
	name := filepath.Base(filename)
	ext := filepath.Ext(name)
	return strings.TrimSuffix(name, ext)
}

// CheckAllowDownload 检查文件是否允许下载 对应Java后端的checkAllowDownload方法
func CheckAllowDownload(fileName string) bool {
	if fileName == "" {
		return false
	}

	// 检查是否包含路径遍历攻击（对应Java后端的".."检查）
	if strings.Contains(fileName, "..") {
		return false
	}

	// 检查是否包含路径分隔符（安全考虑）
	if strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		return false
	}

	// 检查文件扩展名是否在允许下载的列表中（对应Java后端的MimeTypeUtils.DEFAULT_ALLOWED_EXTENSION检查）
	extension := GetFileExtension(fileName)
	return IsAllowedExtension(extension, DefaultAllowedExtensions)
}

// SetAttachmentResponseHeader 设置文件下载响应头 对应Java后端的setAttachmentResponseHeader方法
func SetAttachmentResponseHeader(filename string) map[string]string {
	headers := make(map[string]string)

	// 设置内容类型为二进制流
	headers["Content-Type"] = "application/octet-stream"

	// 对文件名进行URL编码（对应Java后端的percentEncode）
	encodedFilename := PercentEncode(filename)

	// 设置Content-Disposition头（对应Java后端的格式）
	contentDisposition := fmt.Sprintf("attachment; filename=%s;filename*=utf-8''%s",
		encodedFilename, encodedFilename)
	headers["Content-Disposition"] = contentDisposition

	// 设置CORS相关头（对应Java后端的Access-Control-Expose-Headers）
	headers["Access-Control-Expose-Headers"] = "Content-Disposition,download-filename"
	headers["download-filename"] = encodedFilename

	return headers
}

// StripPrefix 移除路径前缀 对应Java后端的stripPrefix方法
func StripPrefix(resource string) string {
	prefix := config.AppConfig.File.ResourcePrefix
	if strings.HasPrefix(resource, prefix) {
		return strings.TrimPrefix(resource, prefix)
	}
	return resource
}

// GetFileInfo 获取文件信息
func GetFileInfo(filePath string) (*FileInfo, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	return &FileInfo{
		Name:         stat.Name(),
		Size:         stat.Size(),
		ModTime:      stat.ModTime(),
		IsDir:        stat.IsDir(),
		Extension:    GetFileExtension(stat.Name()),
		SizeReadable: GetFileSize(stat.Size()),
	}, nil
}

// FileInfo 文件信息结构
type FileInfo struct {
	Name         string    `json:"name"`         // 文件名
	Size         int64     `json:"size"`         // 文件大小（字节）
	SizeReadable string    `json:"sizeReadable"` // 可读的文件大小
	ModTime      time.Time `json:"modTime"`      // 修改时间
	IsDir        bool      `json:"isDir"`        // 是否为目录
	Extension    string    `json:"extension"`    // 文件扩展名
}

// CopyFile 复制文件
func CopyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer sourceFile.Close()

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制文件失败: %v", err)
	}

	return nil
}

// MoveFile 移动文件
func MoveFile(src, dst string) error {
	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return fmt.Errorf("创建目标目录失败: %v", err)
	}

	err := os.Rename(src, dst)
	if err != nil {
		return fmt.Errorf("移动文件失败: %v", err)
	}

	return nil
}

// ListFiles 列出目录中的文件
func ListFiles(dirPath string) ([]*FileInfo, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	var files []*FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		fileInfo := &FileInfo{
			Name:         info.Name(),
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			IsDir:        info.IsDir(),
			Extension:    GetFileExtension(info.Name()),
			SizeReadable: GetFileSize(info.Size()),
		}
		files = append(files, fileInfo)
	}

	return files, nil
}

// CreateDirectory 创建目录
func CreateDirectory(dirPath string) error {
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return fmt.Errorf("创建目录失败: %v", err)
	}
	return nil
}

// RemoveDirectory 删除目录及其内容
func RemoveDirectory(dirPath string) error {
	err := os.RemoveAll(dirPath)
	if err != nil {
		return fmt.Errorf("删除目录失败: %v", err)
	}
	return nil
}

// GetDirectorySize 获取目录大小
func GetDirectorySize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("计算目录大小失败: %v", err)
	}

	return size, nil
}

// IsImageFile 判断是否为图片文件（通过扩展名）
func IsImageFileByName(filename string) bool {
	extension := GetFileExtension(filename)
	return IsImageFile(extension)
}

// IsVideoFileByName 判断是否为视频文件（通过扩展名）
func IsVideoFileByName(filename string) bool {
	extension := GetFileExtension(filename)
	return IsVideoFile(extension)
}

// GetMimeType 根据文件扩展名获取MIME类型
func GetMimeType(filename string) string {
	extension := strings.ToLower(GetFileExtension(filename))

	mimeTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"bmp":  "image/bmp",
		"pdf":  "application/pdf",
		"doc":  "application/msword",
		"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"xls":  "application/vnd.ms-excel",
		"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"ppt":  "application/vnd.ms-powerpoint",
		"pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		"txt":  "text/plain",
		"html": "text/html",
		"htm":  "text/html",
		"zip":  "application/zip",
		"rar":  "application/x-rar-compressed",
		"mp4":  "video/mp4",
		"avi":  "video/x-msvideo",
		"mp3":  "audio/mpeg",
	}

	if mimeType, exists := mimeTypes[extension]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// PercentEncode 百分号编码工具方法 对应Java后端的percentEncode方法
func PercentEncode(s string) string {
	// 使用url.QueryEscape进行编码，然后将+替换为%20
	encoded := url.QueryEscape(s)
	return strings.ReplaceAll(encoded, "+", "%20")
}
