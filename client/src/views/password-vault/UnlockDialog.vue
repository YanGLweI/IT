<template>
  <el-dialog :visible.sync="visible" :title="entryName" width="420px" :close-on-click-modal="false" @close="handleClose">
    <div v-if="!unlocked" class="unlock-verify">
      <p class="unlock-hint">请输入您的 LDAP 密码以验证身份</p>
      <el-input v-model="ldapPassword" type="password" placeholder="LDAP 密码" show-password @keyup.enter.native="handleUnlock" />
      <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>
    </div>
    <div v-else class="unlock-result">
      <div class="password-display">
        <code>{{ password }}</code>
      </div>
      <div class="password-actions">
        <el-button type="primary" size="small" icon="el-icon-document-copy" @click="copyPassword">复制密码</el-button>
        <span class="countdown" :class="{ warning: countdown <= 10 }">{{ countdown }}s 后自动隐藏</span>
      </div>
    </div>
    <span slot="footer">
      <el-button @click="handleClose">关闭</el-button>
      <el-button v-if="!unlocked" type="primary" :loading="loading" @click="handleUnlock">验证</el-button>
    </span>
  </el-dialog>
</template>

<script>
import { encryptPassword } from '@/utils/rsa'
import { unlockPasswordEntry } from '@/api/password_vault'

export default {
  name: 'UnlockDialog',
  data() {
    return {
      visible: false,
      entryId: null,
      entryName: '',
      ldapPassword: '',
      password: '',
      unlocked: false,
      loading: false,
      errorMsg: '',
      countdown: 30,
      timer: null
    }
  },
  methods: {
    open(entry) {
      this.entryId = entry.id
      this.entryName = entry.name
      this.ldapPassword = ''
      this.password = ''
      this.unlocked = false
      this.loading = false
      this.errorMsg = ''
      this.countdown = 30
      if (this.timer) clearInterval(this.timer)
      this.visible = true
    },
    async handleUnlock() {
      if (!this.ldapPassword) {
        this.errorMsg = '请输入密码'
        return
      }
      this.loading = true
      this.errorMsg = ''
      try {
        const encrypted = await encryptPassword(this.ldapPassword)
        const res = await unlockPasswordEntry(this.entryId, encrypted)
        if (res && res.code === 200) {
          this.password = res.data.password
          this.unlocked = true
          this.startCountdown()
        } else {
          this.errorMsg = res.message || '验证失败'
        }
      } catch (err) {
        this.errorMsg = err.response?.data?.message || '验证失败'
      } finally {
        this.loading = false
      }
    },
    startCountdown() {
      this.countdown = 30
      this.timer = setInterval(() => {
        this.countdown--
        if (this.countdown <= 0) {
          this.handleClose()
        }
      }, 1000)
    },
    copyPassword() {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(this.password).then(() => {
          this.$message.success('已复制')
          this.handleClose()
        }).catch(() => {
          this.fallbackCopy(this.password)
        })
      } else {
        this.fallbackCopy(this.password)
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
        this.handleClose()
      } catch (e) {
        this.$message.error('复制失败，请手动复制')
      }
      document.body.removeChild(ta)
    },
    handleClose() {
      if (this.timer) clearInterval(this.timer)
      this.visible = false
      this.ldapPassword = ''
      this.password = ''
    }
  }
}
</script>

<style scoped>
.unlock-hint {
  font-size: 14px;
  color: #475569;
  margin-bottom: 16px;
}
.error-msg {
  color: #ef4444;
  font-size: 13px;
  margin-top: 10px;
  padding: 8px 12px;
  background: #fef2f2;
  border-radius: 8px;
}
.password-display {
  padding: 16px 20px;
  background: #f1f5f9;
  border-radius: 10px;
  margin-bottom: 16px;
  border: 1px solid #e2e8f0;
}
.password-display code {
  font-family: 'Maple Mono NF', monospace;
  font-size: 18px;
  color: #1e293b;
  word-break: break-all;
  letter-spacing: 0.5px;
}
.password-actions {
  display: flex;
  align-items: center;
  gap: 16px;
}
.password-actions .el-button {
  border-radius: 8px;
}
.countdown {
  font-size: 13px;
  color: #64748b;
}
.countdown.warning {
  color: #ef4444;
}
</style>
