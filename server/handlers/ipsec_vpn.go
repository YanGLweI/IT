package handlers

import (
	"encoding/json"
	"fmt"
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

// Phase2Entry 阶段二条目
type Phase2Entry struct {
	LocalAddr  string `json:"local_addr"`
	RemoteAddr string `json:"remote_addr"`
	Image      string `json:"image"`
}

// saveIPsecVpnImage 保存单张截图
func saveIPsecVpnImage(c *gin.Context, fieldName string) string {
	file, err := c.FormFile(fieldName)
	if err != nil || file == nil {
		return ""
	}
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true}
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !allowedExts[ext] {
		return ""
	}
	if file.Size > 10*1024*1024 {
		return ""
	}
	yearDir := filepath.Join("uploads", "ipsec_vpn", fmt.Sprintf("%d", time.Now().Year()))
	os.MkdirAll(yearDir, 0755)
	filename := fmt.Sprintf("%d_%s%s", time.Now().UnixNano(), strings.TrimSuffix(filepath.Base(file.Filename), ext), ext)
	filePath := filepath.Join(yearDir, filename)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return ""
	}
	return filePath
}

// isValidIPsecImagePath 校验图片路径安全性
func isValidIPsecImagePath(path string) bool {
	return path != "" && strings.HasPrefix(path, "uploads/ipsec_vpn/") && !strings.Contains(path, "..")
}

// ListIPsecVpns 获取IPsec VPN列表（分页）
func ListIPsecVpns(c *gin.Context) {
	db := database.GetDB()
	query := db.Model(&models.IPsecVpn{})

	if keyword := c.Query("keyword"); keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("tunnel_name LIKE ? OR remote_ip LIKE ? OR local_ip LIKE ? OR owner LIKE ?", kw, kw, kw, kw)
	}

	var total int64
	query.Count(&total)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	offset := (page - 1) * pageSize

	var list []models.IPsecVpn
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list, "total": total})
}

// CreateIPsecVpn 创建IPsec VPN记录
func CreateIPsecVpn(c *gin.Context) {
	db := database.GetDB()

	tunnelName := c.PostForm("tunnel_name")
	owner := c.PostForm("owner")
	remoteIP := c.PostForm("remote_ip")
	localIP := c.PostForm("local_ip")
	psk := c.PostForm("psk")
	ikeVersion, _ := strconv.Atoi(c.PostForm("ike_version"))
	mode := c.PostForm("mode")
	phase2JSON := c.PostForm("phase2_entries")

	// 必填校验
	if tunnelName == "" || owner == "" || remoteIP == "" || localIP == "" || psk == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写所有必填字段"})
		return
	}
	if ikeVersion != 1 && ikeVersion != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "IKE版本必须为1或2"})
		return
	}
	if ikeVersion == 1 && mode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "IKE v1时必须选择模式"})
		return
	}

	// 保存截图
	networkImage := saveIPsecVpnImage(c, "network_image")
	phase1Image := saveIPsecVpnImage(c, "phase1_image")

	// 解析阶段二
	var phase2Entries []Phase2Entry
	if phase2JSON != "" {
		json.Unmarshal([]byte(phase2JSON), &phase2Entries)
	}
	// 保存阶段二截图
	for i := range phase2Entries {
		img := saveIPsecVpnImage(c, fmt.Sprintf("phase2_image_%d", i))
		if img != "" {
			phase2Entries[i].Image = img
		}
	}
	phase2Bytes, _ := json.Marshal(phase2Entries)

	username, displayName, approver := services.GetUserContext(c)
	record := models.IPsecVpn{
		TunnelName:    tunnelName,
		Owner:         owner,
		RemoteIP:      remoteIP,
		LocalIP:       localIP,
		NetworkImage:  networkImage,
		PSK:           psk,
		IKEVersion:    ikeVersion,
		Mode:          mode,
		Phase1Image:   phase1Image,
		Phase2Entries: string(phase2Bytes),
		CreatedBy:     displayName,
		UpdatedBy:     displayName,
	}

	if err := db.Create(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}

	services.LogOperation(username, displayName, "创建IPsec VPN", "ipsec_vpn", record.ID, tunnelName, approver, c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": record})
}

