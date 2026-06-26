package handlers

import (
	"net/http"
	"strconv"

	"it-platform-server/database"
	"it-platform-server/models"

	"github.com/gin-gonic/gin"
)

// ListApprovedSoftware 获取核准软件列表
func ListApprovedSoftware(c *gin.Context) {
	var list []models.ApprovedSoftware
	if err := database.GetDB().Order("id desc").Find(&list).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

// CreateApprovedSoftware 创建核准软件
func CreateApprovedSoftware(c *gin.Context) {
	var sw models.ApprovedSoftware
	if err := c.ShouldBindJSON(&sw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误: " + err.Error()})
		return
	}
	if err := database.GetDB().Create(&sw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": sw})
}

// UpdateApprovedSoftware 更新核准软件
func UpdateApprovedSoftware(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var sw models.ApprovedSoftware
	if err := database.GetDB().First(&sw, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "软件不存在"})
		return
	}

	var input struct {
		Name          string `json:"name" binding:"required"`
		Version       string `json:"version"`
		LatestVersion string `json:"latest_version"`
		NeedUpdate    bool   `json:"need_update"`
		UpdateReason  string `json:"update_reason"`
		Vendor        string `json:"vendor"`
		VendorWebsite string `json:"vendor_website"`
		LicenseType   string `json:"license_type"`
		Purpose       string `json:"purpose"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	sw.Name = input.Name
	sw.Version = input.Version
	sw.LatestVersion = input.LatestVersion
	sw.NeedUpdate = input.NeedUpdate
	sw.UpdateReason = input.UpdateReason
	sw.Vendor = input.Vendor
	sw.VendorWebsite = input.VendorWebsite
	sw.LicenseType = input.LicenseType
	sw.Purpose = input.Purpose

	if err := database.GetDB().Save(&sw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": sw})
}

// DeleteApprovedSoftware 删除核准软件
func DeleteApprovedSoftware(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var sw models.ApprovedSoftware
	if err := database.GetDB().First(&sw, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "软件不存在"})
		return
	}
	// 删除关联记录
	database.GetDB().Where("approved_software_id = ?", id).Delete(&models.AssetSoftware{})
	if err := database.GetDB().Unscoped().Delete(&sw).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ListAssetSoftware 获取资产与软件关联列表（展示所有资产，附带已关联软件信息）
func ListAssetSoftware(c *gin.Context) {
	// 分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	// 查询资产总数
	var total int64
	database.GetDB().Model(&models.Asset{}).Count(&total)

	// 分页查询资产
	var assets []models.Asset
	offset := (page - 1) * pageSize
	if err := database.GetDB().Order("id desc").Offset(offset).Limit(pageSize).Find(&assets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 为每个资产查询关联的核准软件
	type AssetRow struct {
		ID           uint                      `json:"id"`
		ComputerName string                    `json:"computer_name"`
		IPAddress    string                    `json:"ip_address"`
		SoftwareList []models.ApprovedSoftware `json:"software_list"`
	}

	var result []AssetRow
	for _, a := range assets {
		row := AssetRow{
			ID:           a.ID,
			ComputerName: a.ComputerName,
			IPAddress:    a.IPAddress,
		}
		var links []models.AssetSoftware
		database.GetDB().Where("asset_id = ?", a.ID).Find(&links)
		var swList []models.ApprovedSoftware
		for _, l := range links {
			var sw models.ApprovedSoftware
			if err := database.GetDB().First(&sw, l.ApprovedSoftwareID).Error; err == nil {
				swList = append(swList, sw)
			}
		}
		if swList == nil {
			swList = []models.ApprovedSoftware{}
		}
		row.SoftwareList = swList
		result = append(result, row)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":  200,
		"data":  result,
		"total": total,
		"page":  page,
		"page_size": pageSize,
	})
}

// GetAssetSoftwareLinks 获取某资产的已关联软件ID列表
func GetAssetSoftwareLinks(c *gin.Context) {
	assetID, _ := strconv.Atoi(c.Param("id"))
	var links []models.AssetSoftware
	database.GetDB().Where("asset_id = ?", assetID).Find(&links)
	var ids []uint
	for _, l := range links {
		ids = append(ids, l.ApprovedSoftwareID)
	}
	if ids == nil {
		ids = []uint{}
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ids})
}

// UpdateAssetSoftwareLinks 更新资产的软件关联（全量替换）
func UpdateAssetSoftwareLinks(c *gin.Context) {
	assetID, _ := strconv.Atoi(c.Param("id"))

	// 验证资产存在
	var asset models.Asset
	if err := database.GetDB().First(&asset, assetID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "资产不存在"})
		return
	}

	var input struct {
		SoftwareIDs []uint `json:"software_ids"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数错误"})
		return
	}

	// 删除旧关联
	database.GetDB().Where("asset_id = ?", assetID).Delete(&models.AssetSoftware{})

	// 创建新关联
	for _, swID := range input.SoftwareIDs {
		link := models.AssetSoftware{
			AssetID:            uint(assetID),
			ApprovedSoftwareID: swID,
		}
		database.GetDB().Create(&link)
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功"})
}
