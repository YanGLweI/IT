<template>
  <el-container style="height: 100vh">
    <el-aside width="210px" class="sidebar-aside">
      <div class="sidebar-logo">
        <img src="/favicon.ico" alt="logo" class="sidebar-logo-icon" />
        <span>管理平台</span>
      </div>
      <el-menu
        ref="sidebarMenu"
        :default-active="activeMenu"
        background-color="#0F172A"
        text-color="#94A3B8"
        active-text-color="#F8FAFC"
        unique-opened
        @select="handleMenuSelect"
      >
        <!-- 我的收藏：固定渲染（即使暂无收藏），避免刷新时收藏加载后该分组才挂载、
             插入菜单顶部把其余菜单项整体下推造成的布局跳动；收藏项只填充到默认折叠的分组内部，不产生可见高度变化 -->
        <el-submenu index="favorites" class="fav-submenu">
          <template slot="title">
            <i class="el-icon-star-on fav-section-icon"></i>
            <span>我的收藏</span>
          </template>
          <el-menu-item v-for="(fav, idx) in favorites" :key="'fav-' + fav.menu_index" :index="'fav-' + fav.menu_index">
            <span class="fav-drag-handle" draggable="true"
                  @dragstart="onFavDragStart($event, idx)"
                  @dragend="onFavDragEnd"
                  @mousedown.stop>
              <svg viewBox="0 0 24 24" width="16" height="16" fill="currentColor">
                <circle cx="9" cy="6" r="1.5"/><circle cx="15" cy="6" r="1.5"/>
                <circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/>
                <circle cx="9" cy="18" r="1.5"/><circle cx="15" cy="18" r="1.5"/>
              </svg>
            </span>
            <svg-icon :name="fav.icon" />
            <span>{{ fav.title }}</span>
            <i
              class="fav-star el-icon-star-on is-faved"
              title="取消收藏"
              @click.stop="toggleFavorite({ index: fav.menu_index, icon: fav.icon, title: fav.title })"
            ></i>
          </el-menu-item>
          <!-- 空状态提示：用普通 div 而非 el-menu-item，避免被 el-menu 注册为菜单项 -->
          <div v-if="!favorites.length" class="fav-empty">暂无收藏，可在各模块菜单右侧点星收藏</div>
        </el-submenu>

        <template v-for="entry in menuConfig">
          <el-menu-item v-if="entry.type === 'item'" :key="'item-' + entry.index" :index="entry.index">
            <svg-icon :name="entry.icon" />
            <span slot="title">{{ entry.title }}</span>
          </el-menu-item>
          <el-submenu v-else :key="'sub-' + entry.index" :index="entry.index">
            <template slot="title">
              <svg-icon :name="entry.icon" />
              <span>{{ entry.title }}</span>
            </template>
            <el-menu-item v-for="child in entry.children" :key="child.index" :index="child.index">
              <svg-icon :name="child.icon" />
              <span>{{ child.title }}</span>
              <i
                class="fav-star"
                :class="isFavorited(child.index) ? 'el-icon-star-on is-faved' : 'el-icon-star-off'"
                :title="isFavorited(child.index) ? '取消收藏' : '收藏'"
                @click.stop="toggleFavorite(child)"
              ></i>
            </el-menu-item>
          </el-submenu>
        </template>
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
          <notification-bell ref="notificationBell" />
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
import { getMenuFavorites, toggleMenuFavorite, reorderMenuFavorites } from '@/api/menu_favorite'
import { animate, scrambleText } from 'animejs'
import SvgIcon from '@/components/SvgIcon.vue'
import NotificationBell from '@/components/NotificationBell.vue'

