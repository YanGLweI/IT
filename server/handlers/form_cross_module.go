package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// CrossModuleSource 跨模块引用源定义
type CrossModuleSource struct {
	ModuleKey  string `json:"module_key"`
	ModuleName string `json:"module_name"`
	SourceType string `json:"source_type"` // "static" 或 "dynamic"
}

// CrossModuleFile 跨模块可引用文件
type CrossModuleFile struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Version  string `json:"version"`
	FileName string `json:"file_name"`
	FilePath string `json:"file_path"`
}

// getCrossModuleRegistry 获取可引用模块列表
func getCrossModuleRegistry() []CrossModuleSource {
	return []CrossModuleSource{
		{ModuleKey: "change_record_template", ModuleName: "变更记录表模板", SourceType: "static"},
		{ModuleKey: "backup_template", ModuleName: "备份与恢复记录表模板", SourceType: "static"},
		{ModuleKey: "user_change_record", ModuleName: "用户变更记录表（动态生成）", SourceType: "dynamic"},
		{ModuleKey: "department_confirmation", ModuleName: "部门用户确认表（动态生成）", SourceType: "dynamic"},
	}
}

// ListCrossModuleSources 获取可引用的跨模块资源列表
func ListCrossModuleSources(c *gin.Context) {
	registry := getCrossModuleRegistry()
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": registry})
}

// ListCrossModuleFiles 获取指定模块内的可引用文件列表
func ListCrossModuleFiles(c *gin.Context) {
	moduleKey := c.Param("module")
	var files []CrossModuleFile

	switch moduleKey {
	case "change_record_template":
		var templates []models.ChangeRecordTemplate
		database.GetDB().Order("is_current DESC, created_at DESC").Find(&templates)
		for _, t := range templates {
			files = append(files, CrossModuleFile{
				ID:       t.ID,
				Name:     t.Version + " - " + t.Description,
				Version:  t.Version,
				FileName: t.FileName,
				FilePath: t.FilePath,
			})
		}

	case "backup_template":
		var templates []models.BackupTemplate
		database.GetDB().Order("created_at DESC").Find(&templates)
		for _, t := range templates {
			files = append(files, CrossModuleFile{
				ID:       t.ID,
				Name:     t.FileName,
				FileName: t.FileName,
				FilePath: t.FilePath,
			})
		}

	case "user_change_record":
		// 动态类型，不需要文件列表
		files = append(files, CrossModuleFile{
			ID:   0,
			Name: "用户变更记录表（实时生成）",
		})

	case "department_confirmation":
		// 动态类型，不需要文件列表
		files = append(files, CrossModuleFile{
			ID:   0,
			Name: "部门用户确认表（实时生成）",
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的模块: " + moduleKey})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": files})
}

// CreateCrossModuleRef 创建跨模块引用
func CreateCrossModuleRef(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	category := c.PostForm("category")
	moduleKey := c.PostForm("module_key")
	sourceType := c.PostForm("source_type")

	if title == "" || moduleKey == "" || sourceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "标题、模块和来源类型不能为空"})
		return
	}

	// 查找模块定义
	var foundSource *CrossModuleSource
	registry := getCrossModuleRegistry()
	for i := range registry {
		if registry[i].ModuleKey == moduleKey {
			foundSource = &registry[i]
			break
		}
	}
	if foundSource == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的模块"})
		return
	}

	item := models.FormVaultItem{
		Title:       title,
		Description: description,
		Category:    category,
		SourceType:  sourceType,
		RefModule:   moduleKey,
	}

	switch sourceType {
	case models.SourceTypeStatic:
		// 静态引用 - 复制文件到快照目录
		refIDStr := c.PostForm("ref_id")
		refID, _ := strconv.Atoi(refIDStr)

		// 获取源文件路径
		var srcPath, srcFileName string
		switch moduleKey {
		case "change_record_template":
			var tpl models.ChangeRecordTemplate
			if err := database.GetDB().First(&tpl, refID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "模板不存在"})
				return
			}
			srcPath = tpl.FilePath
			srcFileName = tpl.FileName
			item.RefID = uint(refID)

		case "backup_template":
			var tpl models.BackupTemplate
			if err := database.GetDB().First(&tpl, refID).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "模板不存在"})
				return
			}
			srcPath = tpl.FilePath
			srcFileName = tpl.FileName
			item.RefID = uint(refID)

		default:
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该模块不支持静态引用"})
			return
		}

		if _, err := os.Stat(srcPath); os.IsNotExist(err) {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "源文件不存在"})
			return
		}

		// 复制文件到快照目录
		snapshotDir := filepath.Join(config.Cfg.Upload.FormVaultSnapshotPath, time.Now().Format("2006"))
		os.MkdirAll(snapshotDir, 0755)

		ext := filepath.Ext(srcFileName)
		snapshotName := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), filepath.Base(srcFileName), "")
		snapshotName = snapshotName[:len(snapshotName)-len(ext)] + ext
		snapshotPath := filepath.Join(snapshotDir, snapshotName)

		if err := copyFile(srcPath, snapshotPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建快照失败: " + err.Error()})
			return
		}

		item.FileName = srcFileName
		item.SnapshotPath = snapshotPath
		// 获取文件大小
		if info, err := os.Stat(snapshotPath); err == nil {
			item.FileSize = info.Size()
		}

	case models.SourceTypeDynamic:
		// 动态引用 - 仅保存引用信息
		handlerName := ""
		switch moduleKey {
		case "user_change_record":
			handlerName = "export_user_change_record"
			item.FileName = "用户变更记录表.xlsx"
		case "department_confirmation":
			handlerName = "export_department_confirmation"
			item.FileName = "部门用户确认表.xlsx"
		default:
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该模块不支持动态引用"})
			return
		}
		item.RefHandler = handlerName

	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "不支持的来源类型"})
		return
	}

	if err := database.GetDB().Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建引用失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Title", FieldLabel: "标题", NewValue: title},
		{FieldName: "SourceType", FieldLabel: "来源类型", NewValue: sourceType},
		{FieldName: "RefModule", FieldLabel: "引用模块", NewValue: moduleKey},
	}
	services.LogOperation(username, displayName, "创建跨模块引用", "form_vault", item.ID, item.Title, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建引用成功", "data": item})
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
