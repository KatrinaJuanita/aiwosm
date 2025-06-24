package common

import (
	"fmt"
	"path/filepath"
	"strings"
	"wosm/internal/config"
	"wosm/pkg/file"
	"wosm/pkg/response"

	"github.com/gin-gonic/gin"
)

// FileController 文件管理控制器 对应Java后端的CommonController
type FileController struct{}

// NewFileController 创建文件管理控制器实例
func NewFileController() *FileController {
	return &FileController{}
}

// Upload 通用文件上传（单个） 对应Java后端的uploadFile方法
// @Summary 上传文件
// @Description 上传单个文件
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传的文件"
// @Success 200 {object} response.Response{data=file.UploadResult}
// @Router /common/upload [post]
func (c *FileController) Upload(ctx *gin.Context) {
	fmt.Printf("FileController.Upload: 上传文件\n")

	// 获取上传的文件
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		response.ErrorWithMessage(ctx, "获取上传文件失败: "+err.Error())
		return
	}

	// 上传文件
	result, err := file.Upload(fileHeader)
	if err != nil {
		response.ErrorWithMessage(ctx, err.Error())
		return
	}

	// 构建完整的访问URL
	serverURL := c.getServerURL(ctx)
	result.URL = serverURL + result.URL

	// 返回成功结果（与Java后端格式一致）
	data := map[string]any{
		"url":              result.URL,
		"fileName":         result.FileName,
		"newFileName":      result.NewFileName,
		"originalFilename": result.OriginalFilename,
		"size":             result.Size,
		"extension":        result.Extension,
	}

	response.SuccessWithData(ctx, data)
}

// Uploads 通用文件上传（多个） 对应Java后端的uploadFiles方法
// @Summary 批量上传文件
// @Description 上传多个文件
// @Tags 文件管理
// @Accept multipart/form-data
// @Produce json
// @Param files formData file true "上传的文件列表"
// @Success 200 {object} response.Response
// @Router /common/uploads [post]
func (c *FileController) Uploads(ctx *gin.Context) {
	fmt.Printf("FileController.Uploads: 批量上传文件\n")

	// 获取多文件上传表单
	form, err := ctx.MultipartForm()
	if err != nil {
		response.ErrorWithMessage(ctx, "获取上传表单失败: "+err.Error())
		return
	}

	files := form.File["files"]
	if len(files) == 0 {
		response.ErrorWithMessage(ctx, "没有选择文件")
		return
	}

	var urls []string
	var fileNames []string
	var newFileNames []string
	var originalFilenames []string

	serverURL := c.getServerURL(ctx)

	// 逐个上传文件
	for _, fileHeader := range files {
		result, err := file.Upload(fileHeader)
		if err != nil {
			response.ErrorWithMessage(ctx, fmt.Sprintf("上传文件 %s 失败: %s", fileHeader.Filename, err.Error()))
			return
		}

		// 构建完整的访问URL
		fullURL := serverURL + result.URL

		urls = append(urls, fullURL)
		fileNames = append(fileNames, result.FileName)
		newFileNames = append(newFileNames, result.NewFileName)
		originalFilenames = append(originalFilenames, result.OriginalFilename)
	}

	// 返回成功结果（与Java后端格式一致：逗号分隔的字符串）
	data := map[string]any{
		"urls":              strings.Join(urls, ","),
		"fileNames":         strings.Join(fileNames, ","),
		"newFileNames":      strings.Join(newFileNames, ","),
		"originalFilenames": strings.Join(originalFilenames, ","),
	}

	response.SuccessWithData(ctx, data)
}

// Download 通用文件下载 对应Java后端的fileDownload方法
// @Summary 下载文件
// @Description 下载指定文件
// @Tags 文件管理
// @Accept json
// @Produce application/octet-stream
// @Param fileName query string true "文件名"
// @Param delete query bool false "下载后是否删除文件"
// @Success 200 {file} file "文件内容"
// @Router /common/download [get]
func (c *FileController) Download(ctx *gin.Context) {
	fileName := ctx.Query("fileName")
	deleteAfter := ctx.Query("delete") == "true"

	fmt.Printf("FileController.Download: 下载文件, FileName=%s, Delete=%v\n", fileName, deleteAfter)

	if fileName == "" {
		response.ErrorWithMessage(ctx, "文件名不能为空")
		return
	}

	// 检查文件名是否合法
	if !file.CheckAllowDownload(fileName) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("文件名称(%s)非法，不允许下载", fileName))
		return
	}

	// 构建文件路径
	filePath := filepath.Join(file.GetDownloadPath(), fileName)

	// 检查文件是否存在
	if !file.FileExists(filePath) {
		response.ErrorWithMessage(ctx, "文件不存在")
		return
	}

	// 生成真实文件名（移除时间戳前缀）
	realFileName := c.getRealFileName(fileName)

	// 设置响应头
	headers := file.SetAttachmentResponseHeader(realFileName)
	for key, value := range headers {
		ctx.Header(key, value)
	}

	// 发送文件
	ctx.File(filePath)

	// 如果需要删除文件
	if deleteAfter {
		go func() {
			if err := file.DeleteFile(filePath); err != nil {
				fmt.Printf("删除文件失败: %v\n", err)
			}
		}()
	}
}

