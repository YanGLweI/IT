package models

import "time"

// DedicatedLine 专线信息
type DedicatedLine struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Factory       string `gorm:"type:varchar(100);not null" json:"factory"`        // 厂区
	Carrier       string `gorm:"type:varchar(50);not null" json:"carrier"`         // 运营商: 电信/联通/移动/广电
	BandwidthUp   int    `gorm:"type:int;default:0" json:"bandwidth_up"`           // 上行带宽(Mbps)
	BandwidthDown int    `gorm:"type:int;default:0" json:"bandwidth_down"`         // 下行带宽(Mbps)
	IPStart       string `gorm:"type:varchar(50);not null" json:"ip_start"`        // IP范围起始
	IPEnd         string `gorm:"type:varchar(50);not null" json:"ip_end"`          // IP范围结束
	SubnetMask    string `gorm:"type:varchar(50);not null" json:"subnet_mask"`     // 子网掩码
	Gateway       string `gorm:"type:varchar(50);not null" json:"gateway"`         // 网关
	DNS           string `gorm:"type:varchar(200)" json:"dns"`                     // DNS
	IPCount       int    `gorm:"type:int;default:0" json:"ip_count"`               // IP数(自动计算)
	Images        string `gorm:"type:text" json:"images"`                          // 图片路径JSON数组
	Notes         string `gorm:"type:text" json:"notes"`                           // 备注
	CreatedBy     string `gorm:"type:varchar(100)" json:"created_by"`              // 创建人
	UpdatedBy     string `gorm:"type:varchar(100)" json:"updated_by"`              // 更新人
}

func (DedicatedLine) TableName() string {
	return "dedicated_lines"
}