// UpdateIPsecVpn 更新IPsec VPN记录
func UpdateIPsecVpn(c *gin.Context) {
	db := database.GetDB()
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var record models.IPsecVpn
	if err := db.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	tunnelName := c.PostForm("tunnel_name")
	owner := c.PostForm("owner")
	remoteIP := c.PostForm("remote_ip")
	localIP := c.PostForm("local_ip")
	psk := c.PostForm("psk")
	ikeVersion, _ := strconv.Atoi(c.PostForm("ike_version"))
	mode := c.PostForm("mode")
	phase2JSON := c.PostForm("phase2_entries")
	existingNetworkImage := c.PostForm("existing_network_image")
	existingPhase1Image := c.PostForm("existing_phase1_image")

	if tunnelName == "" || owner == "" || remoteIP == "" || localIP == "" || psk == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请填写所有必填字段"})
		return
	}
	if ikeVersion != 1 && ikeVersion != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "IKE版本必须为1或2"})
		return
	}
	if ikeVersion == 1 && mode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "IKE v1时必须选择模式"})
		return
	}

	// 处理网络截图：新上传 > 保留旧图
	var oldNetworkImage, oldPhase1Image string
	networkImage := saveIPsecVpnImage(c, "network_image")
	if networkImage == "" {
		if isValidIPsecImagePath(existingNetworkImage) && existingNetworkImage == record.NetworkImage {
			networkImage = existingNetworkImage
		}
		oldNetworkImage = record.NetworkImage
	} else {
		oldNetworkImage = record.NetworkImage
	}

	// 处理阶段一截图
	phase1Image := saveIPsecVpnImage(c, "phase1_image")
	if phase1Image == "" {
		if isValidIPsecImagePath(existingPhase1Image) && existingPhase1Image == record.Phase1Image {
			phase1Image = existingPhase1Image
		}
		oldPhase1Image = record.Phase1Image
	} else {
		oldPhase1Image = record.Phase1Image
	}

	// 解析旧阶段二（用于白名单校验）
	var oldPhase2 []Phase2Entry
	if record.Phase2Entries != "" && record.Phase2Entries != "[]" {
		json.Unmarshal([]byte(record.Phase2Entries), &oldPhase2)
	}
	oldPhase2Images := make(map[string]bool)
	for _, p := range oldPhase2 {
		if p.Image != "" {
			oldPhase2Images[p.Image] = true
		}
	}

	// 解析新阶段二
	var phase2Entries []Phase2Entry
	if phase2JSON != "" {
		json.Unmarshal([]byte(phase2JSON), &phase2Entries)
	}
	var removedPhase2Images []string
	for i := range phase2Entries {
		img := saveIPsecVpnImage(c, fmt.Sprintf("phase2_image_%d", i))
		if img != "" {
			phase2Entries[i].Image = img
		} else if phase2Entries[i].Image != "" {
			// 保留旧图 - 白名单校验
			if !oldPhase2Images[phase2Entries[i].Image] || !isValidIPsecImagePath(phase2Entries[i].Image) {
				phase2Entries[i].Image = ""
			}
		}
	}
	// 找出被移除的旧阶段二图片
	newPhase2Images := make(map[string]bool)
	for _, p := range phase2Entries {
		if p.Image != "" {
			newPhase2Images[p.Image] = true
		}
	}
	for _, old := range oldPhase2 {
		if old.Image != "" && !newPhase2Images[old.Image] {
			removedPhase2Images = append(removedPhase2Images, old.Image)
		}
	}

	phase2Bytes, _ := json.Marshal(phase2Entries)
	username, displayName, approver := services.GetUserContext(c)

	record.TunnelName = tunnelName
	record.Owner = owner
	record.RemoteIP = remoteIP
	record.LocalIP = localIP
	record.NetworkImage = networkImage
	record.PSK = psk
	record.IKEVersion = ikeVersion
	record.Mode = mode
	record.Phase1Image = phase1Image
	record.Phase2Entries = string(phase2Bytes)
	record.UpdatedBy = displayName

	if err := db.Save(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}

	// DB保存成功后删除旧文件
	if oldNetworkImage != "" && oldNetworkImage != networkImage && isValidIPsecImagePath(oldNetworkImage) {
		os.Remove(oldNetworkImage)
	}
	if oldPhase1Image != "" && oldPhase1Image != phase1Image && isValidIPsecImagePath(oldPhase1Image) {
		os.Remove(oldPhase1Image)
	}
	for _, img := range removedPhase2Images {
		if isValidIPsecImagePath(img) {
			os.Remove(img)
		}
	}

	services.LogOperation(username, displayName, "更新IPsec VPN", "ipsec_vpn", record.ID, tunnelName, approver, c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": record})
}

// DeleteIPsecVpn 删除IPsec VPN记录
func DeleteIPsecVpn(c *gin.Context) {
	db := database.GetDB()
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "无效的ID"})
		return
	}

	var record models.IPsecVpn
	if err := db.First(&record, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "记录不存在"})
		return
	}

	if err := db.Delete(&record).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}

	// 清理关联图片
	if isValidIPsecImagePath(record.NetworkImage) {
		os.Remove(record.NetworkImage)
	}
	if isValidIPsecImagePath(record.Phase1Image) {
		os.Remove(record.Phase1Image)
	}
	var phase2 []Phase2Entry
	if record.Phase2Entries != "" && record.Phase2Entries != "[]" {
		json.Unmarshal([]byte(record.Phase2Entries), &phase2)
		for _, p := range phase2 {
			if isValidIPsecImagePath(p.Image) {
				os.Remove(p.Image)
			}
		}
	}

	username, displayName, approver := services.GetUserContext(c)
	services.LogOperation(username, displayName, "删除IPsec VPN", "ipsec_vpn", record.ID, record.TunnelName, approver, c.ClientIP(), nil)
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}
