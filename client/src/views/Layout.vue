<template>
  <el-container style="height: 100vh">
    <el-aside width="210px" class="sidebar-aside">
      <div class="sidebar-logo">
        <img src="/favicon.ico" alt="logo" class="sidebar-logo-icon" />
        <span>管理平台</span>
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
        <el-menu-item index="/policies">
          <i class="el-icon-document"></i>
          <span slot="title">IT政策</span>
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
        <el-submenu index="network-security">
          <template slot="title">
            <i class="el-icon-connection"></i>
            <span>网络安全</span>
          </template>
          <el-menu-item index="/topology">
            <i class="el-icon-share"></i>
            <span>网络拓扑图</span>
          </el-menu-item>
          <el-menu-item index="/change-management">
            <i class="el-icon-document-copy"></i>
            <span>变更管理</span>
          </el-menu-item>
          <el-menu-item index="/vulnerability-scan">
            <i class="el-icon-search"></i>
            <span>漏洞扫描</span>
          </el-menu-item>
          <el-menu-item index="/penetration-test">
            <i class="el-icon-cpu"></i>
            <span>渗透测试</span>
          </el-menu-item>
          <el-menu-item index="/firewall-check">
            <i class="el-icon-s-check"></i>
            <span>防火墙检查</span>
          </el-menu-item>
          <el-menu-item index="/security-rectification">
            <i class="el-icon-edit"></i>
            <span>安全整改记录</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="system-security">
          <template slot="title">
            <i class="el-icon-setting"></i>
            <span>系统安全</span>
          </template>
          <el-menu-item index="/system-hardening">
            <i class="el-icon-s-tools"></i>
            <span>系统加固</span>
          </el-menu-item>
          <el-menu-item index="/patch-update">
            <i class="el-icon-refresh"></i>
            <span>补丁更新</span>
          </el-menu-item>
          <el-menu-item index="/virus-control">
            <i class="el-icon-warning"></i>
            <span>病毒控制</span>
          </el-menu-item>
          <el-menu-item index="/backup-management">
            <i class="el-icon-files"></i>
            <span>备份管理</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="permission">
          <template slot="title">
            <i class="el-icon-user"></i>
            <span>用户管理</span>
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
          <el-menu-item index="/user-change-history">
            <i class="el-icon-document-copy"></i>
            <span>用户变更记录</span>
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
      <el-header class="app-header">
        <div class="header-title-wrap">
          <h3 class="header-title">{{ $route.meta.title }}</h3>
          <p ref="subtitleText" :key="$route.path" class="header-subtitle">{{ $route.meta.enTitle }}</p>
        </div>
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
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<script>
import { logout } from '@/api/audit_log'
import { animate, scrambleText } from 'animejs'

export default {
  name: 'Layout',
  data() {
    return {
      displayName: localStorage.getItem('display_name') || '用户',
      subtitleAnimation: null
    }
  },
  computed: {
    activeMenu() {
      return this.$route.path
    },
    currentEnTitle() {
      return this.$route.meta.enTitle || ''
    }
  },
  watch: {
    '$route.path'() {
      // key 改变会导致元素重建，等待下一个 tick 后再启动动画
      this.$nextTick(() => {
        this.startSubtitleAnimation()
      })
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.startSubtitleAnimation()
    })
  },
  beforeDestroy() {
    if (this.subtitleAnimation) {
      this.subtitleAnimation.pause()
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
    },
    startSubtitleAnimation() {
      const el = this.$refs.subtitleText
      if (!el) return
      // 停止之前的动画
      if (this.subtitleAnimation) {
        this.subtitleAnimation.pause()
        this.subtitleAnimation = null
      }
      // 启动循环 scramble 动画
      this.subtitleAnimation = animate(el, {
        innerHTML: scrambleText(),
        loop: true,
        loopDelay: 2000
      })
    }
  }
}
</script>

<style scoped>
.el-main {
  padding: 0;
  background: transparent;
}
.sidebar-aside {
  background-color: #304156;
  overflow-y: auto;
  scrollbar-width: none; /* Firefox */
  -ms-overflow-style: none; /* IE/Edge */
}
::v-deep .sidebar-aside .el-menu {
  border-right: none;
  padding: 0;
  margin: 0;
}
::v-deep .sidebar-aside::-webkit-scrollbar {
  display: none; /* Chrome/Safari */
}
.sidebar-logo {
  padding: 20px;
  text-align: center;
  color: #fff;
  font-size: 18px;
  font-weight: bold;
  display: flex;
  align-items: center;
  justify-content: center;
  /* gap: 8px; */
}
.sidebar-logo-icon {
  width: 24px;
  height: 24px;
  vertical-align: middle;
}

</style>

<style>
/* 非看板页面恢复默认内边距 */
.el-main > *:not(.dashboard) {
  padding: 20px;
}
</style>
