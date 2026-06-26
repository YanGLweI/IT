<template>
  <el-container style="height: 100vh">
    <el-aside width="200px" style="background-color: #304156">
      <div style="padding: 20px; text-align: center; color: #fff; font-size: 18px; font-weight: bold">
        IT管理平台
      </div>
      <el-menu
        :default-active="activeMenu"
        background-color="#304156"
        text-color="#bfcbd9"
        active-text-color="#409EFF"
        router
      >
        <el-menu-item index="/dashboard">
          <i class="el-icon-data-board"></i>
          <span slot="title">看板</span>
        </el-menu-item>
        <el-submenu index="asset">
          <template slot="title">
            <i class="el-icon-monitor"></i>
            <span>资产管理</span>
          </template>
          <el-menu-item index="/assets">
            <i class="el-icon-s-order"></i>
            <span>资产列表</span>
          </el-menu-item>
          <el-menu-item index="/regions">
            <i class="el-icon-place"></i>
            <span>区域管理</span>
          </el-menu-item>
          <el-menu-item index="/os-types">
            <i class="el-icon-s-platform"></i>
            <span>操作系统管理</span>
          </el-menu-item>
        </el-submenu>
        <el-menu-item index="/policies">
          <i class="el-icon-document"></i>
          <span slot="title">IT政策</span>
        </el-menu-item>
        <el-menu-item index="/topology">
          <i class="el-icon-share"></i>
          <span slot="title">网络拓扑图</span>
        </el-menu-item>
        <el-submenu index="permission">
          <template slot="title">
            <i class="el-icon-s-check"></i>
            <span>岗位权限</span>
          </template>
          <el-menu-item index="/permissions">
            <i class="el-icon-setting"></i>
            <span>岗位权限设置</span>
          </el-menu-item>
          <el-menu-item index="/user-permissions">
            <i class="el-icon-user"></i>
            <span>用户权限一览</span>
          </el-menu-item>
          <el-menu-item index="/sftp-accounts">
            <i class="el-icon-connection"></i>
            <span>SFTP账号一览</span>
          </el-menu-item>
          <el-menu-item index="/monthly-check-history">
            <i class="el-icon-document-checked"></i>
            <span>月度检查历史</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="third-party">
          <template slot="title">
            <i class="el-icon-apple"></i>
            <span>第三方应用</span>
          </template>
          <el-menu-item index="/approved-software">
            <i class="el-icon-document-checked"></i>
            <span>核准软件目录</span>
          </el-menu-item>
          <el-menu-item index="/asset-software">
            <i class="el-icon-s-grid"></i>
            <span>资产对应表</span>
          </el-menu-item>
          <el-menu-item index="/quarterly-check-history">
            <i class="el-icon-document-checked"></i>
            <span>季度检查历史</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="log">
          <template slot="title">
            <i class="el-icon-notebook-2"></i>
            <span>日志管理</span>
          </template>
          <el-menu-item index="/login-logs">
            <i class="el-icon-s-check"></i>
            <span>登录日志</span>
          </el-menu-item>
          <el-menu-item index="/operation-logs">
            <i class="el-icon-document"></i>
            <span>操作日志</span>
          </el-menu-item>
        </el-submenu>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header style="background: #fff; border-bottom: 1px solid #e6e6e6; display: flex; align-items: center; justify-content: space-between; padding: 0 20px">
        <h3 style="margin: 0; color: #333">{{ $route.meta.title }}</h3>
        <div class="header-right">
          <el-dropdown @command="handleCommand">
            <span class="user-info">
              <i class="el-icon-user-solid"></i>
              {{ displayName }}
              <i class="el-icon-arrow-down el-icon--right"></i>
            </span>
            <el-dropdown-menu slot="dropdown">
              <el-dropdown-item command="logout">
                <i class="el-icon-switch-button"></i> 退出登录
              </el-dropdown-item>
            </el-dropdown-menu>
          </el-dropdown>
        </div>
      </el-header>
      <el-main style="background: #f0f2f5; padding: 20px">
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import { logout } from '@/api/audit_log'

export default {
  name: 'Layout',
  data() {
    return {
      displayName: localStorage.getItem('display_name') || '用户'
    }
  },
  computed: {
    activeMenu() {
      return this.$route.path
    }
  },
  methods: {
    handleCommand(command) {
      if (command === 'logout') {
        this.$confirm('确定要退出登录吗？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        }).then(async () => {
          try {
            await logout()
          } catch (error) {
            // 忽略错误（包括401），继续执行退出流程
          }
          
          // 请求完成后才清除本地存储
          localStorage.removeItem('token')
          localStorage.removeItem('username')
          localStorage.removeItem('display_name')
          
          // 如果当前不在登录页才跳转
          if (this.$route.path !== '/login') {
            this.$router.replace('/login').catch(() => {})
          }
        }).catch(() => {})
      }
    }
  }
}
</script>

<style scoped>
.header-right {
  display: flex;
  align-items: center;
}
.user-info {
  cursor: pointer;
  display: flex;
  align-items: center;
  font-size: 14px;
  color: #606266;
}
.user-info i:first-child {
  margin-right: 5px;
  font-size: 16px;
  color: #409EFF;
}
</style>
