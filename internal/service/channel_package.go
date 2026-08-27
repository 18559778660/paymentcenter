package service

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

var (
	ErrChannelPackageNotFound = errors.New("channel package not found")
)

const channelPackageRoot = "channels"

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
	out, err := a.buildChannelListItem(*item)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// ChannelPackageFile 返回通道压缩包绝对路径与下载文件名。
func (a *App) ChannelPackageFile(id uint) (absPath, downloadName string, err error) {
	item, err := a.store.GetChannelByID(id)
	if err != nil {
		if isNotFound(err) {
			return "", "", ErrChannelNotFound
		}
		return "", "", err
	}
	relativePath := strings.TrimSpace(item.PackageName)
	if relativePath == "" {
		return "", "", ErrChannelPackageNotFound
	}
	absPath = channelPackageAbsPath(relativePath)
	if _, statErr := os.Stat(absPath); statErr != nil {
		if os.IsNotExist(statErr) {
			return "", "", ErrChannelPackageNotFound
		}
		return "", "", statErr
	}
	return absPath, path.Base(relativePath), nil
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
