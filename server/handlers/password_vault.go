package handlers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"it-platform-server/config"
	"it-platform-server/database"
	"it-platform-server/models"
	"it-platform-server/services"

	"github.com/gin-gonic/gin"
)

// ============ AES 加密/解密 ============

// aesEncrypt AES-256-GCM 加密
func aesEncrypt(plaintext string) (string, error) {
	keyHex := config.Cfg.PasswordVault.AESKey
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("AES密钥无效")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %v", err)
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("生成随机IV失败: %v", err)
	}

	ciphertext := aesGCM.Seal(nil, nonce, []byte(plaintext), nil)

	ivHex := hex.EncodeToString(nonce)
	ctHex := hex.EncodeToString(ciphertext)
	return ivHex + ":" + ctHex, nil
}

// aesDecrypt AES-256-GCM 解密
func aesDecrypt(encrypted string) (string, error) {
	keyHex := config.Cfg.PasswordVault.AESKey
	key, err := hex.DecodeString(keyHex)
	if err != nil || len(key) != 32 {
		return "", fmt.Errorf("AES密钥无效")
	}

	parts := strings.SplitN(encrypted, ":", 2)
	if len(parts) != 2 {
		return "", fmt.Errorf("密文格式无效")
	}

	nonce, err := hex.DecodeString(parts[0])
	if err != nil {
		return "", fmt.Errorf("解码IV失败: %v", err)
	}

	ciphertext, err := hex.DecodeString(parts[1])
	if err != nil {
		return "", fmt.Errorf("解码密文失败: %v", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("创建AES cipher失败: %v", err)
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("创建GCM失败: %v", err)
	}

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("解密失败: %v", err)
	}

	return string(plaintext), nil
}

// ============ 分类管理 ============

// ListPasswordCategories 获取密码分类列表
func ListPasswordCategories(c *gin.Context) {
	db := database.GetDB()

	// 获取当前用户
	currentUsername, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未获取到用户信息"})
		return
	}
	usernameStr := currentUsername.(string)

	var categories []models.PasswordCategory
	if err := db.Order("sort_order ASC, name ASC").Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询分类失败"})
		return
	}

	// 一次性查询所有分类的条目数量（优化 N+1）
	type catCount struct {
		CategoryID uint
		Count      int64
	}
	var counts []catCount
	db.Model(&models.PasswordEntry{}).
		Select("category_id, COUNT(*) as count").
		Where("id IN (SELECT entry_id FROM password_entry_viewers WHERE username = ?)", usernameStr).
		Group("category_id").
		Scan(&counts)

	countMap := make(map[uint]int64)
	for _, cc := range counts {
		countMap[cc.CategoryID] = cc.Count
	}
	for i := range categories {
		categories[i].EntryCount = countMap[categories[i].ID]
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": categories})
}

