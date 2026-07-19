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

// ListDedicatedLines 获取专线信息列表（分页）
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

	// 总数
	var total int64
	query.Count(&total)

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var lines []models.DedicatedLine
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&lines).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 获取全量厂区列表（不受分页影响）
	var factories []string
	db.Model(&models.DedicatedLine{}).Distinct("factory").Order("factory").Pluck("factory", &factories)

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": lines, "total": total, "factories": factories})
}

// saveDedicatedLineImages 保存上传的图片文件，返回路径数组
func saveDedicatedLineImages(c *gin.Context) ([]string, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, nil // 没有文件上传
	}
	files := form.File["images"]
	if len(files) == 0 {
		return nil, nil
	}

	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	yearDir := filepath.Join("uploads", "dedicated_lines", fmt.Sprintf("%d", time.Now().Year()))
	os.MkdirAll(yearDir, 0755)

	var paths []string
	for _, file := range files {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if !allowedExts[ext] {
			continue
		}
		if file.Size > 10*1024*1024 {
			continue
		}
		filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
		filePath := filepath.Join(yearDir, filename)
		if err := c.SaveUploadedFile(file, filePath); err != nil {
			continue
		}
		paths = append(paths, filePath)
	}
	return paths, nil
}

// CreateDedicatedLine 创建专线信息（multipart/form-data）
func CreateDedicatedLine(c *gin.Context) {
	factory := strings.TrimSpace(c.PostForm("factory"))
	carrier := c.PostForm("carrier")
	ipStart := strings.TrimSpace(c.PostForm("ip_start"))
	ipEnd := strings.TrimSpace(c.PostForm("ip_end"))
	subnetMask := strings.TrimSpace(c.PostForm("subnet_mask"))
	gateway := strings.TrimSpace(c.PostForm("gateway"))

	if factory == "" || carrier == "" || ipStart == "" || ipEnd == "" || subnetMask == "" || gateway == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少必填字段"})
		return
	}

	bandwidthUp, _ := strconv.Atoi(c.PostForm("bandwidth_up"))
	bandwidthDown, _ := strconv.Atoi(c.PostForm("bandwidth_down"))
	dns := strings.TrimSpace(c.PostForm("dns"))
	notes := c.PostForm("notes")

	// 保存上传的图片
	imagePaths, _ := saveDedicatedLineImages(c)
	imagesJSON := "[]"
	if len(imagePaths) > 0 {
		data, _ := json.Marshal(imagePaths)
		imagesJSON = string(data)
	}

	username, displayName, approver := services.GetUserContext(c)

	line := models.DedicatedLine{
		Factory:       factory,
		Carrier:       carrier,
		BandwidthUp:   bandwidthUp,
		BandwidthDown: bandwidthDown,
		IPStart:       ipStart,
		IPEnd:         ipEnd,
		SubnetMask:    subnetMask,
		Gateway:       gateway,
		DNS:           dns,
		IPCount:       calcIPCount(ipStart, ipEnd),
		Images:        imagesJSON,
		Notes:         notes,
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

// UpdateDedicatedLine 更新专线信息（multipart/form-data）
func UpdateDedicatedLine(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var line models.DedicatedLine
	if err := database.GetDB().First(&line, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "专线信息不存在"})
		return
	}

	oldLine := line

	factory := strings.TrimSpace(c.PostForm("factory"))
	carrier := c.PostForm("carrier")
	ipStart := strings.TrimSpace(c.PostForm("ip_start"))
	ipEnd := strings.TrimSpace(c.PostForm("ip_end"))
	subnetMask := strings.TrimSpace(c.PostForm("subnet_mask"))
	gateway := strings.TrimSpace(c.PostForm("gateway"))

	if factory == "" || carrier == "" || ipStart == "" || ipEnd == "" || subnetMask == "" || gateway == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "缺少必填字段"})
		return
	}

	bandwidthUp, _ := strconv.Atoi(c.PostForm("bandwidth_up"))
	bandwidthDown, _ := strconv.Atoi(c.PostForm("bandwidth_down"))
	dns := strings.TrimSpace(c.PostForm("dns"))
	notes := c.PostForm("notes")
	existingImages := c.PostForm("existing_images") // 保留的已有图片路径 JSON

	// 解析保留的已有图片（白名单校验：仅允许本记录原有且在上传目录下的图片）
	var oldImages []string
	if oldLine.Images != "" && oldLine.Images != "[]" {
		json.Unmarshal([]byte(oldLine.Images), &oldImages)
	}
	oldSet := make(map[string]bool, len(oldImages))
	for _, o := range oldImages {
		oldSet[o] = true
	}
	var keptImages []string
	if existingImages != "" && existingImages != "[]" {
		var submitted []string
		if json.Unmarshal([]byte(existingImages), &submitted) == nil {
			for _, k := range submitted {
				if oldSet[k] && strings.HasPrefix(k, "uploads/dedicated_lines/") && !strings.Contains(k, "..") {
					keptImages = append(keptImages, k)
				}
			}
		}
	}

	// 保存新上传的图片
	newPaths, _ := saveDedicatedLineImages(c)

	// 合并：保留的旧图 + 新上传的图
	allImages := append(keptImages, newPaths...)
	imagesJSON := "[]"
	if len(allImages) > 0 {
		data, _ := json.Marshal(allImages)
		imagesJSON = string(data)
	}

	_, displayName, approver := services.GetUserContext(c)

	line.Factory = factory
	line.Carrier = carrier
	line.BandwidthUp = bandwidthUp
	line.BandwidthDown = bandwidthDown
	line.IPStart = ipStart
	line.IPEnd = ipEnd
	line.SubnetMask = subnetMask
	line.Gateway = gateway
	line.DNS = dns
	line.IPCount = calcIPCount(ipStart, ipEnd)
	line.Images = imagesJSON
	line.Notes = notes
	line.UpdatedBy = displayName

	if err := database.GetDB().Save(&line).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// DB 保存成功后再删除被移除的旧图片文件
	keptSet := make(map[string]bool, len(keptImages))
	for _, k := range keptImages {
		keptSet[k] = true
	}
	for _, old := range oldImages {
		if !keptSet[old] {
			os.Remove(old)
		}
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

// DeleteDedicatedLineImage 删除专线图片（同时更新记录）
func DeleteDedicatedLineImage(c *gin.Context) {
	var input struct {
		ImagePath string `json:"image_path" binding:"required"`
		LineID    uint   `json:"line_id"`
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

	// 如果提供了 line_id，更新记录的 images 字段
	if input.LineID > 0 {
		var line models.DedicatedLine
		if err := database.GetDB().First(&line, input.LineID).Error; err == nil {
			if line.Images != "" && line.Images != "[]" {
				var images []string
				if json.Unmarshal([]byte(line.Images), &images) == nil {
					var kept []string
					for _, img := range images {
						if img != input.ImagePath {
							kept = append(kept, img)
						}
					}
					newJSON := "[]"
					if len(kept) > 0 {
						data, _ := json.Marshal(kept)
						newJSON = string(data)
					}
					database.GetDB().Model(&line).Update("images", newJSON)
				}
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
