<template>
  <div class="password-generator">
    <div class="generator-options">
      <div class="option-row">
        <label>密码长度</label>
        <el-slider v-model="length" :min="8" :max="20" show-input :step="1" />
      </div>
      <div class="option-row checkbox-row">
        <el-checkbox v-model="useUppercase">大写字母 (A-Z)</el-checkbox>
        <el-checkbox v-model="useLowercase">小写字母 (a-z)</el-checkbox>
        <el-checkbox v-model="useNumbers">数字 (0-9)</el-checkbox>
        <el-checkbox v-model="useSymbols">特殊符号 (!@#$...)</el-checkbox>
      </div>
      <div v-if="generatedPassword" class="generated-preview">
        <code>{{ generatedPassword }}</code>
        <el-button type="text" size="mini" icon="el-icon-document-copy" @click="copyGenerated">复制</el-button>
      </div>
    </div>
    <div class="generator-actions">
      <el-button type="primary" size="small" @click="generate" :disabled="!canGenerate">生成密码</el-button>
      <el-button size="small" @click="applyAndClose" :disabled="!generatedPassword">使用此密码</el-button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PasswordGenerator',
  data() {
    return {
      length: 14,
      useUppercase: true,
      useLowercase: true,
      useNumbers: true,
      useSymbols: true,
      generatedPassword: ''
    }
  },
  computed: {
    canGenerate() {
      let count = 0
      if (this.useUppercase) count++
      if (this.useLowercase) count++
      if (this.useNumbers) count++
      if (this.useSymbols) count++
      return count >= 2
    }
  },
  methods: {
    generate() {
      let chars = ''
      const required = []
      if (this.useUppercase) {
        chars += 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
        required.push('ABCDEFGHIJKLMNOPQRSTUVWXYZ')
      }
      if (this.useLowercase) {
        chars += 'abcdefghijklmnopqrstuvwxyz'
        required.push('abcdefghijklmnopqrstuvwxyz')
      }
      if (this.useNumbers) {
        chars += '0123456789'
        required.push('0123456789')
      }
      if (this.useSymbols) {
        chars += '!@#$%^&*()_+-=[]{}|;:,.<>?'
        required.push('!@#$%^&*()_+-=[]{}|;:,.<>?')
      }

      const array = new Uint32Array(this.length)
      crypto.getRandomValues(array)

      let password = ''
      // 确保每种字符至少出现一次
      for (let i = 0; i < required.length && i < this.length; i++) {
        const charSet = required[i]
        password += charSet[array[i] % charSet.length]
      }
      // 填充剩余长度
      for (let i = required.length; i < this.length; i++) {
        password += chars[array[i] % chars.length]
      }
      // 打乱顺序
      const shuffleArray = new Uint32Array(password.length)
      crypto.getRandomValues(shuffleArray)
      password = password.split('').sort((a, b) => shuffleArray[password.indexOf(a)] - shuffleArray[password.indexOf(b)]).join('')

      this.generatedPassword = password
    },
    copyGenerated() {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(this.generatedPassword).then(() => {
          this.$message.success('已复制')
        }).catch(() => {
          this.fallbackCopy(this.generatedPassword)
        })
      } else {
        this.fallbackCopy(this.generatedPassword)
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
    applyAndClose() {
      this.$emit('apply', this.generatedPassword)
    }
  }
}
</script>

<style scoped>
.password-generator {
  padding: 8px 0;
}
.option-row {
  margin-bottom: 16px;
}
.option-row label {
  display: block;
  font-size: 13px;
  color: #475569;
  margin-bottom: 8px;
  font-weight: 600;
}
.checkbox-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}
.checkbox-row .el-checkbox {
  margin-right: 0;
}
.generated-preview {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 16px;
  background: #f1f5f9;
  border-radius: 10px;
  margin-top: 12px;
  border: 1px solid #e2e8f0;
}
.generated-preview code {
  flex: 1;
  font-family: 'Maple Mono NF', monospace;
  font-size: 14px;
  color: #1e293b;
  word-break: break-all;
  letter-spacing: 0.3px;
}
.generator-actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}
.generator-actions .el-button {
  border-radius: 8px;
}
</style>