// CreatePasswordCategory 创建密码分类
func CreatePasswordCategory(c *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
		Icon string `json:"icon" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类名称和图标不能为空"})
		return
	}

	category := models.PasswordCategory{
		Name: req.Name,
		Icon: req.Icon,
	}

	if err := database.GetDB().Create(&category).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类名称已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建分类失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "分类名称", NewValue: req.Name},
		{FieldName: "Icon", FieldLabel: "图标", NewValue: req.Icon},
	}
	services.LogOperation(username, displayName, "创建密码分类", "password_category", category.ID, category.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": category})
}

// UpdatePasswordCategory 更新密码分类
func UpdatePasswordCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var category models.PasswordCategory
	if err := database.GetDB().First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "分类不存在"})
		return
	}

	oldCategory := category

	var req struct {
		Name string `json:"name" binding:"required"`
		Icon string `json:"icon" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数无效"})
		return
	}

	category.Name = req.Name
	category.Icon = req.Icon

	if err := database.GetDB().Save(&category).Error; err != nil {
		if strings.Contains(err.Error(), "Duplicate") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类名称已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新分类失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	fieldLabels := services.GetFieldLabels("password_category")
	details := services.DiffStructs(oldCategory, category, fieldLabels)
	services.LogOperation(username, displayName, "更新密码分类", "password_category", category.ID, category.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": category})
}

// DeletePasswordCategory 删除密码分类
func DeletePasswordCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var category models.PasswordCategory
	if err := database.GetDB().First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "分类不存在"})
		return
	}

	if category.IsPreset {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "预设分类不可删除"})
		return
	}

	// 检查分类下是否有密码条目
	var count int64
	database.GetDB().Model(&models.PasswordEntry{}).Where("category_id = ?", id).Count(&count)
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "该分类下有密码条目，不可删除"})
		return
	}

	if err := database.GetDB().Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除分类失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "分类名称", OldValue: category.Name},
	}
	services.LogOperation(username, displayName, "删除密码分类", "password_category", category.ID, category.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// SortPasswordCategory 调整分类排序（上移/下移）
func SortPasswordCategory(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	db := database.GetDB()

	var category models.PasswordCategory
	if err := db.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "分类不存在"})
		return
	}

	var req struct {
		Direction string `json:"direction" binding:"required"` // "up" or "down"
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数无效"})
		return
	}

	// 获取所有分类按 sort_order 排序
	var categories []models.PasswordCategory
	db.Order("sort_order ASC, id ASC").Find(&categories)

	// 找到当前分类的索引
	idx := -1
	for i, cat := range categories {
		if cat.ID == uint(id) {
			idx = i
			break
		}
	}
	if idx == -1 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "分类位置异常"})
		return
	}

	if req.Direction == "up" {
		if idx == 0 {
			c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已在最顶部"})
			return
		}
		// 与上一个交换 sort_order
		categories[idx].SortOrder, categories[idx-1].SortOrder = categories[idx-1].SortOrder, categories[idx].SortOrder
		db.Save(&categories[idx])
		db.Save(&categories[idx-1])
	} else if req.Direction == "down" {
		if idx == len(categories)-1 {
			c.JSON(http.StatusOK, gin.H{"code": 200, "message": "已在最底部"})
			return
		}
		// 与下一个交换 sort_order
		categories[idx].SortOrder, categories[idx+1].SortOrder = categories[idx+1].SortOrder, categories[idx].SortOrder
		db.Save(&categories[idx])
		db.Save(&categories[idx+1])
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "direction 必须为 up 或 down"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "排序成功"})
}

// ============ 密码条目管理 ============

