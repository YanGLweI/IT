package handlers

import (
	"fmt"
	"net/http"

	"it-platform-server/config"

	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
)

// LDAPUser LDAP用户信息
type LDAPUser struct {
	DN             string `json:"dn"`
	DisplayName    string `json:"display_name"`
	SAMAccountName string `json:"sAMAccountName"`
}

// GetLDAPUsers 获取LDAP IT部成员列表
func GetLDAPUsers(c *gin.Context) {
	users, err := getLDAPUsersFromGroup()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "获取LDAP用户列表失败: " + err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": users,
	})
}

// getLDAPUsersFromGroup 从安全组获取所有成员
func getLDAPUsersFromGroup() ([]LDAPUser, error) {
	cfg := &config.Cfg.LDAP

	// 创建LDAP连接
	l, err := ldap.DialURL(cfg.Server, ldap.DialWithTLSConfig(getTLSConfig(cfg)))
	if err != nil {
		return nil, fmt.Errorf("连接 LDAP 失败: %v", err)
	}
	defer l.Close()

	// 使用服务账号绑定
	err = l.Bind(cfg.Username, cfg.Password)
	if err != nil {
		return nil, fmt.Errorf("LDAP 绑定失败: %v", err)
	}

	// 搜索安全组成员
	searchRequest := ldap.NewSearchRequest(
		cfg.SecurityGroupDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=group)",
		[]string{"member"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, fmt.Errorf("查询安全组失败: %v", err)
	}

	if len(sr.Entries) == 0 {
		return nil, nil
	}

	var users []LDAPUser
	// 获取每个成员的详细信息
	for _, entry := range sr.Entries {
		memberDNs := entry.GetAttributeValues("member")
		for _, memberDN := range memberDNs {
			userInfo, err := getUserInfoByDN(memberDN, l)
			if err != nil {
				continue
			}
			users = append(users, *userInfo)
		}
	}

	return users, nil
}

// getUserInfoByDN 根据DN获取用户详细信息
func getUserInfoByDN(userDN string, l *ldap.Conn) (*LDAPUser, error) {
	searchRequest := ldap.NewSearchRequest(
		userDN,
		ldap.ScopeBaseObject, ldap.NeverDerefAliases, 0, 0, false,
		"(objectClass=*)",
		[]string{"displayName", "sAMAccountName", "cn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil || len(sr.Entries) == 0 {
		return nil, err
	}

	entry := sr.Entries[0]
	displayName := entry.GetAttributeValue("displayName")
	if displayName == "" {
		displayName = entry.GetAttributeValue("cn")
	}
	sAMAccountName := entry.GetAttributeValue("sAMAccountName")

	return &LDAPUser{
		DN:             userDN,
		DisplayName:    displayName,
		SAMAccountName: sAMAccountName,
	}, nil
}
