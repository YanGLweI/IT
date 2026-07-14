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
		&models.PasswordEntryViewer{},
		&models.PasswordViewLog{},
	)
	if err != nil {
		log.Fatalf("数据库迁移失败: %v", err)
	}

	// 初始化基础数据
	SeedPermissionRules()
	SeedChangeTypes()
	SeedPasswordCategories()

	fmt.Println("数据库初始化成功!")
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
