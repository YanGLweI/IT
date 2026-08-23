<template>
  <el-dialog :visible.sync="visible" :title="entryName" width="480px" :close-on-click-modal="false" @close="handleClose">
    <div v-if="!unlocked" class="unlock-verify">
      <p class="unlock-hint">请输入您的 LDAP 密码以验证身份</p>
      <el-input ref="ldapInput" v-model="ldapPassword" type="password" placeholder="LDAP 密码" show-password @keyup.enter.native="handleUnlock" />
      <div v-if="errorMsg" class="error-msg">{{ errorMsg }}</div>
    </div>
    <div v-else class="unlock-result">
      <div class="account-item" v-for="acc in accounts" :key="acc.id">
        <div class="account-name">
          <el-tag v-if="acc.label" size="mini" type="info">{{ acc.label }}</el-tag>
          <span>{{ acc.username }}</span>
        </div>
        <div class="password-display">
          <code>{{ acc.password }}</code>
        </div>
        <div class="password-actions">
          <el-button type="primary" size="small" icon="el-icon-document-copy" @click="copyPassword(acc.password)">复制密码</el-button>
        </div>
      </div>
      <div class="countdown-row">
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
      accounts: [],
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
      this.accounts = []
      this.unlocked = false
      this.loading = false
      this.errorMsg = ''
      this.countdown = 30
      if (this.timer) clearInterval(this.timer)
      this.visible = true
      this.$nextTick(() => {
        this.$refs.ldapInput && this.$refs.ldapInput.focus()
      })
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
          this.accounts = (res.data && res.data.accounts) || []
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
    copyPassword(password) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(password).then(() => {
          this.$message.success('已复制')
          this.handleClose()
        }).catch(() => {
          this.fallbackCopy(password)
        })
      } else {
        this.fallbackCopy(password)
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
      this.accounts = []
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
.account-item {
  margin-bottom: 16px;
}
.account-item:last-child {
  margin-bottom: 0;
}
.account-name {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 500;
  color: #475569;
  margin-bottom: 8px;
}
.password-display {
  padding: 12px 16px;
  background: #f1f5f9;
  border-radius: 10px;
  margin-bottom: 8px;
  border: 1px solid #e2e8f0;
}
.password-display code {
  font-family: 'Maple Mono NF', monospace;
  font-size: 16px;
  color: #1e293b;
  word-break: break-all;
  letter-spacing: 0.5px;
}
.password-actions .el-button {
  border-radius: 8px;
}
.countdown-row {
  margin-top: 16px;
  text-align: right;
}
.countdown {
  font-size: 13px;
  color: #64748b;
}
.countdown.warning {
  color: #ef4444;
}
</style>
