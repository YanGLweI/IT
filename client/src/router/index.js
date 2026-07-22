import Vue from 'vue'
import VueRouter from 'vue-router'
import Layout from '../views/Layout.vue'
import PublicLayout from '../views/public/PublicLayout.vue'

Vue.use(VueRouter)

const routes = [
  {
    path: '/login',
    name: 'Login',
    component: () => import('../views/login/Login.vue'),
    meta: { title: '登录', public: true }
  },

  // 免登录模块
  {
    path: '/public',
    component: PublicLayout,
    redirect: '/public/forms',
    children: [
      {
        path: 'forms',
        name: 'PublicForms',
        component: () => import('../views/public/FormDownload.vue'),
        meta: { title: '表单下载', public: true }
      },
      {
        path: 'it-guides',
        name: 'PublicITGuides',
        component: () => import('../views/public/ITGuideList.vue'),
        meta: { title: 'IT指南', public: true }
      },
      {
        path: 'it-guides/:id',
        name: 'PublicITGuideDetail',
        component: () => import('../views/public/ITGuideDetail.vue'),
        meta: { title: '指南详情', public: true }
      }
    ]
  },

  // 管理端
  {
    path: '/',
    component: Layout,
    redirect: '/dashboard',
    children: [
      { path: 'dashboard', name: 'Dashboard', component: () => import('../views/dashboard/Dashboard.vue'), meta: { title: '数据看板', enTitle: 'Data Dashboard' } },
      { path: 'regions', name: 'Regions', component: () => import('../views/region/RegionList.vue'), meta: { title: '区域管理', enTitle: 'Region Management' } },
      { path: 'os-types', name: 'OSTypes', component: () => import('../views/ostype/OSTypeList.vue'), meta: { title: '操作系统管理', enTitle: 'OS Type Management' } },
      { path: 'assets', name: 'Assets', component: () => import('../views/asset/AssetList.vue'), meta: { title: '资产管理', enTitle: 'Asset Management' } },
      { path: 'policies', name: 'Policies', component: () => import('../views/policy/PolicyList.vue'), meta: { title: 'IT政策', enTitle: 'IT Policy' } },
      { path: 'topology', name: 'Topology', component: () => import('../views/topology/TopologyView.vue'), meta: { title: '网络拓扑图', enTitle: 'Network Topology' } },
      { path: 'change-management', name: 'ChangeManagement', component: () => import('../views/network-security/ChangeManagement.vue'), meta: { title: '变更管理', enTitle: 'Change Management' } },
      { path: 'vulnerability-scan', name: 'VulnerabilityScan', component: () => import('../views/network-security/VulnerabilityScan.vue'), meta: { title: '漏洞扫描', enTitle: 'Vulnerability Scan' } },
      { path: 'penetration-test', name: 'PenetrationTest', component: () => import('../views/network-security/PenetrationTest.vue'), meta: { title: '渗透测试', enTitle: 'Penetration Test' } },
      { path: 'firewall-check', name: 'FirewallCheck', component: () => import('../views/network-security/FirewallCheck.vue'), meta: { title: '防火墙检查', enTitle: 'Firewall Check' } },
      { path: 'security-rectification', name: 'SecurityRectification', component: () => import('../views/network-security/SecurityRectification.vue'), meta: { title: '安全整改记录', enTitle: 'Security Rectification' } },
      { path: 'system-hardening', name: 'SystemHardening', component: () => import('../views/system-security/SystemHardening.vue'), meta: { title: '系统加固', enTitle: 'System Hardening' } },
      { path: 'patch-update', name: 'PatchUpdate', component: () => import('../views/system-security/PatchUpdate.vue'), meta: { title: '补丁更新', enTitle: 'Patch Update' } },
      { path: 'virus-control', name: 'VirusControl', component: () => import('../views/system-security/VirusControl.vue'), meta: { title: '病毒控制', enTitle: 'Virus Control' } },
      { path: 'backup-management', name: 'BackupManagement', component: () => import('../views/system-security/BackupManagement.vue'), meta: { title: '备份管理', enTitle: 'Backup Management' } },
      { path: 'permissions', name: 'Permissions', component: () => import('../views/permission/PermissionList.vue'), meta: { title: '岗位权限设置', enTitle: 'Permission Settings' } },
      { path: 'user-permissions', name: 'UserPermissions', component: () => import('../views/user-permission/UserPermissionList.vue'), meta: { title: '用户权限一览', enTitle: 'User Permissions' } },
      { path: 'sftp-accounts', name: 'SftpAccounts', component: () => import('../views/sftp/SftpAccountList.vue'), meta: { title: 'SFTP账号一览', enTitle: 'SFTP Accounts' } },
      { path: 'approved-software', name: 'ApprovedSoftware', component: () => import('../views/approved-software/ApprovedSoftwareList.vue'), meta: { title: '核准软件目录', enTitle: 'Approved Software' } },
      { path: 'asset-software', name: 'AssetSoftware', component: () => import('../views/approved-software/AssetSoftwareList.vue'), meta: { title: '资产对应表', enTitle: 'Asset Software Map' } },
      { path: 'login-logs', name: 'LoginLogs', component: () => import('../views/log/LoginLogList.vue'), meta: { title: '登录日志', enTitle: 'Login Logs' } },
      { path: 'operation-logs', name: 'OperationLogs', component: () => import('../views/log/OperationLogList.vue'), meta: { title: '操作日志', enTitle: 'Operation Logs' } },
      { path: 'monthly-check-history', name: 'MonthlyCheckHistory', component: () => import('../views/permission/MonthlyCheckHistory.vue'), meta: { title: '月度检查历史', enTitle: 'Monthly Check History' } },
      { path: 'quarterly-check-history', name: 'QuarterlyCheckHistory', component: () => import('../views/approved-software/QuarterlyCheckHistory.vue'), meta: { title: '季度检查历史', enTitle: 'Quarterly Check History' } },
      { path: 'user-change-history', name: 'UserChangeHistory', component: () => import('../views/permission/UserChangeHistory.vue'), meta: { title: '用户变更记录', enTitle: 'User Change History' } },
      { path: 'form-publish', name: 'FormPublish', component: () => import('../views/form-publish/FormVault.vue'), meta: { title: '表单发布', enTitle: 'Form Publishing' } },
      { path: 'it-guide', name: 'ITGuide', component: () => import('../views/it-guide/ITGuideList.vue'), meta: { title: 'IT指南', enTitle: 'IT Guide' } },
      { path: 'calendar', name: 'Calendar', component: () => import('../views/calendar/CalendarView.vue'), meta: { title: '日程管理', enTitle: 'Calendar Management' } },
      { path: 'password-vault', name: 'PasswordVault', component: () => import('../views/password-vault/index.vue'), meta: { title: '密码本', enTitle: 'Password Vault' } },
      { path: 'dedicated-lines', name: 'DedicatedLines', component: () => import('../views/dedicated-line/DedicatedLineList.vue'), meta: { title: '专线信息', enTitle: 'Dedicated Lines' } },
      { path: 'ipsec-vpn', name: 'IPsecVPN', component: () => import('../views/ipsec-vpn/IPsecVpnList.vue'), meta: { title: 'IPsec VPN', enTitle: 'IPsec VPN' } }
    ]
  },

  // 404 通配符路由（必须放在最后）
  {
    path: '*',
    name: 'NotFound',
    component: () => import('../views/error/NotFound.vue'),
    meta: { title: '页面未找到', public: true }
  }
]

const router = new VueRouter({
  mode: 'history',
  routes
})

// 路由守卫：检查登录状态
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = to.meta.title ? `${to.meta.title} - IT管理平台` : 'IT管理平台'

  // 检查匹配链中是否有 public 标记（支持嵌套路由）
  if (to.matched.some(record => record.meta.public)) {
    next()
    return
  }

  // 检查是否已登录
  const token = localStorage.getItem('token')
  if (!token) {
    next('/login')
    return
  }

  next()
})

export default router
