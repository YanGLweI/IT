package database

import (
	"fmt"
	"log"

	"it-platform-server/config"
	"it-platform-server/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// InitDB 初始化数据库连接
func InitDB() {
	dsn := config.GetDSN()
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("数据库连接失败: %v", err)
	}

	// 自动创建表
	err = DB.AutoMigrate(
		&models.Region{},
		&models.Asset{},
		&models.Policy{},
		&models.Topology{},
		&models.OSType{},
		&models.PermissionRule{},
		&models.Department{},
		&models.DepartmentPosition{},
		&models.UserPermission{},
		&models.SftpServer{},
		&models.SftpAccount{},
		&models.ApprovedSoftware{},
		&models.AssetSoftware{},
		&models.LoginLog{},
		&models.OperationLog{},
		&models.OperationLogDetail{},
		&models.MonthlyCheckHistory{},
		&models.QuarterlyCheckHistory{},
		&models.QuarterlyCheckSoftware{},
		&models.UserChangeHistory{},
		&models.ChangeRecordTemplate{},
		&models.ChangeType{},
		&models.ChangeRecord{},
		&models.VulnerabilityScan{},
		&models.SystemHardeningHistory{},
		&models.PenetrationTest{},
		&models.FirewallCheck{},
		&models.PatchUpdate{},
		&models.BackupRecord{},
		&models.BackupRecovery{},
		&models.BackupTemplate{},
		&models.FormVaultItem{},
		&models.Calendar{},
		&models.CalendarParticipant{},
		&models.CalendarNotification{},
		&models.PasswordCategory{},
		&models.PasswordEntry{},
		&models.PasswordEntryAccount{},
		&models.PasswordEntryViewer{},
		&models.PasswordEntryStar{},
		&models.PasswordViewLog{},
		&models.ITGuide{},
		&models.ITGuideStep{},
		&models.ITGuideMedia{},
		&models.ITGuideAttachment{},
		&models.DedicatedLine{},
		&models.IPsecVpn{},
		&models.MenuFavorite{},
		&models.FirewallNode{},
		&models.RegionFirewall{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化区域排序：为历史数据按 id 顺序填充 sort_order（排序后的记录均非 0，不会误伤）
	DB.Model(&models.Region{}).Where("sort_order = 0").Update("sort_order", gorm.Expr("id"))

	// 初始化基础数据
	SeedPermissionRules()
	SeedChangeTypes()
	SeedPasswordCategories()

	// 密码本存量迁移：旧条目（单账号）转换为账号子表记录（幂等）
	var legacyEntries []models.PasswordEntry
	DB.Where("username != '' AND encrypted_password != '' AND id NOT IN (SELECT entry_id FROM password_entry_accounts)").
		Find(&legacyEntries)
	for _, e := range legacyEntries {
		DB.Create(&models.PasswordEntryAccount{
			EntryID:           e.ID,
			Username:          e.Username,
			EncryptedPassword: e.EncryptedPassword,
			URL:               e.URL,
			Port:              e.Port,
		})
	}

	fmt.Println("数据库初始化成功!")
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