// ListPasswordEntries 获取密码条目列表
func ListPasswordEntries(c *gin.Context) {
	db := database.GetDB()

	// 获取当前用户
	currentUsername, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未获取到用户信息"})
		return
	}
	usernameStr := currentUsername.(string)

	query := db.Model(&models.PasswordEntry{})

	// 只返回当前用户被授权的条目
	query = query.Where("id IN (SELECT entry_id FROM password_entry_viewers WHERE username = ?)", usernameStr)

	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	if isStarred := c.Query("is_starred"); isStarred == "true" {
		// 收藏筛选：只返回当前用户收藏的条目
		query = query.Where("id IN (SELECT entry_id FROM password_entry_stars WHERE username = ?)", usernameStr)
	}
	if keyword := c.Query("keyword"); keyword != "" {
		kw := "%" + keyword + "%"
		query = query.Where("name LIKE ? OR url LIKE ? OR EXISTS (SELECT 1 FROM password_entry_accounts a WHERE a.entry_id = password_entries.id AND (a.username LIKE ? OR a.label LIKE ?))", kw, kw, kw, kw)
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Count(&total)

	var entries []models.PasswordEntry
	if err := query.Order("created_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&entries).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	// 批量填充分类名称、查看用户列表、收藏状态、是否创建人
	if len(entries) > 0 {
		// 收集所有 categoryID 和 entryID
		categoryIDs := make(map[uint]bool)
		entryIDs := make([]uint, 0, len(entries))
		for _, e := range entries {
			categoryIDs[e.CategoryID] = true
			entryIDs = append(entryIDs, e.ID)
		}

		// 批量查询分类
		catMap := make(map[uint]string)
		var cats []models.PasswordCategory
		db.Where("id IN ?", func() []uint {
			ids := make([]uint, 0, len(categoryIDs))
			for id := range categoryIDs {
				ids = append(ids, id)
			}
			return ids
		}()).Find(&cats)
		for _, c := range cats {
			catMap[c.ID] = c.Name
		}

		// 批量查询 viewers
		viewerMap := make(map[uint][]string)
		var viewers []models.PasswordEntryViewer
		db.Where("entry_id IN ?", entryIDs).Find(&viewers)
		for _, v := range viewers {
			viewerMap[v.EntryID] = append(viewerMap[v.EntryID], v.Username)
		}

		// 批量查询收藏状态
		starSet := make(map[uint]bool)
		var stars []models.PasswordEntryStar
		db.Where("entry_id IN ? AND username = ?", entryIDs, usernameStr).Find(&stars)
		for _, s := range stars {
			starSet[s.EntryID] = true
		}

		// 批量查询条目下的账号（密文不返回前端）
		accountMap := make(map[uint][]models.PasswordEntryAccount)
		var accounts []models.PasswordEntryAccount
		db.Where("entry_id IN ?", entryIDs).Order("sort_order ASC, id ASC").Find(&accounts)
		for _, a := range accounts {
			accountMap[a.EntryID] = append(accountMap[a.EntryID], a)
		}

		// 填充数据
		for i := range entries {
			entries[i].CategoryName = catMap[entries[i].CategoryID]
			entries[i].Viewers = viewerMap[entries[i].ID]
			entries[i].Accounts = accountMap[entries[i].ID]
			if entries[i].Accounts == nil {
				entries[i].Accounts = []models.PasswordEntryAccount{}
			}
			entries[i].AccountCount = len(entries[i].Accounts)
			entries[i].IsStarred = starSet[entries[i].ID]
			entries[i].IsCreator = entries[i].CreatedBy == usernameStr
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": entries, "total": total, "page_size": pageSize})
}

// CreatePasswordEntry 创建密码条目
func CreatePasswordEntry(c *gin.Context) {
	var req struct {
		CategoryID        uint     `json:"category_id" binding:"required"`
		Icon              string   `json:"icon" binding:"required"`
		Name              string   `json:"name" binding:"required"`
		Username          string   `json:"username"`          // 存量兼容：无 accounts 时作为单账号
		EncryptedPassword string   `json:"encrypted_password"` // 存量兼容
		URL               string   `json:"url"`
		Port              int      `json:"port"`
		Notes             string   `json:"notes"`
		Viewers           []string `json:"viewers"`
		Accounts          []struct {
			Label             string `json:"label"`
			Username          string `json:"username" binding:"required"`
			EncryptedPassword string `json:"encrypted_password" binding:"required"`
			URL               string `json:"url"`
			Port              int    `json:"port"`
			Notes             string `json:"notes"`
			SortOrder         int    `json:"sort_order"`
		} `json:"accounts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "必填字段缺失"})
		return
	}

	// 存量兼容：未传 accounts 时，顶层账号字段作为单账号处理
	if len(req.Accounts) == 0 {
		if req.Username == "" || req.EncryptedPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请至少添加一个账号"})
			return
		}
		req.Accounts = append(req.Accounts, struct {
			Label             string `json:"label"`
			Username          string `json:"username" binding:"required"`
			EncryptedPassword string `json:"encrypted_password" binding:"required"`
			URL               string `json:"url"`
			Port              int    `json:"port"`
			Notes             string `json:"notes"`
			SortOrder         int    `json:"sort_order"`
		}{Username: req.Username, EncryptedPassword: req.EncryptedPassword, URL: req.URL, Port: req.Port})
	}

	// 逐个账号：RSA 解密后 AES 加密存储
	newAccounts := make([]models.PasswordEntryAccount, 0, len(req.Accounts))
	for i, a := range req.Accounts {
		passwordPlain, err := DecryptPassword(a.EncryptedPassword)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码解密失败"})
			return
		}
		encryptedPwd, err := aesEncrypt(passwordPlain)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
			return
		}
		sortOrder := a.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		newAccounts = append(newAccounts, models.PasswordEntryAccount{
			Label:             a.Label,
			Username:          a.Username,
			EncryptedPassword: encryptedPwd,
			URL:               a.URL,
			Port:              a.Port,
			Notes:             a.Notes,
			SortOrder:         sortOrder,
		})
	}

	username, displayName, approver := services.GetUserContext(c)

	entry := models.PasswordEntry{
		CategoryID: req.CategoryID,
		Icon:       req.Icon,
		Name:       req.Name,
		URL:        req.URL,
		Port:       req.Port,
		Notes:      req.Notes,
		CreatedBy:  username,
		UpdatedBy:  username,
	}

	db := database.GetDB()
	if err := db.Create(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "创建密码条目失败"})
		return
	}

	// 保存账号
	for i := range newAccounts {
		newAccounts[i].EntryID = entry.ID
		if err := db.Create(&newAccounts[i]).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存账号失败"})
			return
		}
	}
	entry.Accounts = newAccounts
	entry.AccountCount = len(newAccounts)

	// 保存查看授权用户（确保创建者在列表中）
	hasCreator := false
	for _, v := range req.Viewers {
		if v == username {
			hasCreator = true
			break
		}
	}
	if !hasCreator {
		req.Viewers = append(req.Viewers, username)
	}
	for _, viewer := range req.Viewers {
		db.Create(&models.PasswordEntryViewer{
			EntryID:  entry.ID,
			Username: viewer,
		})
	}

	accountNames := make([]string, 0, len(newAccounts))
	for _, a := range newAccounts {
		accountNames = append(accountNames, a.Username)
	}
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "条目名称", NewValue: req.Name},
		{FieldName: "CategoryID", FieldLabel: "分类", NewValue: fmt.Sprintf("%d", req.CategoryID)},
		{FieldName: "Accounts", FieldLabel: "账号", NewValue: strings.Join(accountNames, ", ")},
	}
	services.LogOperation(username, displayName, "创建密码条目", "password_entry", entry.ID, entry.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "创建成功", "data": entry})
}

// UpdatePasswordEntry 更新密码条目
func UpdatePasswordEntry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var entry models.PasswordEntry
	if err := database.GetDB().First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "密码条目不存在"})
		return
	}

	oldEntry := entry

	var req struct {
		CategoryID uint     `json:"category_id" binding:"required"`
		Icon       string   `json:"icon" binding:"required"`
		Name       string   `json:"name" binding:"required"`
		URL        string   `json:"url"`
		Port       int      `json:"port"`
		Notes      string   `json:"notes"`
		Viewers    []string `json:"viewers"`
		Accounts   []struct {
			ID                uint   `json:"id"`
			Label             string `json:"label"`
			Username          string `json:"username" binding:"required"`
			EncryptedPassword string `json:"encrypted_password"` // 为空则保留原密码
			URL               string `json:"url"`
			Port              int    `json:"port"`
			Notes             string `json:"notes"`
			SortOrder         int    `json:"sort_order"`
		} `json:"accounts"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "必填字段缺失"})
		return
	}

	db := database.GetDB()

	// 查询现有账号
	var oldAccounts []models.PasswordEntryAccount
	db.Where("entry_id = ?", id).Find(&oldAccounts)
	oldAccountMap := make(map[uint]models.PasswordEntryAccount)
	for _, a := range oldAccounts {
		oldAccountMap[a.ID] = a
	}

	if len(req.Accounts) == 0 && len(oldAccounts) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请至少保留一个账号"})
		return
	}

	// 账号全量替换：有 id 的更新（无新密码则保留原密文），无 id 的新增，请求中缺失的删除
	newAccounts := make([]models.PasswordEntryAccount, 0, len(req.Accounts))
	keptIDs := make(map[uint]bool)
	for i, a := range req.Accounts {
		sortOrder := a.SortOrder
		if sortOrder == 0 {
			sortOrder = i
		}
		if a.ID > 0 {
			old, ok := oldAccountMap[a.ID]
			if !ok {
				continue // 不属于该条目的 id 忽略
			}
			keptIDs[a.ID] = true
			old.Label = a.Label
			old.Username = a.Username
			old.URL = a.URL
			old.Port = a.Port
			old.Notes = a.Notes
			old.SortOrder = sortOrder
			if a.EncryptedPassword != "" {
				passwordPlain, err := DecryptPassword(a.EncryptedPassword)
				if err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码解密失败"})
					return
				}
				encryptedPwd, err := aesEncrypt(passwordPlain)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
					return
				}
				old.EncryptedPassword = encryptedPwd
			}
			if err := db.Save(&old).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新账号失败"})
				return
			}
			newAccounts = append(newAccounts, old)
		} else {
			if a.EncryptedPassword == "" {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "新账号必须填写密码"})
				return
			}
			passwordPlain, err := DecryptPassword(a.EncryptedPassword)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码解密失败"})
				return
			}
			encryptedPwd, err := aesEncrypt(passwordPlain)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码加密失败"})
				return
			}
			acc := models.PasswordEntryAccount{
				EntryID:           uint(id),
				Label:             a.Label,
				Username:          a.Username,
				EncryptedPassword: encryptedPwd,
				URL:               a.URL,
				Port:              a.Port,
				Notes:             a.Notes,
				SortOrder:         sortOrder,
			}
			if err := db.Create(&acc).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "保存账号失败"})
				return
			}
			newAccounts = append(newAccounts, acc)
		}
	}
	// 删除请求中缺失的账号
	for _, old := range oldAccounts {
		if !keptIDs[old.ID] {
			db.Delete(&old)
		}
	}

	entry.CategoryID = req.CategoryID
	entry.Icon = req.Icon
	entry.Name = req.Name
	entry.URL = req.URL
	entry.Port = req.Port
	entry.Notes = req.Notes
	// 存量字段保持不变，避免审计 diff 噪音
	entry.Username = oldEntry.Username
	entry.EncryptedPassword = oldEntry.EncryptedPassword

	username, displayName, approver := services.GetUserContext(c)
	entry.UpdatedBy = username

	if err := db.Save(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "更新密码条目失败"})
		return
	}

	// 更新查看授权用户（仅创建人可修改）
	if entry.CreatedBy == username {
		db.Where("entry_id = ?", id).Delete(&models.PasswordEntryViewer{})
		for _, viewer := range req.Viewers {
			db.Create(&models.PasswordEntryViewer{
				EntryID:  entry.ID,
				Username: viewer,
			})
		}
	}

	entry.Accounts = newAccounts
	entry.AccountCount = len(newAccounts)

	fieldLabels := services.GetFieldLabels("password_entry")
	details := services.DiffStructs(oldEntry, entry, fieldLabels)

	// 账号变更摘要
	oldNames := make([]string, 0, len(oldAccounts))
	for _, a := range oldAccounts {
		oldNames = append(oldNames, a.Username)
	}
	newNames := make([]string, 0, len(newAccounts))
	for _, a := range newAccounts {
		newNames = append(newNames, a.Username)
	}
	if strings.Join(oldNames, ",") != strings.Join(newNames, ",") {
		details = append(details, services.LogDetail{FieldName: "Accounts", FieldLabel: "账号", OldValue: strings.Join(oldNames, ", "), NewValue: strings.Join(newNames, ", ")})
	}
	services.LogOperation(username, displayName, "更新密码条目", "password_entry", entry.ID, entry.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "更新成功", "data": entry})
}

