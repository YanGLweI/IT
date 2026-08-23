<template>
  <el-drawer
    :visible.sync="visible"
    size="520px"
    :with-header="false"
    :wrapper-closable="true"
    append-to-body
    custom-class="entry-detail-drawer">
    <div class="drawer-content" v-if="entry">
      <!-- 头部：条目信息 + 操作 -->
      <div class="drawer-header">
        <div class="entry-title">
          <svg-icon :name="entry.icon || 'server'" :size="30" />
          <div class="entry-title-text">
            <div class="entry-name">{{ entry.name }}</div>
            <el-tag size="mini" type="info">{{ entry.category_name }}</el-tag>
          </div>
        </div>
        <div class="entry-actions">
          <el-button size="mini" icon="el-icon-edit" @click="handleEdit">编辑</el-button>
          <el-button size="mini" icon="el-icon-delete" class="danger-btn" @click="handleDelete">删除</el-button>
        </div>
      </div>

      <!-- 条目基本信息 -->
      <div class="info-grid">
        <div class="info-item" v-if="entry.url">
          <span class="info-label">URL/端口</span>
          <span class="info-value">{{ entry.url }}<span v-if="entry.port">:{{ entry.port }}</span></span>
        </div>
        <div class="info-item" v-if="entry.notes">
          <span class="info-label">备注</span>
          <span class="info-value">{{ entry.notes }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">创建人</span>
          <span class="info-value">{{ entry.created_by }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">更新人</span>
          <span class="info-value">{{ entry.updated_by }}</span>
        </div>
        <div class="info-item">
          <span class="info-label">更新时间</span>
          <span class="info-value">{{ formatDate(entry.updated_at) }}</span>
        </div>
        <div class="info-item" v-if="entry.viewers && entry.viewers.length">
          <span class="info-label">可查看用户</span>
          <span class="info-value">{{ entry.viewers.join('、') }}</span>
        </div>
      </div>

      <!-- 账号列表 -->
      <div class="section-title">账号（{{ (entry.accounts || []).length }}）</div>
      <div class="account-list">
        <div class="account-card" v-for="acc in entry.accounts" :key="acc.id">
          <div class="account-main">
            <div class="account-user">
              <el-tag v-if="acc.label" size="mini" class="account-label">{{ acc.label }}</el-tag>
              <span class="account-username" :title="acc.username">{{ acc.username }}</span>
              <el-button type="text" size="mini" icon="el-icon-document-copy" @click="copyText(acc.username)" />
            </div>
            <div class="account-meta" v-if="acc.url || acc.notes">
              <span v-if="acc.url">{{ acc.url }}<span v-if="acc.port">:{{ acc.port }}</span></span>
              <span v-if="acc.notes"><span v-if="acc.url"> · </span>{{ acc.notes }}</span>
            </div>
          </div>
          <el-button size="mini" icon="el-icon-lock" @click="handleUnlock">查看密码</el-button>
        </div>
        <div v-if="!entry.accounts || !entry.accounts.length" class="empty-tip">暂无账号</div>
      </div>
    </div>
  </el-drawer>
</template>

<script>
import SvgIcon from '@/components/SvgIcon.vue'

export default {
  name: 'EntryDetailDrawer',
  components: { SvgIcon },
  data() {
    return {
      visible: false,
      entry: null
    }
  },
  methods: {
    open(entry) {
      this.entry = entry
      this.visible = true
    },
    close() {
      this.visible = false
    },
    handleEdit() {
      this.visible = false
      this.$emit('edit', this.entry)
    },
    handleDelete() {
      this.visible = false
      this.$emit('delete', this.entry)
    },
    handleUnlock() {
      this.$emit('unlock', this.entry)
    },
    copyText(text) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
          this.$message.success('已复制')
        }).catch(() => {
          this.fallbackCopy(text)
        })
      } else {
        this.fallbackCopy(text)
      }
    },
    fallbackCopy(text) {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.left = '-9999px'
      document.body.appendChild(ta)
      ta.select()
      try {
        document.execCommand('copy')
        this.$message.success('已复制')
      } catch (e) {
        this.$message.error('复制失败，请手动复制')
      }
      document.body.removeChild(ta)
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const d = new Date(dateStr)
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    }
  }
}
</script>

<style scoped>
.drawer-content {
  padding: 24px 24px 24px;
}
.drawer-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 20px;
}
.entry-title {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}
.entry-title-text {
  min-width: 0;
}
.entry-name {
  font-size: 17px;
  font-weight: 600;
  color: #1e293b;
  margin-bottom: 6px;
  word-break: break-all;
}
.entry-actions {
  flex-shrink: 0;
}
.entry-actions .el-button {
  border-radius: 8px;
}
.danger-btn {
  color: #cf222e;
}

.info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px 16px;
  padding: 14px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  margin-bottom: 20px;
}
.info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}
.info-label {
  font-size: 12px;
  color: #94a3b8;
}
.info-value {
  font-size: 13px;
  color: #1e293b;
  word-break: break-all;
}

.section-title {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 12px;
}
.account-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}
.account-card {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 14px;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  transition: border-color 0.15s ease;
}
.account-card:hover {
  border-color: #cbd5e1;
}
.account-main {
  flex: 1;
  min-width: 0;
}
.account-user {
  display: flex;
  align-items: center;
  gap: 6px;
}
.account-label {
  border-radius: 6px;
  border: none;
  background: #e2e8f0;
  color: #475569;
}
.account-username {
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.account-user .el-button {
  flex-shrink: 0;
  padding: 0;
}
.account-meta {
  font-size: 12px;
  color: #64748b;
  margin-top: 4px;
  word-break: break-all;
}
.account-card .el-button {
  flex-shrink: 0;
  border-radius: 8px;
}
.empty-tip {
  font-size: 13px;
  color: #94a3b8;
  text-align: center;
  padding: 20px 0;
}
</style>
