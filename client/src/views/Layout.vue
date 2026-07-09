<template>
  <el-container style="height: 100vh">
    <el-aside width="210px" class="sidebar-aside">
      <div class="sidebar-logo">
        <img src="/favicon.ico" alt="logo" class="sidebar-logo-icon" />
        <span>管理平台</span>
      </div>
      <el-menu
        :default-active="activeMenu"
        background-color="#0F172A"
        text-color="#94A3B8"
        active-text-color="#F8FAFC"
        router
      >
        <el-menu-item index="/dashboard">
          <svg-icon name="bar-chart-2" />
          <span slot="title">数据看板</span>
        </el-menu-item>
        <el-menu-item index="/policies">
          <svg-icon name="file-text" />
          <span slot="title">IT政策</span>
        </el-menu-item>
        <el-submenu index="asset">
          <template slot="title">
            <svg-icon name="monitor" />
            <span>资产管理</span>
          </template>
          <el-menu-item index="/assets">
            <svg-icon name="list" />
            <span>资产列表</span>
          </el-menu-item>
          <el-menu-item index="/regions">
            <svg-icon name="map-pin" />
            <span>区域管理</span>
          </el-menu-item>
          <el-menu-item index="/os-types">
            <svg-icon name="layers" />
            <span>操作系统管理</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="network-security">
          <template slot="title">
            <svg-icon name="shield" />
            <span>网络安全</span>
          </template>
          <el-menu-item index="/topology">
            <svg-icon name="network" />
            <span>网络拓扑图</span>
          </el-menu-item>
          <el-menu-item index="/change-management">
            <svg-icon name="git-branch" />
            <span>变更管理</span>
          </el-menu-item>
          <el-menu-item index="/vulnerability-scan">
            <svg-icon name="search" />
            <span>漏洞扫描</span>
          </el-menu-item>
          <el-menu-item index="/penetration-test">
            <svg-icon name="crosshair" />
            <span>渗透测试</span>
          </el-menu-item>
          <el-menu-item index="/firewall-check">
            <svg-icon name="shield-check" />
            <span>防火墙检查</span>
          </el-menu-item>
          <el-menu-item index="/security-rectification">
            <svg-icon name="pencil" />
            <span>安全整改记录</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="system-security">
          <template slot="title">
            <svg-icon name="settings" />
            <span>系统安全</span>
          </template>
          <el-menu-item index="/system-hardening">
            <svg-icon name="hammer" />
            <span>系统加固</span>
          </el-menu-item>
          <el-menu-item index="/patch-update">
            <svg-icon name="download" />
            <span>补丁更新</span>
          </el-menu-item>
          <el-menu-item index="/virus-control">
            <svg-icon name="shield-alert" />
            <span>病毒控制</span>
          </el-menu-item>
          <el-menu-item index="/backup-management">
            <svg-icon name="database" />
            <span>备份管理</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="permission">
          <template slot="title">
            <svg-icon name="users" />
            <span>用户管理</span>
          </template>
          <el-menu-item index="/permissions">
            <svg-icon name="key" />
            <span>岗位权限设置</span>
          </el-menu-item>
          <el-menu-item index="/user-permissions">
            <svg-icon name="user-check" />
            <span>用户权限一览</span>
          </el-menu-item>
          <el-menu-item index="/sftp-accounts">
            <svg-icon name="terminal" />
            <span>SFTP账号一览</span>
          </el-menu-item>
          <el-menu-item index="/monthly-check-history">
            <svg-icon name="calendar-check" />
            <span>月度检查历史</span>
          </el-menu-item>
          <el-menu-item index="/user-change-history">
            <svg-icon name="user-cog" />
            <span>用户变更记录</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="third-party">
          <template slot="title">
            <svg-icon name="layout-grid" />
            <span>第三方应用</span>
          </template>
          <el-menu-item index="/approved-software">
            <svg-icon name="check-circle" />
            <span>核准软件目录</span>
          </el-menu-item>
          <el-menu-item index="/asset-software">
            <svg-icon name="table" />
            <span>资产对应表</span>
          </el-menu-item>
          <el-menu-item index="/quarterly-check-history">
            <svg-icon name="calendar" />
            <span>季度检查历史</span>
          </el-menu-item>
        </el-submenu>
        <el-submenu index="log">
          <template slot="title">
            <svg-icon name="scroll-text" />
            <span>日志管理</span>
          </template>
          <el-menu-item index="/login-logs">
            <svg-icon name="log-in" />
            <span>登录日志</span>
          </el-menu-item>
          <el-menu-item index="/operation-logs">
            <svg-icon name="file-search" />
            <span>操作日志</span>
          </el-menu-item>
        </el-submenu>
        <el-menu-item index="/form-publish">
          <svg-icon name="upload-cloud" />
          <span>表单发布</span>
        </el-menu-item>
      </el-menu>
    </el-aside>
    <el-container>
      <el-header class="app-header">
        <div class="header-title-wrap">
          <div class="header-title">
            <span class="header-prefix">~/</span>
            <span ref="titleTyping" class="header-title-text"></span>
            <span ref="titleCursor" class="header-cursor"></span>
          </div>
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
import SvgIcon from '@/components/SvgIcon.vue'

export default {
  name: 'Layout',
  components: { SvgIcon },
  data() {
    return {
      displayName: localStorage.getItem('display_name') || '用户',
      subtitleAnimation: null,
      titleTypingTimer: null
    }
  },
  computed: {
    activeMenu() {
      return this.$route.path
    }
  },
  watch: {
    '$route.path'() {
      this.$nextTick(() => {
        this.startSubtitleAnimation()
        this.startTitleTypingAnimation()
      })
    }
  },
  mounted() {
    this.$nextTick(() => {
      this.startSubtitleAnimation()
      this.startTitleTypingAnimation()
    })
  },
  beforeDestroy() {
    if (this.subtitleAnimation) {
      this.subtitleAnimation.pause()
    }
    if (this.titleTypingTimer) {
      clearTimeout(this.titleTypingTimer)
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
    },
    startTitleTypingAnimation() {
      const el = this.$refs.titleTyping
      if (!el) return
      // 清除之前的定时器
      if (this.titleTypingTimer) {
        clearTimeout(this.titleTypingTimer)
        this.titleTypingTimer = null
      }
      const title = this.$route.meta.title || ''
      el.textContent = ''
      this._typeTitle(el, title, 0)
    },
    _typeTitle(el, title, index) {
      // 逐字输入阶段
      if (index <= title.length) {
        el.textContent = title.substring(0, index)
        this.titleTypingTimer = setTimeout(() => {
          this._typeTitle(el, title, index + 1)
        }, 100)
        return
      }
      // 输入完成，停留 2.5 秒后开始删除
      this.titleTypingTimer = setTimeout(() => {
        this._deleteTitle(el, title, title.length)
      }, 2500)
    },
    _deleteTitle(el, title, index) {
      // 逐字删除阶段
      if (index > 0) {
        el.textContent = title.substring(0, index - 1)
        this.titleTypingTimer = setTimeout(() => {
          this._deleteTitle(el, title, index - 1)
        }, 60)
        return
      }
      // 删除完成，停留 500ms 后重新开始
      this.titleTypingTimer = setTimeout(() => {
        this._typeTitle(el, title, 0)
      }, 500)
    }
  }
}
</script>

<style scoped>
.el-main {
  padding: 0;
  background: transparent;
}
</style>

<style>
/* 非看板页面恢复默认内边距 */
.el-main > *:not(.dashboard) {
  padding: 20px;
}
</style>
