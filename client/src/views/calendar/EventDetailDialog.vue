<template>
  <el-dialog
    title="日程详情"
    :visible.sync="dialogVisible"
    width="480px"
    :close-on-click-modal="true"
    custom-class="event-detail-dialog"
    @close="handleClose"
  >
    <div v-if="event" class="event-detail">
      <h2 class="detail-title">{{ event.title }}</h2>

      <div class="detail-section">
        <div class="detail-label">
          <i class="el-icon-time"></i> 时间
        </div>
        <div class="detail-value">
          <template v-if="event.is_all_day">
            全天 · {{ formatDate(event.start_time) }} - {{ formatDate(event.end_time) }}
          </template>
          <template v-else>
            {{ formatDateTime(event.start_time) }} - {{ formatDateTime(event.end_time) }}
          </template>
        </div>
      </div>

      <div class="detail-section" v-if="repeatRuleText">
        <div class="detail-label">
          <i class="el-icon-refresh"></i> 重复
        </div>
        <div class="detail-value">{{ repeatRuleText }}</div>
      </div>

      <div class="detail-section" v-if="event.participants && event.participants.length > 0">
        <div class="detail-label">
          <i class="el-icon-user"></i> 参与者
        </div>
        <div class="detail-value">
          <div v-for="p in event.participants" :key="p.id" class="participant-item">
            <span class="participant-avatar-lg" :style="{ background: getAvatarColor(p.id) }">
              {{ p.display_name ? p.display_name[0] : '?' }}
            </span>
            <span class="participant-name">{{ p.display_name || p.user_dn }}</span>
          </div>
        </div>
      </div>

      <div class="detail-section" v-if="event.created_by">
        <div class="detail-label">
          <i class="el-icon-edit"></i> 创建者
        </div>
        <div class="detail-value">{{ event.created_by }}</div>
      </div>

      <div class="detail-section" v-if="event.created_at">
        <div class="detail-label">
          <i class="el-icon-date"></i> 创建时间
        </div>
        <div class="detail-value">{{ formatDateTime(event.created_at) }}</div>
      </div>

      <div class="detail-section" v-if="event.description">
        <div class="detail-label">
          <i class="el-icon-document"></i> 描述
        </div>
        <div class="detail-value description-text">{{ event.description }}</div>
      </div>
    </div>

    <div slot="footer" class="dialog-footer">
      <el-button @click="handleClose">关闭</el-button>
      <el-button v-if="isCreator" type="primary" @click="handleEdit">编辑</el-button>
      <el-button v-if="isCreator" type="danger" @click="handleDelete">删除</el-button>
    </div>
  </el-dialog>
</template>

<script>
import { deleteCalendar } from '@/api/calendar'

export default {
  name: 'EventDetailDialog',
  props: {
    visible: { type: Boolean, default: false },
    event: { type: Object, default: null }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(val) { this.$emit('close') }
    },
    isCreator() {
      if (!this.event) return false
      const currentUser = localStorage.getItem('username') || ''
      return this.event.created_by === currentUser
    },
    repeatRuleText() {
      if (!this.event || !this.event.repeat_rule_json) return ''
      try {
        const rule = JSON.parse(this.event.repeat_rule_json)
        const weekdayNames = ['日', '一', '二', '三', '四', '五', '六']
        const ordinals = ['一', '二', '三', '四', '五']
        switch (rule.type) {
          case 'daily': return rule.interval === 1 ? '每天' : `每${rule.interval}天`
          case 'weekly': return rule.interval === 1 ? `每周${weekdayNames[rule.weekday]}` : `每${rule.interval}周周${weekdayNames[rule.weekday]}`
          case 'monthly_week': return (rule.interval === 1 ? '每月' : `每${rule.interval}个月`) + `第${ordinals[rule.weekOfMonth - 1]}个周${weekdayNames[rule.weekday]}`
          case 'monthly_day': return rule.interval === 1 ? `每月${rule.monthDay}日` : `每${rule.interval}个月${rule.monthDay}日`
          case 'yearly': return `每年${rule.monthOfYear}月${rule.monthDay}日`
          case 'workday': return '每个工作日（周一至周五）'
          case 'custom': {
            if (rule.unit === 'weeks' && rule.weekdays && rule.weekdays.length > 0) {
              const names = rule.weekdays.map(d => '周' + weekdayNames[d]).join('、')
              return `每${rule.interval}周（${names}）`
            }
            return `每${rule.interval}${this.getUnitName(rule.unit)}`
          }
          default: return ''
        }
      } catch {
        return ''
      }
    }
  },
  methods: {
    getAvatarColor(id) {
      const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
      return colors[(id || 0) % colors.length]
    },
    getUnitName(unit) {
      const map = { days: '天', weeks: '周', months: '个月', years: '年' }
      return map[unit] || unit
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      const d = new Date(dateStr)
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    },
    formatDateTime(dateStr) {
      if (!dateStr) return ''
      const d = new Date(dateStr)
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')} ${String(d.getHours()).padStart(2, '0')}:${String(d.getMinutes()).padStart(2, '0')}`
    },
    handleEdit() {
      this.$emit('edit', this.event)
    },
    async handleDelete() {
      try {
        await this.$confirm('确定要删除此日程吗？', '确认删除', { type: 'warning' })
      } catch {
        return
      }
      try {
        const res = await deleteCalendar(this.event.id)
        if (res && res.code === 200) {
          this.$message.success('删除成功')
          this.$emit('deleted')
        } else {
          this.$message.error(res ? res.message : '删除失败')
        }
      } catch (err) {
        this.$message.error('删除失败')
      }
    },
    handleClose() {
      this.$emit('close')
    }
  }
}
</script>

<style>
.event-detail-dialog.el-dialog {
  border-radius: 16px;
  overflow: hidden;
}

.event-detail-dialog .el-dialog__header {
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
  margin: 0;
}

.event-detail-dialog .el-dialog__title {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
}

.event-detail-dialog .el-dialog__body {
  padding: 20px;
}

.event-detail-dialog .el-dialog__footer {
  padding: 12px 20px;
  border-top: 1px solid #f1f5f9;
}
</style>

<style scoped>
.event-detail {
  padding: 0;
}

.detail-title {
  font-size: 20px;
  font-weight: 700;
  color: #1e293b;
  margin: 0 0 16px 0;
  padding-bottom: 12px;
  border-bottom: 1px solid #f1f5f9;
  letter-spacing: -0.01em;
}

.detail-section {
  margin-bottom: 14px;
}

.detail-label {
  font-size: 12px;
  color: #94a3b8;
  margin-bottom: 4px;
  font-weight: 500;
}

.detail-label i {
  margin-right: 4px;
}

.detail-value {
  font-size: 14px;
  color: #334155;
  line-height: 1.5;
}

.description-text {
  white-space: pre-wrap;
  color: #475569;
  background: #f8fafc;
  padding: 10px 12px;
  border-radius: 8px;
  margin-top: 4px;
}

.participant-item {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 8px;
}

.participant-avatar-lg {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 600;
  flex-shrink: 0;
}

.participant-name {
  font-size: 14px;
  color: #334155;
}
</style>
