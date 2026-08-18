package controller

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"paymentcenter/internal/middleware"
	"paymentcenter/internal/service"
	"paymentcenter/internal/util/response"
)

// MerchantController 控制层：商户列表与新建。
type MerchantController struct {
	app *service.App
}

// NewMerchantController 创建商户控制器。
func NewMerchantController(app *service.App) *MerchantController {
	return &MerchantController{app: app}
}

// List 商户列表，支持筛选。
func (m *MerchantController) List(c *gin.Context) {
	q := service.MerchantListQuery{
		Name: c.Query("name"),
	}
	if v := c.Query("parentId"); v != "" {
		id, err := strconv.ParseUint(v, 10, 64)
		if err == nil {
			pid := uint(id)
			q.ParentID = &pid
		}
	}
	if v := c.Query("status"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			q.Status = &n
		}
	}
	if v := c.Query("holdStatus"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			q.HoldStatus = &n
		}
	}
	if v := c.Query("mutualHoldStatus"); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			q.MutualHoldStatus = &n
		}
	}
	list, err := m.app.ListMerchants(q)
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, list)
}

// Options 上级商户下拉。
func (m *MerchantController) Options(c *gin.Context) {
	opts, err := m.app.ListMerchantOptions()
	if err != nil {
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, opts)
}

// Create 新建商户（同时创建可登录账号）。
func (m *MerchantController) Create(c *gin.Context) {
	var req service.CreateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	operator := "system"
	if user, ok := middleware.CurrentUser(c); ok {
		operator = user.Username
	}
	item, err := m.app.CreateMerchant(req, operator)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMerchantNameInvalid):
			response.Fail(c, "商户名可由英文字母、数字、- 组成")
		case errors.Is(err, service.ErrMerchantNameExists):
			response.Fail(c, "商户名已存在")
		default:
			response.Fail(c, err.Error())
		}
		return
	}
	response.SuccessMsg(c, item, "created")
}

// SetStar 设置商户星标。
func (m *MerchantController) SetStar(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	var req struct {
		Starred bool `json:"starred"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	operator := "system"
	if user, ok := middleware.CurrentUser(c); ok {
		operator = user.Username
	}
	item, err := m.app.SetMerchantStarred(uint(id), req.Starred, operator)
	if err != nil {
		if errors.Is(err, service.ErrMerchantNotFound) {
			response.Fail(c, "商户不存在")
			return
		}
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, item)
}

// SetStatus 启用/禁用商户。
func (m *MerchantController) SetStatus(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	var req struct {
		Status bool `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	operator := "system"
	if user, ok := middleware.CurrentUser(c); ok {
		operator = user.Username
	}
	item, err := m.app.SetMerchantStatus(uint(id), req.Status, operator)
	if err != nil {
		if errors.Is(err, service.ErrMerchantNotFound) {
			response.Fail(c, "商户不存在")
			return
		}
		response.Fail(c, err.Error())
		return
	}
	response.Success(c, item)
}

// Update 编辑商户 / 用户信息。
func (m *MerchantController) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil || id == 0 {
		response.Fail(c, "参数错误")
		return
	}
	var req service.UpdateMerchantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "参数错误")
		return
	}
	operator := "system"
	if user, ok := middleware.CurrentUser(c); ok {
		operator = user.Username
	}
	item, err := m.app.UpdateMerchant(uint(id), req, operator)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrMerchantNotFound):
			response.Fail(c, "商户不存在")
		case errors.Is(err, service.ErrMerchantNameInvalid):
			response.Fail(c, "商户名可由英文字母、数字、- 组成")
		case errors.Is(err, service.ErrMerchantNameExists):
			response.Fail(c, "商户名已存在")
		case errors.Is(err, service.ErrMerchantParentInvalid):
			response.Fail(c, "上级商户无效")
		case errors.Is(err, service.ErrMerchantPasswordInvalid):
			response.Fail(c, "密码需为 6-20 位")
		default:
			response.Fail(c, err.Error())
		}
		return
	}
	response.Success(c, item)
}

// UploadAvatar 上传头像，返回可访问路径。
func (m *MerchantController) UploadAvatar(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "请选择图片")
		return
	}
	if file.Size > 2*1024*1024 {
		response.Fail(c, "图片不能超过 2MB")
		return
	}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
	default:
		response.Fail(c, "仅支持 jpg/png/gif/webp")
		return
	}
	if err := os.MkdirAll(filepath.Join("uploads", "avatars"), 0o755); err != nil {
		response.Fail(c, "上传失败")
		return
	}
	name := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	dst := filepath.Join("uploads", "avatars", name)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.Fail(c, "上传失败")
		return
	}
	response.Success(c, gin.H{"url": "/api/files/avatars/" + name})
}
