package models

import "time"

// IPsecVpn IPsec VPN配置信息
type IPsecVpn struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 网络
	TunnelName   string `gorm:"type:varchar(200);not null" json:"tunnel_name"`  // 隧道名
	Owner        string `gorm:"type:varchar(100);not null" json:"owner"`        // 负责人
	RemoteIP     string `gorm:"type:varchar(50);not null" json:"remote_ip"`     // 对端IP
	LocalIP      string `gorm:"type:varchar(50);not null" json:"local_ip"`      // 本端IP
	NetworkImage string `gorm:"type:varchar(500)" json:"network_image"`         // 网络配置截图路径

	// 认证
	PSK        string `gorm:"type:varchar(500);not null" json:"psk"`       // 预共享密钥
	IKEVersion int    `gorm:"type:tinyint;not null" json:"ike_version"`    // IKE版本: 1 或 2
	Mode       string `gorm:"type:varchar(20)" json:"mode"`                // 模式: 野蛮模式/主模式（IKE=1时必填）

	// 阶段一
	Phase1Image string `gorm:"type:varchar(500)" json:"phase1_image"` // 阶段一截图路径

	// 阶段二（JSON数组: [{local_addr, remote_addr, image}]）
	Phase2Entries string `gorm:"type:text" json:"phase2_entries"`

	CreatedBy string `gorm:"type:varchar(100)" json:"created_by"`
	UpdatedBy string `gorm:"type:varchar(100)" json:"updated_by"`
}

func (IPsecVpn) TableName() string {
	return "ipsec_vpns"
}
