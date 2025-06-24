package file

import (
	"strings"
)

// MimeTypeUtils MIME类型工具 对应Java后端的MimeTypeUtils
type MimeTypeUtils struct{}

// MIME类型常量
const (
	ImagePNG  = "image/png"
	ImageJPG  = "image/jpg"
	ImageJPEG = "image/jpeg"
	ImageBMP  = "image/bmp"
	ImageGIF  = "image/gif"
)

// 文件扩展名常量 对应Java后端的MimeTypeUtils
var (
	// 图片文件扩展名
	ImageExtensions = []string{"bmp", "gif", "jpg", "jpeg", "png"}
	
	// Flash文件扩展名
	FlashExtensions = []string{"swf", "flv"}
	
	// 媒体文件扩展名
	MediaExtensions = []string{"swf", "flv", "mp3", "wav", "wma", "wmv", "mid", "avi", "mpg", "asf", "rm", "rmvb"}
	
	// 视频文件扩展名
	VideoExtensions = []string{"mp4", "avi", "rmvb"}
	
	// 默认允许的文件扩展名 对应Java后端的DEFAULT_ALLOWED_EXTENSION
	DefaultAllowedExtensions = []string{
		// 图片
		"bmp", "gif", "jpg", "jpeg", "png",
		// word excel powerpoint
		"doc", "docx", "xls", "xlsx", "ppt", "pptx", "html", "htm", "txt",
		// 压缩文件
		"rar", "zip", "gz", "bz2",
		// 视频格式
		"mp4", "avi", "rmvb",
		// pdf
		"pdf",
	}
)

// GetExtensionByMimeType 根据MIME类型获取文件扩展名 对应Java后端的getExtension方法
func GetExtensionByMimeType(mimeType string) string {
	switch mimeType {
	case ImagePNG:
		return "png"
	case ImageJPG:
		return "jpg"
	case ImageJPEG:
		return "jpeg"
	case ImageBMP:
		return "bmp"
	case ImageGIF:
		return "gif"
	default:
		return ""
	}
}

// IsAllowedExtension 判断文件扩展名是否允许 对应Java后端的isAllowedExtension方法
func IsAllowedExtension(extension string, allowedExtensions []string) bool {
	if allowedExtensions == nil {
		return true
	}
	
	extension = strings.ToLower(extension)
	for _, allowed := range allowedExtensions {
		if strings.ToLower(allowed) == extension {
			return true
		}
	}
	return false
}

// GetFileExtension 获取文件扩展名
func GetFileExtension(filename string) string {
	if filename == "" {
		return ""
	}
	
	lastDot := strings.LastIndex(filename, ".")
	if lastDot == -1 || lastDot == len(filename)-1 {
		return ""
	}
	
	return strings.ToLower(filename[lastDot+1:])
}

// IsImageFile 判断是否为图片文件
func IsImageFile(extension string) bool {
	return IsAllowedExtension(extension, ImageExtensions)
}

// IsVideoFile 判断是否为视频文件
func IsVideoFile(extension string) bool {
	return IsAllowedExtension(extension, VideoExtensions)
}

// IsMediaFile 判断是否为媒体文件
func IsMediaFile(extension string) bool {
	return IsAllowedExtension(extension, MediaExtensions)
}

// IsFlashFile 判断是否为Flash文件
func IsFlashFile(extension string) bool {
	return IsAllowedExtension(extension, FlashExtensions)
}

// GetFileTypeByExtension 根据扩展名获取文件类型描述
func GetFileTypeByExtension(extension string) string {
	extension = strings.ToLower(extension)
	
	if IsImageFile(extension) {
		return "图片文件"
	}
	if IsVideoFile(extension) {
		return "视频文件"
	}
	if IsMediaFile(extension) {
		return "媒体文件"
	}
	if IsFlashFile(extension) {
		return "Flash文件"
	}
	
	// 文档文件
	docExtensions := []string{"doc", "docx", "xls", "xlsx", "ppt", "pptx", "pdf", "txt", "html", "htm"}
	if IsAllowedExtension(extension, docExtensions) {
		return "文档文件"
	}
	
	// 压缩文件
	archiveExtensions := []string{"rar", "zip", "gz", "bz2"}
	if IsAllowedExtension(extension, archiveExtensions) {
		return "压缩文件"
	}
	
	return "其他文件"
}

// ValidateFileType 验证文件类型 对应Java后端的assertAllowed方法的类型检查部分
func ValidateFileType(filename string, allowedExtensions []string) error {
	extension := GetFileExtension(filename)
	if extension == "" {
		return NewInvalidExtensionError(allowedExtensions, "", filename)
	}
	
	if !IsAllowedExtension(extension, allowedExtensions) {
		// 根据允许的扩展名类型返回具体的错误
		if isSliceEqual(allowedExtensions, ImageExtensions) {
			return NewInvalidImageExtensionError(allowedExtensions, extension, filename)
		}
		if isSliceEqual(allowedExtensions, VideoExtensions) {
			return NewInvalidVideoExtensionError(allowedExtensions, extension, filename)
		}
		if isSliceEqual(allowedExtensions, MediaExtensions) {
			return NewInvalidMediaExtensionError(allowedExtensions, extension, filename)
		}
		if isSliceEqual(allowedExtensions, FlashExtensions) {
			return NewInvalidFlashExtensionError(allowedExtensions, extension, filename)
		}
		
		return NewInvalidExtensionError(allowedExtensions, extension, filename)
	}
	
	return nil
}

// isSliceEqual 比较两个字符串切片是否相等
func isSliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// GetFileTypeByBytes 根据文件字节头判断文件类型 对应Java后端的getFileExtendName方法
func GetFileTypeByBytes(data []byte) string {
	if len(data) < 10 {
		return "unknown"
	}
	
	// GIF
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x38 &&
		(data[4] == 0x37 || data[4] == 0x39) && data[5] == 0x61 {
		return "gif"
	}
	
	// JPEG
	if data[6] == 0x4A && data[7] == 0x46 && data[8] == 0x49 && data[9] == 0x46 {
		return "jpg"
	}
	
	// BMP
	if data[0] == 0x42 && data[1] == 0x4D {
		return "bmp"
	}
	
	// PNG
	if data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "png"
	}
	
	return "unknown"
}