// ResourceDownload 本地资源下载 对应Java后端的resourceDownload方法
// @Summary 下载资源文件
// @Description 下载本地资源文件
// @Tags 文件管理
// @Accept json
// @Produce application/octet-stream
// @Param resource query string true "资源路径"
// @Success 200 {file} file "文件内容"
// @Router /common/download/resource [get]
func (c *FileController) ResourceDownload(ctx *gin.Context) {
	resource := ctx.Query("resource")

	fmt.Printf("FileController.ResourceDownload: 下载资源文件, Resource=%s\n", resource)

	if resource == "" {
		response.ErrorWithMessage(ctx, "资源路径不能为空")
		return
	}

	// 检查资源路径是否合法
	if !file.CheckAllowDownload(filepath.Base(resource)) {
		response.ErrorWithMessage(ctx, fmt.Sprintf("资源文件(%s)非法，不允许下载", resource))
		return
	}

	// 构建本地文件路径
	localPath := config.AppConfig.File.UploadPath
	downloadPath := filepath.Join(localPath, file.StripPrefix(resource))

	// 检查文件是否存在
	if !file.FileExists(downloadPath) {
		response.ErrorWithMessage(ctx, "资源文件不存在")
		return
	}

	// 获取下载文件名
	downloadName := filepath.Base(downloadPath)

	// 设置响应头
	headers := file.SetAttachmentResponseHeader(downloadName)
	for key, value := range headers {
		ctx.Header(key, value)
	}

	// 发送文件
	ctx.File(downloadPath)
}

// getServerURL 获取服务器URL
func (c *FileController) getServerURL(ctx *gin.Context) string {
	scheme := "http"
	if ctx.Request.TLS != nil {
		scheme = "https"
	}

	host := ctx.Request.Host
	return fmt.Sprintf("%s://%s", scheme, host)
}

// getRealFileName 获取真实文件名（移除时间戳前缀） 对应Java后端的realFileName逻辑
func (c *FileController) getRealFileName(fileName string) string {
	// 对应Java后端：System.currentTimeMillis() + fileName.substring(fileName.indexOf("_") + 1)
	// 这里我们需要移除时间戳前缀，返回原始文件名
	underscoreIndex := strings.Index(fileName, "_")
	if underscoreIndex > 0 {
		// 返回下划线后的部分（原始文件名）
		return fileName[underscoreIndex+1:]
	}
	return fileName
}

// GetFileInfo 获取文件信息
// @Summary 获取文件信息
// @Description 获取指定文件的详细信息
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param fileName query string true "文件名"
// @Success 200 {object} response.Response{data=file.FileInfo}
// @Router /common/fileInfo [get]
func (c *FileController) GetFileInfo(ctx *gin.Context) {
	fileName := ctx.Query("fileName")

	fmt.Printf("FileController.GetFileInfo: 获取文件信息, FileName=%s\n", fileName)

	if fileName == "" {
		response.ErrorWithMessage(ctx, "文件名不能为空")
		return
	}

	// 构建文件路径
	filePath := filepath.Join(config.AppConfig.File.UploadPath, file.StripPrefix(fileName))

	// 获取文件信息
	fileInfo, err := file.GetFileInfo(filePath)
	if err != nil {
		response.ErrorWithMessage(ctx, "获取文件信息失败: "+err.Error())
		return
	}

	response.SuccessWithData(ctx, fileInfo)
}

// ListFiles 列出目录文件
// @Summary 列出目录文件
// @Description 列出指定目录下的文件列表
// @Tags 文件管理
// @Accept json
// @Produce json
// @Param path query string false "目录路径"
// @Success 200 {object} response.Response{data=[]file.FileInfo}
// @Router /common/listFiles [get]
func (c *FileController) ListFiles(ctx *gin.Context) {
	dirPath := ctx.Query("path")
	if dirPath == "" {
		dirPath = config.AppConfig.File.UploadPath
	}

	fmt.Printf("FileController.ListFiles: 列出目录文件, Path=%s\n", dirPath)

	// 安全检查：确保路径在上传目录内
	if !strings.HasPrefix(dirPath, config.AppConfig.File.UploadPath) {
		response.ErrorWithMessage(ctx, "访问路径非法")
		return
	}

	// 列出文件
	files, err := file.ListFiles(dirPath)
	if err != nil {
		response.ErrorWithMessage(ctx, "列出文件失败: "+err.Error())
		return
	}

	response.SuccessWithData(ctx, files)
}
