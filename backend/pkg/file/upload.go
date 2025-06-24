package file

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"
	"wosm/internal/config"

	"github.com/google/uuid"
)

// FileUploadUtils 文件上传工具类 对应Java后端的FileUploadUtils
type FileUploadUtils struct{}

// 序列号生成器 对应Java后端的Seq.getId
var uploadSeq int64

// UploadResult 上传结果 对应Java后端的AjaxResult返回格式
type UploadResult struct {
	FileName         string `json:"fileName"`         // 服务器文件名（完整路径）
	OriginalFilename string `json:"originalFilename"` // 原始文件名
	URL              string `json:"url"`              // 访问URL
	NewFileName      string `json:"newFileName"`      // 新文件名（不含路径）
	Size             int64  `json:"size"`             // 文件大小（字节）
	SizeReadable     string `json:"sizeReadable"`     // 可读的文件大小
	Extension        string `json:"extension"`        // 文件扩展名
	MimeType         string `json:"mimeType"`         // MIME类型
	FileType         string `json:"fileType"`         // 文件类型描述
	UploadTime       string `json:"uploadTime"`       // 上传时间
}

// Upload 文件上传（使用默认配置） 对应Java后端的upload(MultipartFile file)方法
func Upload(file *multipart.FileHeader) (*UploadResult, error) {
	return UploadWithConfig(file, config.AppConfig.File.UploadPath, config.AppConfig.File.AllowedExtensions, false)
}

// UploadWithPath 文件上传（指定路径） 对应Java后端的upload(String baseDir, MultipartFile file)方法
func UploadWithPath(file *multipart.FileHeader, uploadPath string) (*UploadResult, error) {
	return UploadWithConfig(file, uploadPath, config.AppConfig.File.AllowedExtensions, false)
}

// UploadWithExtensions 文件上传（指定允许的扩展名） 对应Java后端的upload方法
func UploadWithExtensions(file *multipart.FileHeader, uploadPath string, allowedExtensions []string) (*UploadResult, error) {
	return UploadWithConfig(file, uploadPath, allowedExtensions, false)
}

// UploadWithConfig 文件上传（完整配置） 对应Java后端的upload方法
func UploadWithConfig(file *multipart.FileHeader, uploadPath string, allowedExtensions []string, useCustomNaming bool) (*UploadResult, error) {
	// 验证文件名长度
	if len(file.Filename) > config.AppConfig.File.MaxNameLength {
		return nil, NewFileNameLengthLimitExceededError(config.AppConfig.File.MaxNameLength)
	}

	// 验证文件大小和类型
	if err := validateFile(file, allowedExtensions); err != nil {
		return nil, err
	}

	// 生成文件名
	var fileName string
	if useCustomNaming {
		fileName = generateUUIDFilename(file.Filename)
	} else {
		fileName = generateDateFilename(file.Filename)
	}

	// 获取绝对路径
	absPath, err := getAbsoluteFilePath(uploadPath, fileName)
	if err != nil {
		return nil, err
	}

	// 保存文件
	if err := saveFile(file, absPath); err != nil {
		return nil, err
	}

	// 生成访问路径
	relativePath := getRelativePath(uploadPath, fileName)

	// 获取文件扩展名和类型信息
	extension := GetFileExtension(file.Filename)

	// 构建返回结果
	result := &UploadResult{
		FileName:         relativePath,
		OriginalFilename: file.Filename,
		URL:              config.AppConfig.File.ResourcePrefix + "/" + relativePath,
		NewFileName:      filepath.Base(fileName),
		Size:             file.Size,
		SizeReadable:     GetFileSize(file.Size),
		Extension:        extension,
		MimeType:         GetMimeType(file.Filename),
		FileType:         GetFileTypeByExtension(extension),
		UploadTime:       time.Now().Format("2006-01-02 15:04:05"),
	}

	return result, nil
}

// validateFile 验证文件 对应Java后端的assertAllowed方法
func validateFile(file *multipart.FileHeader, allowedExtensions []string) error {
	// 验证文件大小
	if file.Size > config.AppConfig.File.MaxSize {
		return NewFileSizeLimitExceededError(config.AppConfig.File.MaxSize)
	}

	// 验证文件类型
	if err := ValidateFileType(file.Filename, allowedExtensions); err != nil {
		return err
	}

	return nil
}

