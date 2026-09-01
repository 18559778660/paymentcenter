package controller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// ChannelController 控制层：通道列表。
type ChannelController struct {
	app *service.App
}

// NewChannelController 创建通道控制器。
func NewChannelController(app *service.App) *ChannelController {
	return &ChannelController{app: app}
}

// List 通道列表。
func (m *ChannelController) List(c *gin.Context) {
	q := service.ChannelListQuery{
		Name: c.Query("name"),
	}
	if v := c.Query("id"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			uid := uint(id)
			q.ID = &uid
		}
	}
	list, err := m.app.ListChannels(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Platforms 通道平台分类下拉。
func (m *ChannelController) Platforms(c *gin.Context) {
	list, err := m.app.ListPlatformOptions()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Create 新增通道。
func (m *ChannelController) Create(c *gin.Context) {
	var req service.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.CreateChannel(req, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, item, "created")
}

// Update 编辑通道信息。
func (m *ChannelController) Update(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	var req service.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateChannel(id, req, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// UpdateLimits 更新限制配置。
func (m *ChannelController) UpdateLimits(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	var req service.UpdateChannelLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.UpdateChannelLimits(id, req, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, item, "updated")
}

// SetStatus 启用/禁用通道。
func (m *ChannelController) SetStatus(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	var req struct {
		Status bool `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	item, err := m.app.SetChannelStatus(id, req.Status, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.Success(c, item)
}

// UploadPackage 上传并绑定通道压缩包。
func (m *ChannelController) UploadPackage(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "请选择压缩包")
		return
	}
	if file.Size > 50*1024*1024 {
		response.Fail(c, "压缩包不能超过 50MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".zip", ".rar", ".7z":
	default:
		response.Fail(c, "仅支持 zip/rar/7z")
		return
	}
	relativePath := service.BuildChannelPackageRelativePath(id, file.Filename)
	absPath := filepath.Join("uploads", filepath.FromSlash(relativePath))
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		response.Fail(c, "上传失败")
		return
	}
	if err := c.SaveUploadedFile(file, absPath); err != nil {
		response.Fail(c, "上传失败")
		return
	}
	item, err := m.app.SetChannelPackage(id, relativePath, operatorName(c))
	if err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, item, "上传成功")
}

// Delete 删除通道。
func (m *ChannelController) Delete(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	if err := m.app.DeleteChannel(id); err != nil {
		writeChannelError(c, err)
		return
	}
	response.SuccessMsg(c, nil, "deleted")
}

// DownloadPackage 下载通道压缩包。
func (m *ChannelController) DownloadPackage(c *gin.Context) {
	id, err := parseChannelID(c)
	if err != nil {
		response.Fail(c, "无效的通道ID")
		return
	}
	absPath, downloadName, err := m.app.ChannelPackageFile(id)
	if err != nil {
		writeChannelPackageError(c, err)
		return
	}
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", downloadName))
	c.File(absPath)
}

func writeChannelPackageError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrChannelNotFound):
		response.Fail(c, "通道不存在")
	case errors.Is(err, service.ErrChannelPackageNotFound):
		response.Fail(c, "压缩包不存在")
	default:
		response.Fail(c, err.Error())
	}
}

func parseChannelID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, err
	}
	return uint(id), nil
}

func writeChannelError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrChannelNotFound):
		response.Fail(c, "通道不存在")
	case errors.Is(err, service.ErrChannelNameExists):
		response.Fail(c, "通道名已存在")
	case errors.Is(err, service.ErrChannelNameInvalid):
		response.Fail(c, "请输入通道名")
	case errors.Is(err, service.ErrChannelInterceptRangeInvalid):
		response.Fail(c, "限制最小金额不能高于限制最大金额")
	case errors.Is(err, service.ErrChannelSuccessSettingInvalid):
		response.Fail(c, "成功设置需同时配置支付频率，以及成功次数或失败次数")
	case errors.Is(err, service.ErrChannelPlatformInvalid):
		response.Fail(c, "请选择有效通道平台")
	case errors.Is(err, service.ErrChannelHasAccounts):
		response.Fail(c, "该通道仍有关联通道账号，无法删除")
	default:
		response.Fail(c, err.Error())
	}
}