// DeletePasswordEntry 删除密码条目
func DeletePasswordEntry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var entry models.PasswordEntry
	if err := database.GetDB().First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "密码条目不存在"})
		return
	}

	db := database.GetDB()
	// 删除关联的账号、查看授权、收藏和查看日志
	db.Where("entry_id = ?", id).Delete(&models.PasswordEntryAccount{})
	db.Where("entry_id = ?", id).Delete(&models.PasswordEntryViewer{})
	db.Where("entry_id = ?", id).Delete(&models.PasswordViewLog{})
	db.Where("entry_id = ?", id).Delete(&models.PasswordEntryStar{})

	if err := db.Delete(&entry).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "删除密码条目失败"})
		return
	}

	username, displayName, approver := services.GetUserContext(c)
	details := []services.LogDetail{
		{FieldName: "Name", FieldLabel: "条目名称", OldValue: entry.Name},
	}
	services.LogOperation(username, displayName, "删除密码条目", "password_entry", entry.ID, entry.Name, approver, c.ClientIP(), details)

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "删除成功"})
}

// ============ 密码查看 ============

// UnlockPasswordEntry 验证身份并查看密码
func UnlockPasswordEntry(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var entry models.PasswordEntry
	if err := database.GetDB().First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "密码条目不存在"})
		return
	}

	var req struct {
		LDAPPassword string `json:"ldap_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "请输入密码"})
		return
	}

	// RSA 解密 LDAP 密码
	ldapPasswordPlain, err := DecryptPassword(req.LDAPPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "密码解密失败"})
		return
	}

	// 获取当前用户
	currentUsername, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未获取到用户信息"})
		return
	}
	usernameStr := currentUsername.(string)

	// LDAP 验证身份
	_, _, _, _, err = ldapAuthenticate(usernameStr, ldapPasswordPlain)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "身份验证失败"})
		return
	}

	// 检查查看权限
	var viewer models.PasswordEntryViewer
	if err := database.GetDB().Where("entry_id = ? AND username = ?", id, usernameStr).First(&viewer).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "您无权查看此密码"})
		return
	}

	// 查询条目下所有账号并逐个 AES 解密
	db := database.GetDB()
	var accounts []models.PasswordEntryAccount
	db.Where("entry_id = ?", id).Order("sort_order ASC, id ASC").Find(&accounts)

	result := make([]gin.H, 0, len(accounts))
	for _, a := range accounts {
		passwordPlain, err := aesDecrypt(a.EncryptedPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "密码解密失败"})
			return
		}
		result = append(result, gin.H{
			"id":       a.ID,
			"label":    a.Label,
			"username": a.Username,
			"password": passwordPlain,
		})
	}

	// 记录查看日志（按条目一条）
	db.Create(&models.PasswordViewLog{
		EntryID:   entry.ID,
		EntryName: entry.Name,
		Viewer:    usernameStr,
		ViewedAt:  time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": gin.H{"accounts": result}})
}

// ============ 收藏管理 ============

// TogglePasswordEntryStar 切换收藏状态（per-user）
func TogglePasswordEntryStar(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var entry models.PasswordEntry
	if err := database.GetDB().First(&entry, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "密码条目不存在"})
		return
	}

	// 获取当前用户
	currentUsername, ok := c.Get("username")
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"code": 401, "message": "未获取到用户信息"})
		return
	}
	usernameStr := currentUsername.(string)

	var req struct {
		IsStarred bool `json:"is_starred"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "参数无效"})
		return
	}

	db := database.GetDB()
	if req.IsStarred {
		// 收藏：插入记录（忽略重复）
		db.FirstOrCreate(&models.PasswordEntryStar{}, map[string]interface{}{
			"entry_id":  uint(id),
			"username": usernameStr,
		})
	} else {
		// 取消收藏：删除记录
		db.Where("entry_id = ? AND username = ?", id, usernameStr).Delete(&models.PasswordEntryStar{})
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "message": "操作成功"})
}

// ============ 审计日志 ============

// ListPasswordViewLogs 获取密码查看审计日志
func ListPasswordViewLogs(c *gin.Context) {
	db := database.GetDB()
	query := db.Model(&models.PasswordViewLog{})

	if entryID := c.Query("entry_id"); entryID != "" {
		query = query.Where("entry_id = ?", entryID)
	}
	if viewer := c.Query("viewer"); viewer != "" {
		query = query.Where("viewer = ?", viewer)
	}
	if startDate := c.Query("start_date"); startDate != "" {
		query = query.Where("viewed_at >= ?", startDate)
	}
	if endDate := c.Query("end_date"); endDate != "" {
		query = query.Where("viewed_at <= ?", endDate+" 23:59:59")
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	var total int64
	query.Count(&total)

	var logs []models.PasswordViewLog
	if err := query.Order("viewed_at DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "查询失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": logs, "total": total, "page_size": pageSize})
}
