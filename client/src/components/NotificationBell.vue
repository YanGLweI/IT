<template>
  <div class="notification-bell">
    <el-popover
      placement="bottom-end"
      width="360"
      trigger="click"
      v-model="popoverVisible"
    >
      <div class="notification-panel">
        <div class="panel-header">
          <span>日程通知</span>
          <el-button type="text" size="mini" @click="markAllRead" v-if="notifications.length > 0">全部已读</el-button>
        </div>
        <div class="notification-list" v-if="notifications.length > 0">
          <div
            v-for="n in notifications"
            :key="n.id"
            class="notification-item"
            :class="{ unread: !n.read_at }"
            @click="handleClickNotification(n)"
          >
            <div class="notif-icon">
              <i class="el-icon-bell"></i>
            </div>
            <div class="notif-content">
              <div class="notif-title">{{ n.calendar_title }}</div>
              <div class="notif-time">
                {{ n.is_all_day ? '全天' : formatTime(n.start_time) }}
                <span v-if="!n.read_at" class="unread-dot"></span>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="no-notifications">
          <i class="el-icon-bell" style="font-size: 24px; color: #c0c4cc"></i>
          <p>暂无通知</p>
        </div>
      </div>

      <el-badge slot="reference" :value="unreadCount" :hidden="unreadCount === 0" :max="99" class="bell-badge">
        <i class="el-icon-bell bell-icon" @click="togglePopover"></i>
      </el-badge>
    </el-popover>
  </div>
</template>

<script>
import { getTodayNotifications, getUnreadCount, getPendingNotifications, markNotificationRead, markNotificationPopupShown } from '@/api/calendar'

export default {
  name: 'NotificationBell',
  data() {
    return {
      notifications: [],
      unreadCount: 0,
      popoverVisible: false,
      pollTimer: null,
      pendingTimer: null,
      shownNotificationIds: new Set()
    }
  },
  mounted() {
    this.fetchUnreadCount()
    this.fetchTodayNotifications()
    // 每60秒轮询未读数量
    this.pollTimer = setInterval(() => {
      this.fetchUnreadCount()
    }, 60000)
    // 每60秒检查待通知的日程
    this.pendingTimer = setInterval(() => {
      this.checkPendingNotifications()
    }, 60000)
    // 首次检查
    this.checkPendingNotifications()
  },
  beforeDestroy() {
    if (this.pollTimer) clearInterval(this.pollTimer)
    if (this.pendingTimer) clearInterval(this.pendingTimer)
  },
  methods: {
    async fetchUnreadCount() {
      try {
        const res = await getUnreadCount()
        if (res && res.code === 200) {
          this.unreadCount = (res.data && res.data.count) || 0
        }
      } catch (err) {
        // ignore
      }
    },
    async fetchTodayNotifications() {
      try {
        const res = await getTodayNotifications()
        if (res && res.code === 200) {
          this.notifications = res.data || []
        }
      } catch (err) {
        // ignore
      }
    },
    async checkPendingNotifications() {
      try {
        const res = await getPendingNotifications()
        if (res && res.code === 200) {
          const pending = res.data || []
          for (const n of pending) {
            // 后端已过滤 notify_time <= now，直接弹框
            this.showPopupNotification(n)
            markNotificationPopupShown(n.id).catch(() => {})
          }
        }
      } catch (err) {
        // ignore
      }
    },
    showPopupNotification(n) {
      const timeStr = n.is_all_day ? '全天' : this.formatTime(n.start_time)
      this.$notify({
        title: '日程提醒',
        message: `${n.calendar_title} - ${timeStr}`,
        type: 'info',
        duration: 0,
        position: 'bottom-right',
        onClick: () => {
          markNotificationRead(n.id).catch(() => {})
          this.fetchUnreadCount()
        }
      })
    },
    async handleClickNotification(n) {
      if (!n.read_at) {
        try {
          await markNotificationRead(n.id)
          // 从列表中移除已读通知
          this.notifications = this.notifications.filter(item => item.id !== n.id)
          this.unreadCount = Math.max(0, this.unreadCount - 1)
        } catch (err) {
          // ignore
        }
      }
    },
    async markAllRead() {
      try {
        for (const n of this.notifications) {
          if (!n.read_at) {
            await markNotificationRead(n.id)
          }
        }
        // 清空列表和未读数
        this.notifications = []
        this.unreadCount = 0
      } catch (err) {
        // ignore
      }
    },
    togglePopover() {
      if (!this.popoverVisible) {
        // 打开时实时请求最新通知
        this.fetchTodayNotifications()
        this.fetchUnreadCount()
      }
    },
    formatTime(timeStr) {
      if (!timeStr) return ''
      const d = new Date(timeStr)
      return `${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    // 供Login.vue调用：登录时弹出通知
    async showLoginNotifications() {
      await this.fetchTodayNotifications()
      const unread = this.notifications.filter(n => !n.read_at)
      for (let i = 0; i < unread.length; i++) {
        setTimeout(() => {
          this.showPopupNotification(unread[i])
          this.shownNotificationIds.add(unread[i].id)
          markNotificationPopupShown(unread[i].id).catch(() => {})
        }, i * 500)
      }
    }
  }
}
</script>

<style scoped>
.notification-bell {
  display: inline-flex;
  align-items: center;
  margin-right: 16px;
}

.bell-icon {
  font-size: 20px;
  color: #606266;
  cursor: pointer;
  transition: color 0.2s;
}

.bell-icon:hover {
  color: #409eff;
}

.notification-panel {
  margin: -12px;
}

.panel-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
  font-size: 14px;
}

.notification-list {
  max-height: 400px;
  overflow-y: auto;
}

.notification-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
  transition: background 0.2s;
}

.notification-item:hover {
  background: #f5f7fa;
}

.notification-item.unread {
  background: #ecf5ff;
}

.notif-icon {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: #409eff;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.notif-content {
  flex: 1;
  min-width: 0;
}

.notif-title {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.notif-time {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.unread-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #409eff;
  display: inline-block;
}

.no-notifications {
  text-align: center;
  padding: 40px 0;
  color: #c0c4cc;
}

.no-notifications p {
  margin-top: 8px;
  font-size: 14px;
}
</style>
