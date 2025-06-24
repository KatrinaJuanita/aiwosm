package file

import (
	"fmt"
	"strings"
)

// FileUploadError 文件上传基础错误 对应Java后端的FileUploadException
type FileUploadError struct {
	Message string
	Cause   error
}

func (e *FileUploadError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// NewFileUploadError 创建文件上传错误
func NewFileUploadError(message string, cause error) *FileUploadError {
	return &FileUploadError{
		Message: message,
		Cause:   cause,
	}
}

// FileSizeLimitExceededError 文件大小超限错误 对应Java后端的FileSizeLimitExceededException
type FileSizeLimitExceededError struct {
	*FileUploadError
	MaxSize int64
}

func (e *FileSizeLimitExceededError) Error() string {
	maxSizeMB := e.MaxSize / 1024 / 1024
	return fmt.Sprintf("文件大小超出限制，最大允许 %d MB", maxSizeMB)
}

// NewFileSizeLimitExceededError 创建文件大小超限错误
func NewFileSizeLimitExceededError(maxSize int64) *FileSizeLimitExceededError {
	return &FileSizeLimitExceededError{
		FileUploadError: &FileUploadError{
			Message: fmt.Sprintf("文件大小超出限制，最大允许 %d MB", maxSize/1024/1024),
		},
		MaxSize: maxSize,
	}
}

// FileNameLengthLimitExceededError 文件名长度超限错误 对应Java后端的FileNameLengthLimitExceededException
type FileNameLengthLimitExceededError struct {
	*FileUploadError
	MaxLength int
}

func (e *FileNameLengthLimitExceededError) Error() string {
	return fmt.Sprintf("文件名长度超出限制，最大允许 %d 个字符", e.MaxLength)
}

// NewFileNameLengthLimitExceededError 创建文件名长度超限错误
func NewFileNameLengthLimitExceededError(maxLength int) *FileNameLengthLimitExceededError {
	return &FileNameLengthLimitExceededError{
		FileUploadError: &FileUploadError{
			Message: fmt.Sprintf("文件名长度超出限制，最大允许 %d 个字符", maxLength),
		},
		MaxLength: maxLength,
	}
}

// InvalidExtensionError 无效文件扩展名错误 对应Java后端的InvalidExtensionException
type InvalidExtensionError struct {
	*FileUploadError
	AllowedExtensions []string
	Extension         string
	Filename          string
}

func (e *InvalidExtensionError) Error() string {
	return fmt.Sprintf("文件[%s]后缀[%s]不正确，请上传%s格式",
		e.Filename, e.Extension, strings.Join(e.AllowedExtensions, ","))
}

// NewInvalidExtensionError 创建无效文件扩展名错误
func NewInvalidExtensionError(allowedExtensions []string, extension, filename string) *InvalidExtensionError {
	return &InvalidExtensionError{
		FileUploadError: &FileUploadError{
			Message: fmt.Sprintf("文件[%s]后缀[%s]不正确，请上传%s格式",
				filename, extension, strings.Join(allowedExtensions, ",")),
		},
		AllowedExtensions: allowedExtensions,
		Extension:         extension,
		Filename:          filename,
	}
}

// InvalidImageExtensionError 无效图片文件扩展名错误 对应Java后端的InvalidImageExtensionException
type InvalidImageExtensionError struct {
	*InvalidExtensionError
}

func (e *InvalidImageExtensionError) Error() string {
	return fmt.Sprintf("图片文件[%s]后缀[%s]不正确，请上传%s格式的图片",
		e.Filename, e.Extension, strings.Join(e.AllowedExtensions, ","))
}

// NewInvalidImageExtensionError 创建无效图片文件扩展名错误
func NewInvalidImageExtensionError(allowedExtensions []string, extension, filename string) *InvalidImageExtensionError {
	return &InvalidImageExtensionError{
		InvalidExtensionError: &InvalidExtensionError{
			FileUploadError: &FileUploadError{
				Message: fmt.Sprintf("图片文件[%s]后缀[%s]不正确，请上传%s格式的图片",
					filename, extension, strings.Join(allowedExtensions, ",")),
			},
			AllowedExtensions: allowedExtensions,
			Extension:         extension,
			Filename:          filename,
		},
	}
}

// InvalidVideoExtensionError 无效视频文件扩展名错误 对应Java后端的InvalidVideoExtensionException
type InvalidVideoExtensionError struct {
	*InvalidExtensionError
}

func (e *InvalidVideoExtensionError) Error() string {
	return fmt.Sprintf("视频文件[%s]后缀[%s]不正确，请上传%s格式的视频",
		e.Filename, e.Extension, strings.Join(e.AllowedExtensions, ","))
}

// NewInvalidVideoExtensionError 创建无效视频文件扩展名错误
func NewInvalidVideoExtensionError(allowedExtensions []string, extension, filename string) *InvalidVideoExtensionError {
	return &InvalidVideoExtensionError{
		InvalidExtensionError: &InvalidExtensionError{
			FileUploadError: &FileUploadError{
				Message: fmt.Sprintf("视频文件[%s]后缀[%s]不正确，请上传%s格式的视频",
					filename, extension, strings.Join(allowedExtensions, ",")),
			},
			AllowedExtensions: allowedExtensions,
			Extension:         extension,
			Filename:          filename,
		},
	}
}

// InvalidMediaExtensionError 无效媒体文件扩展名错误 对应Java后端的InvalidMediaExtensionException
type InvalidMediaExtensionError struct {
	*InvalidExtensionError
}

func (e *InvalidMediaExtensionError) Error() string {
	return fmt.Sprintf("媒体文件[%s]后缀[%s]不正确，请上传%s格式的媒体文件",
		e.Filename, e.Extension, strings.Join(e.AllowedExtensions, ","))
}

// NewInvalidMediaExtensionError 创建无效媒体文件扩展名错误
func NewInvalidMediaExtensionError(allowedExtensions []string, extension, filename string) *InvalidMediaExtensionError {
	return &InvalidMediaExtensionError{
		InvalidExtensionError: &InvalidExtensionError{
			FileUploadError: &FileUploadError{
				Message: fmt.Sprintf("媒体文件[%s]后缀[%s]不正确，请上传%s格式的媒体文件",
					filename, extension, strings.Join(allowedExtensions, ",")),
			},
			AllowedExtensions: allowedExtensions,
			Extension:         extension,
			Filename:          filename,
		},
	}
}

// InvalidFlashExtensionError 无效Flash文件扩展名错误 对应Java后端的InvalidFlashExtensionException
type InvalidFlashExtensionError struct {
	*InvalidExtensionError
}

func (e *InvalidFlashExtensionError) Error() string {
	return fmt.Sprintf("Flash文件[%s]后缀[%s]不正确，请上传%s格式的Flash文件",
		e.Filename, e.Extension, strings.Join(e.AllowedExtensions, ","))
}

// NewInvalidFlashExtensionError 创建无效Flash文件扩展名错误
func NewInvalidFlashExtensionError(allowedExtensions []string, extension, filename string) *InvalidFlashExtensionError {
	return &InvalidFlashExtensionError{
		InvalidExtensionError: &InvalidExtensionError{
			FileUploadError: &FileUploadError{
				Message: fmt.Sprintf("Flash文件[%s]后缀[%s]不正确，请上传%s格式的Flash文件",
					filename, extension, strings.Join(allowedExtensions, ",")),
			},
			AllowedExtensions: allowedExtensions,
			Extension:         extension,
			Filename:          filename,
		},
	}
}