export default {
  name: 'Layout',
  components: { SvgIcon, NotificationBell },
  data() {
    return {
      displayName: localStorage.getItem('display_name') || '用户',
      subtitleAnimation: null,
      titleTypingTimer: null,
      favorites: [],
      favActiveIndex: null,
      dragFavIndex: -1,
      dragOverInsertIdx: -1,
      menuConfig: [
        { type: 'item', index: '/dashboard', icon: 'bar-chart-2', title: '数据看板' },
        { type: 'item', index: '/policies', icon: 'file-text', title: 'IT政策' },
        { type: 'submenu', index: 'asset', icon: 'monitor', title: '资产管理', children: [
          { index: '/assets', icon: 'list', title: '资产列表' },
          { index: '/regions', icon: 'map-pin', title: '区域管理' },
          { index: '/os-types', icon: 'layers', title: '操作系统管理' }
        ]},
        { type: 'submenu', index: 'network-security', icon: 'shield', title: '网络安全', children: [
          { index: '/topology', icon: 'network', title: '网络拓扑图' },
          { index: '/change-management', icon: 'git-branch', title: '变更管理' },
          { index: '/vulnerability-scan', icon: 'search', title: '漏洞扫描' },
          { index: '/penetration-test', icon: 'crosshair', title: '渗透测试' },
          { index: '/firewall-check', icon: 'shield-check', title: '防火墙检查' },
          { index: '/security-rectification', icon: 'pencil', title: '安全整改记录' }
        ]},
        { type: 'submenu', index: 'system-security', icon: 'settings', title: '系统安全', children: [
          { index: '/system-hardening', icon: 'hammer', title: '系统加固' },
          { index: '/patch-update', icon: 'download', title: '补丁更新' },
          { index: '/virus-control', icon: 'shield-alert', title: '病毒控制' },
          { index: '/backup-management', icon: 'database', title: '备份管理' }
        ]},
        { type: 'submenu', index: 'permission', icon: 'users', title: '用户管理', children: [
          { index: '/permissions', icon: 'key', title: '岗位权限设置' },
          { index: '/user-permissions', icon: 'user-check', title: '用户权限一览' },
          { index: '/sftp-accounts', icon: 'terminal', title: 'SFTP账号一览' },
          { index: '/monthly-check-history', icon: 'calendar-check', title: '月度检查历史' },
          { index: '/user-change-history', icon: 'user-cog', title: '用户变更记录' }
        ]},
        { type: 'submenu', index: 'third-party', icon: 'layout-grid', title: '第三方应用', children: [
          { index: '/approved-software', icon: 'check-circle', title: '核准软件目录' },
          { index: '/asset-software', icon: 'table', title: '资产对应表' },
          { index: '/quarterly-check-history', icon: 'calendar', title: '季度检查历史' }
        ]},
        { type: 'submenu', index: 'log', icon: 'scroll-text', title: '日志管理', children: [
          { index: '/login-logs', icon: 'log-in', title: '登录日志' },
          { index: '/operation-logs', icon: 'file-search', title: '操作日志' }
        ]},
        { type: 'submenu', index: 'public-service', icon: 'globe', title: '公共服务', children: [
          { index: '/form-publish', icon: 'upload-cloud', title: '表单发布' },
          { index: '/it-guide', icon: 'book-open', title: 'IT指南' }
        ]},
        { type: 'item', index: '/calendar', icon: 'calendar', title: '日程管理' },
        { type: 'submenu', index: 'memo-management', icon: 'clipboard', title: '备忘管理', children: [
          { index: '/password-vault', icon: 'lock', title: '密码本' },
          { index: '/dedicated-lines', icon: 'link', title: '专线信息' },
          { index: '/ipsec-vpn', icon: 'shield-lock', title: 'IPsec VPN' }
        ]}
      ]
    }
  },
  computed: {
    activeMenu() {
      // 收藏导航后指向收藏项自身，避免 el-menu 内部 initOpenedMenu 展开原始父级分组
      return this.favActiveIndex || this.$route.path
    }
  },
  watch: {
    '$route.path'() {
      this.$nextTick(() => {
        this.startSubtitleAnimation()
        this.startTitleTypingAnimation()
        this.expandActiveMenu()
      })
    }
  },
  mounted() {
    this.loadFavorites()
    // 容器级绑定收藏拖拽排序事件（整个收藏分组作为放置区）
    const menuEl = this.$refs.sidebarMenu && this.$refs.sidebarMenu.$el
    if (menuEl) {
      this._onFavDragOver = (e) => this.onFavDragOverContainer(e)
      this._onFavDrop = (e) => this.onFavDropContainer(e)
      menuEl.addEventListener('dragover', this._onFavDragOver)
      menuEl.addEventListener('drop', this._onFavDrop)
    }
    this.$nextTick(() => {
      this.startSubtitleAnimation()
      this.startTitleTypingAnimation()
      this.expandActiveMenu()
      // 检查是否需要弹出登录通知
      if (localStorage.getItem('show_login_notifications') === 'true') {
        localStorage.removeItem('show_login_notifications')
        this.$nextTick(() => {
          if (this.$refs.notificationBell) {
            this.$refs.notificationBell.showLoginNotifications()
          }
        })
      }
    })
  },
  beforeDestroy() {
    if (this.subtitleAnimation) {
      this.subtitleAnimation.pause()
    }
    if (this.titleTypingTimer) {
      clearTimeout(this.titleTypingTimer)
    }
    const menuEl = this.$refs.sidebarMenu && this.$refs.sidebarMenu.$el
    if (menuEl) {
      if (this._onFavDragOver) menuEl.removeEventListener('dragover', this._onFavDragOver)
      if (this._onFavDrop) menuEl.removeEventListener('drop', this._onFavDrop)
    }
  },
  methods: {
    // ============ 菜单收藏 ============
    loadFavorites() {
      getMenuFavorites().then(res => {
        // 首次加载收藏同样会挂载菜单项，触发 el-menu 的 items watcher -> initOpenedMenu，
        // 按激活项重展开其原始父级分组并收起用户当前展开的分组，需同样保护
        const restore = this.disableMenuAutoExpand()
        this.favorites = res.data || []
        this.$nextTick(restore)
      }).catch(() => {})
    },
    isFavorited(index) {
      return this.favorites.some(f => f.menu_index === index)
    },
    // 临时禁用 el-menu 的 initOpenedMenu，返回恢复函数。
    // 收藏变更会使新菜单项挂载 -> addItem 的 $set 触发 el-menu 内部 items watcher
    // -> updateActiveIndex -> initOpenedMenu，按激活项 indexPath 重展开其原始父级分组并
    // 收起用户当前展开的分组；事后恢复 openedMenus 也无法撤销已触发的收起/展开
    // 过渡动画（导致菜单抖动），故在源头禁用。
    // 注意：Vue 2 的 methods 经 initMethods 以 bind 挂为实例自有属性（不在原型链上），
    // 用 delete 会永久删除该方法（之后路由切换时 updateActiveIndex 调用 initOpenedMenu
    // 会抛 TypeError），必须保存原方法后赋值恢复；引用计数避免快速连点时把 no-op
    // 误存为原方法而永久固化。
    disableMenuAutoExpand() {
      const menu = this.$refs.sidebarMenu
      if (!menu) return () => {}
      if (!this.menuPatchDepth) {
        this.menuPatchDepth = 0
        this.origInitOpenedMenu = menu.initOpenedMenu
      }
      this.menuPatchDepth++
      menu.initOpenedMenu = () => {}
      return () => {
        if (--this.menuPatchDepth <= 0) {
          menu.initOpenedMenu = this.origInitOpenedMenu
          this.menuPatchDepth = 0
        }
      }
    },
    toggleFavorite(item) {
      const index = item.index || item.menu_index
      const wasFaved = this.isFavorited(index)
      const prev = this.favorites.slice()
      const restore = this.disableMenuAutoExpand()
      // 乐观更新本地状态
      if (wasFaved) {
        this.favorites = this.favorites.filter(f => f.menu_index !== index)
        // 若取消收藏的正是经收藏项导航的当前激活项，回退到按路由追踪，
        // 避免 default-active 继续指向已销毁的收藏项导致侧边栏无高亮；
        // 此时 initOpenedMenu 已被遮蔽，defaultActive watcher 不会引发手风琴抖动
        if (this.favActiveIndex === 'fav-' + index) {
          this.favActiveIndex = null
        }
      } else {
        this.favorites.push({ menu_index: index, icon: item.icon, title: item.title })
      }
      this.$nextTick(restore)
      // 同步后端；快速连点时较早失败的请求不能用过期快照回滚覆盖后续操作的结果，
      // 仅当仍是该 index 的最新一次操作时才回滚（回滚同样触发 items watcher，需同样保护）
      if (!this.favRequestSeq) this.favRequestSeq = {}
      const seq = (this.favRequestSeq[index] || 0) + 1
      this.favRequestSeq[index] = seq
      toggleMenuFavorite({
        menu_index: index,
        icon: item.icon,
        title: item.title,
        is_favorited: !wasFaved
      }).catch(() => {
        if (this.favRequestSeq[index] !== seq) return
        const restoreRollback = this.disableMenuAutoExpand()
        this.favorites = prev
        this.$nextTick(restoreRollback)
      })
    },
    // ============ 收藏拖动排序 ============
    onFavDragStart(e, idx) {
      this.dragFavIndex = idx
      e.dataTransfer.effectAllowed = 'move'
      const menuItemEl = e.target.closest('.el-menu-item')
      if (menuItemEl) {
        const rect = menuItemEl.getBoundingClientRect()
        // 先设置悬浮阴影，再把整个菜单项设为拖拽影像（快照会带上阴影）
        menuItemEl.style.boxShadow = '0 8px 20px rgba(64, 158, 255, 0.4)'
        e.dataTransfer.setDragImage(menuItemEl, e.clientX - rect.left, e.clientY - rect.top)
        // 快照后再把原元素置暗，表示正在被拖动
        menuItemEl.classList.add('fav-dragging')
      }
    },
    onFavDragEnd() {
      this.clearFavDragState()
    },
    onFavDragOverContainer(e) {
      if (this.dragFavIndex < 0) return
      const submenu = this.$el.querySelector('.fav-submenu')
      if (!submenu) return
      const subRect = submenu.getBoundingClientRect()
      // 仅当鼠标位于收藏分组内才处理
      if (e.clientY < subRect.top || e.clientY > subRect.bottom ||
          e.clientX < subRect.left || e.clientX > subRect.right) {
        this.clearDropIndicator()
        this.dragOverInsertIdx = -1
        return
      }
      e.preventDefault()
      e.dataTransfer.dropEffect = 'move'
      const items = Array.from(submenu.querySelectorAll('.el-menu--inline .el-menu-item'))
      // 根据鼠标垂直位置判断插入点：在项中点上方 = 插到该项之前，否则插到该项之后
      let insertIdx = items.length
      let beforeEl = null
      for (let i = 0; i < items.length; i++) {
        const rect = items[i].getBoundingClientRect()
        if (e.clientY < rect.top + rect.height / 2) {
          insertIdx = i
          beforeEl = items[i]
          break
        }
      }
      this.dragOverInsertIdx = insertIdx
      this.clearDropIndicator()
      if (beforeEl) {
        beforeEl.classList.add('fav-drop-before')
      } else if (items.length) {
        items[items.length - 1].classList.add('fav-drop-after')
      }
    },
    onFavDropContainer(e) {
      if (this.dragFavIndex < 0 || this.dragOverInsertIdx < 0) return
      e.preventDefault()
      const fromIdx = this.dragFavIndex
      let targetIdx = this.dragOverInsertIdx
      if (fromIdx < targetIdx) targetIdx--
      if (fromIdx !== targetIdx) {
        const item = this.favorites.splice(fromIdx, 1)[0]
        this.favorites.splice(targetIdx, 0, item)
        this.persistFavOrder()
      }
      this.clearFavDragState()
    },
    clearDropIndicator() {
      const submenu = this.$el.querySelector('.fav-submenu')
      if (!submenu) return
      submenu.querySelectorAll('.fav-drop-before, .fav-drop-after').forEach(el => {
        el.classList.remove('fav-drop-before', 'fav-drop-after')
      })
    },
    clearFavDragState() {
      this.clearDropIndicator()
      const submenu = this.$el.querySelector('.fav-submenu')
      if (submenu) {
        submenu.querySelectorAll('.fav-dragging').forEach(el => {
          el.classList.remove('fav-dragging')
          el.style.boxShadow = ''
        })
      }
      this.dragFavIndex = -1
      this.dragOverInsertIdx = -1
    },
    persistFavOrder() {
      const indices = this.favorites.map(f => f.menu_index)
      reorderMenuFavorites(indices).catch(() => {})
    },
    handleMenuSelect(index) {
      const isFav = index.startsWith('fav-')
      const path = isFav ? index.slice(4) : index
      if (isFav) {
        // 从收藏项导航：default-active 指向收藏项，且不自动展开原始父级菜单
        // （非收藏点击不在此处清除 favActiveIndex，避免 default-active 中间跳变引发手风琴闪烁）
        this.favActiveIndex = index
        this.navFromFav = true
      }
      if (path === this.$route.path) {
        this.navFromFav = false
        return
      }
      this.$router.push(path).catch(() => {})
    },
    // 自动展开当前激活菜单所属的子菜单（若其已在展开的子菜单中则不切换）
    expandActiveMenu() {
      if (this.navFromFav) {
        this.navFromFav = false
        return
      }
      // 路由因非收藏来源变化（点击原始菜单、浏览器前进后退等），恢复按路由追踪激活项
      this.favActiveIndex = null
      const menu = this.$refs.sidebarMenu
      if (!menu) return
      const path = this.$route.path
      const opened = menu.openedMenus || []
      const inOpened = opened.some(idx => {
        if (idx === 'favorites') return this.favorites.some(f => f.menu_index === path)
        const entry = this.menuConfig.find(e => e.index === idx)
        return entry && entry.children && entry.children.some(c => c.index === path)
      })
      if (inOpened) return
      const parent = this.menuConfig.find(e => e.type === 'submenu' && e.children && e.children.some(c => c.index === path))
      if (parent) menu.open(parent.index)
    },
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
/* 收藏拖动手柄 */
.fav-drag-handle {
  display: inline-flex;
  align-items: center;
  cursor: grab;
  color: #94A3B8;
  margin-left: -30px;
  margin-right: 12px;
  flex-shrink: 0;
  opacity: 0;
  transition: opacity 0.15s;
  line-height: 1;
}
.fav-drag-handle:hover {
  opacity: 1 !important;
  color: #E2E8F0;
}
.fav-drag-handle:active {
  cursor: grabbing;
}
</style>

<style>
/* 非看板、非日历页面恢复默认内边距 */
.el-main > *:not(.dashboard):not(.calendar-page) {
  padding: 20px;
}
/* 收藏项 hover 时显示拖动手柄 */
.sidebar-aside .el-menu--inline .el-menu-item:hover .fav-drag-handle {
  opacity: 0.5;
}
.sidebar-aside .el-menu--inline .el-menu-item:hover .fav-drag-handle:hover {
  opacity: 1;
}
/* 拖动中的原元素置暗 */
.sidebar-aside .el-menu-item.fav-dragging {
  opacity: 0.4;
}
/* 拖拽插入位置指示线 */
.sidebar-aside .el-menu-item.fav-drop-before::after,
.sidebar-aside .el-menu-item.fav-drop-after::after {
  content: '';
  position: absolute;
  left: 8px;
  right: 8px;
  height: 2px;
  background: #409EFF;
  border-radius: 1px;
  z-index: 10;
}
.sidebar-aside .el-menu-item.fav-drop-before::after {
  top: -1px;
}
.sidebar-aside .el-menu-item.fav-drop-after::after {
  bottom: -1px;
}
</style>