// generateDateFilename 生成基于日期的文件名 对应Java后端的extractFilename方法
func generateDateFilename(originalFilename string) string {
	// 获取文件扩展名
	extension := GetFileExtension(originalFilename)

	// 获取文件名（不含扩展名）
	baseName := strings.TrimSuffix(originalFilename, "."+extension)

	// 生成日期路径 (YYYY/MM/DD)
	now := time.Now()
	datePath := now.Format("2006/01/02")

	// 生成序列号（对应Java后端的Seq.getId）
	sequence := atomic.AddInt64(&uploadSeq, 1)

	// 组合文件名: 日期路径/原文件名_序列号.扩展名
	return fmt.Sprintf("%s/%s_%d.%s", datePath, baseName, sequence, extension)
}

// generateUUIDFilename 生成基于UUID的文件名 对应Java后端的uuidFilename方法
func generateUUIDFilename(originalFilename string) string {
	// 获取文件扩展名
	extension := GetFileExtension(originalFilename)

	// 生成日期路径 (YYYY/MM/DD)
	now := time.Now()
	datePath := now.Format("2006/01/02")

	// 生成UUID
	uuidStr := strings.ReplaceAll(uuid.New().String(), "-", "")

	// 组合文件名: 日期路径/UUID.扩展名
	return fmt.Sprintf("%s/%s.%s", datePath, uuidStr, extension)
}

// getAbsoluteFilePath 获取文件的绝对路径 对应Java后端的getAbsoluteFile方法
func getAbsoluteFilePath(uploadPath, fileName string) (string, error) {
	absPath := filepath.Join(uploadPath, fileName)

	// 确保目录存在
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	return absPath, nil
}

// saveFile 保存文件 对应Java后端的file.transferTo方法
func saveFile(fileHeader *multipart.FileHeader, destPath string) error {
	// 打开上传的文件
	src, err := fileHeader.Open()
	if err != nil {
		return fmt.Errorf("打开上传文件失败: %v", err)
	}
	defer src.Close()

	// 创建目标文件
	dst, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %v", err)
	}
	defer dst.Close()

	// 复制文件内容
	_, err = io.Copy(dst, src)
	if err != nil {
		return fmt.Errorf("保存文件失败: %v", err)
	}

	return nil
}

// getRelativePath 获取相对路径 对应Java后端的getPathFileName方法
func getRelativePath(uploadPath, fileName string) string {
	// 移除上传路径前缀，返回相对路径
	uploadPath = filepath.Clean(uploadPath)
	if strings.HasSuffix(uploadPath, string(filepath.Separator)) {
		uploadPath = uploadPath[:len(uploadPath)-1]
	}

	// 返回相对于上传根目录的路径
	return strings.ReplaceAll(fileName, "\\", "/")
}

// GetUploadPath 获取上传路径 对应Java后端的getUploadPath方法
func GetUploadPath() string {
	return config.AppConfig.File.UploadPath
}

// GetAvatarPath 获取头像上传路径 对应Java后端的getAvatarPath方法
func GetAvatarPath() string {
	return filepath.Join(config.AppConfig.File.UploadPath, "avatar")
}

// GetImportPath 获取导入文件路径 对应Java后端的getImportPath方法
func GetImportPath() string {
	return filepath.Join(config.AppConfig.File.UploadPath, "import")
}

// GetDownloadPath 获取下载文件路径 对应Java后端的getDownloadPath方法
func GetDownloadPath() string {
	return filepath.Join(config.AppConfig.File.UploadPath, "download")
}

// IsValidFilename 检查文件名是否有效
func IsValidFilename(filename string) bool {
	if filename == "" {
		return false
	}

	// 检查是否包含非法字符
	invalidChars := []string{"\\", "/", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalidChars {
		if strings.Contains(filename, char) {
			return false
		}
	}

	return true
}

// GetFileSize 获取文件大小的友好显示
func GetFileSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}
