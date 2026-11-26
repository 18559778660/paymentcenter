package service

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"paymentcenter/internal/model"
)

var (
	ErrChannelPackageNotFound = errors.New("channel package not found")
)

const (
	channelPackageRoot         = "channels"
	channelPackageDefaultsRoot = "resources/channel-packages"
)

// BuildPackageURL 生成压缩包下载地址（走鉴权接口）。
func (a *App) BuildPackageURL(channelID uint, packagePath string) string {
	if channelID == 0 || strings.TrimSpace(packagePath) == "" {
		return ""
	}
	return fmt.Sprintf("/api/channels/%d/package", channelID)
}

// SetChannelPackage 绑定通道压缩包相对路径。
func (a *App) SetChannelPackage(id uint, relativePath, operator string) (*ChannelListItem, error) {
	item, err := a.store.GetChannelByID(id)
	if err != nil {
		if isNotFound(err) {
			return nil, ErrChannelNotFound
		}
		return nil, err
	}
	oldPath := item.PackageName
	item.PackageName = relativePath
	item.UpdatedBy = operator
	if err := a.store.SaveChannel(item); err != nil {
		return nil, err
	}
	if oldPath != "" && oldPath != relativePath {
		removeChannelPackageFile(oldPath)
	}
	out, err := a.buildChannelListItem(*item, nil)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelPackageFile 返回通道压缩包绝对路径与下载文件名。
// 已上传文件不存在时，按平台回落到 resources/channel-packages/{stripe|paypal}/ 下的默认包。
func (a *App) ChannelPackageFile(id uint) (absPath, downloadName string, err error) {
	item, err := a.store.GetChannelByID(id)
	if err != nil {
		if isNotFound(err) {
			return "", "", ErrChannelNotFound
		}
		return "", "", err
	}
	platformCode := ""
	if platform, platErr := a.getPlatformByID(item.PlatformID); platErr == nil && platform != nil {
		platformCode = platform.Code
	}
	absPath, downloadName, ok := lookupChannelPackage(*item, platformCode)
	if !ok {
		return "", "", ErrChannelPackageNotFound
	}
	return absPath, downloadName, nil
}

// BuildChannelPackageRelativePath 生成压缩包存储相对路径。
func BuildChannelPackageRelativePath(channelID uint, filename string) string {
	return path.Join(channelPackageRoot, fmt.Sprintf("%d", channelID), sanitizePackageFilename(filename))
}

func channelPackageAbsPath(relativePath string) string {
	return filepath.Join("uploads", filepath.FromSlash(relativePath))
}

func removeChannelPackageFile(relativePath string) {
	if relativePath == "" {
		return
	}
	_ = os.Remove(channelPackageAbsPath(relativePath))
}

// removeChannelPackageAssets 删除通道关联的压缩包及 uploads/channels/{id} 目录。
func removeChannelPackageAssets(channelID uint, relativePath string) {
	removeChannelPackageFile(relativePath)
	if channelID == 0 {
		return
	}
	dir := filepath.Join("uploads", channelPackageRoot, fmt.Sprintf("%d", channelID))
	_ = os.RemoveAll(dir)
}

func sanitizePackageFilename(name string) string {
	name = filepath.Base(strings.TrimSpace(name))
	var builder strings.Builder
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '.', r == '-', r == '_':
			builder.WriteRune(r)
		}
	}
	safeName := builder.String()
	if safeName == "" || safeName == "." {
		return "package.zip"
	}
	return safeName
}

func packageDisplayName(relativePath string) string {
	name := path.Base(relativePath)
	if name == "" || name == "." {
		return ""
	}
	return name
}

// lookupChannelPackage 查找通道压缩包。
func lookupChannelPackage(item model.Channel, platformCode string) (absPath, downloadName string, ok bool) {
	relativePath := strings.TrimSpace(item.PackageName)
	if relativePath != "" {
		absPath = channelPackageAbsPath(relativePath)
		if _, err := os.Stat(absPath); err == nil {
			return absPath, packageDisplayName(relativePath), true
		}
	}
	return lookupDefaultChannelPackage(platformCode)
}

// lookupDefaultChannelPackage 查找默认通道压缩包。
func lookupDefaultChannelPackage(platformCode string) (absPath, downloadName string, ok bool) {
	code := strings.ToLower(strings.TrimSpace(platformCode))
	if code == "" {
		return "", "", false
	}
	dir := filepath.Join("resources", "channel-packages", code)
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", false
	}
	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".zip" && ext != ".rar" && ext != ".7z" {
			continue
		}
		names = append(names, entry.Name())
	}
	if len(names) == 0 {
		return "", "", false
	}
	sort.Strings(names)
	return filepath.Join(dir, names[0]), names[0], true
}
