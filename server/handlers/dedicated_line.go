package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// calcIPCount 根据起止IP计算可用IP数
func calcIPCount(ipStart, ipEnd string) int {
	start := net.ParseIP(ipStart)
	end := net.ParseIP(ipEnd)
	if start == nil || end == nil {
		return 0
	}
	startIP := start.To4()
	endIP := end.To4()
	if startIP == nil || endIP == nil {
		return 0
	}
	startNum := uint32(startIP[0])<<24 | uint32(startIP[1])<<16 | uint32(startIP[2])<<8 | uint32(startIP[3])
	endNum := uint32(endIP[0])<<24 | uint32(endIP[1])<<16 | uint32(endIP[2])<<8 | uint32(endIP[3])
	if endNum < startNum {
		return 0
	}
	return int(endNum - startNum + 1)
}

// ListDedicatedLines 获取专线信息列表
func ListDedicatedLines(c *gin.Context) {
	db := database.GetDB()
	query := db.Model(&models.DedicatedLine{})

	// 筛选条件
	if factory := c.Query("factory"); factory != "" {
		query = query.Where("factory LIKE ?", "%"+factory+"%")
	}
	if carrier := c.Query("carrier"); carrier != "" {
		query = query.Where("carrier = ?", carrier)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("factory LIKE ? OR carrier LIKE ? OR ip_start LIKE ? OR gateway LIKE ? OR notes LIKE ?", kw, kw, kw, kw, kw)
	}

	var lines []models.DedicatedLine
	if err := query.Order("created_at DESC").Find(&lines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": lines})
}

// CreateDedicatedLine 创建专线信息
func CreateDedicatedLine(c *gin.Context) {
	var input struct {
		Factory       string `json:"factory" binding:"required"`
		Carrier       string `json:"carrier" binding:"required"`
		BandwidthUp   int    `json:"bandwidth_up"`
		BandwidthDown int    `json:"bandwidth_down"`
		IPStart       string `json:"ip_start" binding:"required"`
		IPEnd         string `json:"ip_end" binding:"required"`
		SubnetMask    string `json:"subnet_mask" binding:"required"`
		Gateway       string `json:"gateway" binding:"required"`
		DNS           string `json:"dns"`
		Images        string `json:"images"`
		Notes         string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	username, displayName, approver := services.GetUserContext(c)

	line := models.DedicatedLine{
		Factory:       strings.TrimSpace(input.Factory),
		Carrier:       input.Carrier,
		BandwidthUp:   input.BandwidthUp,
		BandwidthDown: input.BandwidthDown,
		IPStart:       strings.TrimSpace(input.IPStart),
		IPEnd:         strings.TrimSpace(input.IPEnd),
		SubnetMask:    strings.TrimSpace(input.SubnetMask),
		Gateway:       strings.TrimSpace(input.Gateway),
		DNS:           strings.TrimSpace(input.DNS),
		IPCount:       calcIPCount(input.IPStart, input.IPEnd),
		Images:        input.Images,
		Notes:         input.Notes,
		CreatedBy:     displayName,
		UpdatedBy:     displayName,
	}

	if err := database.GetDB().Create(&line).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	// 记录操作日志
	details := []services.LogDetail{
		{FieldName: "Factory", FieldLabel: "厂区", NewValue: line.Factory},
		{FieldName: "Carrier", FieldLabel: "运营商", NewValue: line.Carrier},
		{FieldName: "IPStart", FieldLabel: "IP起始", NewValue: line.IPStart},
		{FieldName: "IPEnd", FieldLabel: "IP结束", NewValue: line.IPEnd},
	}
	services.LogOperation(username, displayName, "创建专线信息", "dedicated_line", line.ID, line.Factory+"-"+line.Carrier, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": line})
}

// UpdateDedicatedLine 更新专线信息
func UpdateDedicatedLine(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var line models.DedicatedLine
	if err := database.GetDB().First(&line, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "专线信息不存在"})
		return
	}

	oldLine := line

	var input struct {
		Factory       string `json:"factory" binding:"required"`
		Carrier       string `json:"carrier" binding:"required"`
		BandwidthUp   int    `json:"bandwidth_up"`
		BandwidthDown int    `json:"bandwidth_down"`
		IPStart       string `json:"ip_start" binding:"required"`
		IPEnd         string `json:"ip_end" binding:"required"`
		SubnetMask    string `json:"subnet_mask" binding:"required"`
		Gateway       string `json:"gateway" binding:"required"`
		DNS           string `json:"dns"`
		Images        string `json:"images"`
		Notes         string `json:"notes"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}

	_, displayName, approver := services.GetUserContext(c)

	line.Factory = strings.TrimSpace(input.Factory)
	line.Carrier = input.Carrier
	line.BandwidthUp = input.BandwidthUp
	line.BandwidthDown = input.BandwidthDown
	line.IPStart = strings.TrimSpace(input.IPStart)
	line.IPEnd = strings.TrimSpace(input.IPEnd)
	line.SubnetMask = strings.TrimSpace(input.SubnetMask)
	line.Gateway = strings.TrimSpace(input.Gateway)
	line.DNS = strings.TrimSpace(input.DNS)
	line.IPCount = calcIPCount(input.IPStart, input.IPEnd)
	line.Images = input.Images
	line.Notes = input.Notes
	line.UpdatedBy = displayName

	if err := database.GetDB().Save(&line).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// 记录操作日志
	username, _, _ := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Factory", FieldLabel: "厂区", OldValue: oldLine.Factory, NewValue: line.Factory},
		{FieldName: "Carrier", FieldLabel: "运营商", OldValue: oldLine.Carrier, NewValue: line.Carrier},
		{FieldName: "IPStart", FieldLabel: "IP起始", OldValue: oldLine.IPStart, NewValue: line.IPStart},
		{FieldName: "IPEnd", FieldLabel: "IP结束", OldValue: oldLine.IPEnd, NewValue: line.IPEnd},
	}
	services.LogOperation(username, displayName, "更新专线信息", "dedicated_line", line.ID, line.Factory+"-"+line.Carrier, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": line})
}

// DeleteDedicatedLine 删除专线信息
func DeleteDedicatedLine(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var line models.DedicatedLine
	if err := database.GetDB().First(&line, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "专线信息不存在"})
		return
	}

	// 删除关联图片文件
	if line.Images != "" && line.Images != "[]" {
		var images []string
		if err := json.Unmarshal([]byte(line.Images), &images); err == nil {
			for _, img := range images {
				os.Remove(img)
			}
		}
	}

	if err := database.GetDB().Delete(&line).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 记录操作日志
	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Factory", FieldLabel: "厂区", OldValue: line.Factory},
		{FieldName: "Carrier", FieldLabel: "运营商", OldValue: line.Carrier},
	}
	services.LogOperation(username, displayName, "删除专线信息", "dedicated_line", line.ID, line.Factory+"-"+line.Carrier, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// UploadDedicatedLineImage 上传专线图片
func UploadDedicatedLineImage(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请上传图片文件"})
		return
	}

	// 校验文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	if !allowedExts[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "仅支持 JPG、PNG、GIF、WebP 格式"})
		return
	}
	if file.Size > 10*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "图片大小不能超过 10MB"})
		return
	}

	// 存储到 uploads/dedicated_lines/年份/
	yearDir := filepath.Join("uploads", "dedicated_lines", fmt.Sprintf("%d", time.Now().Year()))
	os.MkdirAll(yearDir, 0755)

	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "文件保存失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "上传成功", "data": gin.H{
		"path": filePath,
		"url":  "/" + filePath,
		"name": file.Filename,
	}})
}

// DeleteDedicatedLineImage 删除专线图片
func DeleteDedicatedLineImage(c *gin.Context) {
	var input struct {
		ImagePath string `json:"image_path" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 安全检查：确保路径在 uploads/dedicated_lines 下
	if !strings.HasPrefix(input.ImagePath, "uploads/dedicated_lines") {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的文件路径"})
		return
	}

	os.Remove(input.ImagePath)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
