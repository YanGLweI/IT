<template>
  <el-dialog :visible.sync="visible" :title="isEdit ? '编辑密码条目' : '新增密码条目'" width="880px" :close-on-click-modal="false" custom-class="password-entry-dialog">
    <div class="dialog-body">
      <!-- 左列：图标选择器 -->
      <div class="left-panel">
        <div class="panel-title">选择图标</div>
        <IconPicker v-model="form.icon" />
      </div>
      <!-- 右列：表单 -->
      <div class="right-panel" ref="rightPanel">
        <el-form :model="form" :rules="rules" ref="form" label-width="80px" size="small">
          <el-form-item label="名称" prop="name">
            <el-input v-model="form.name" placeholder="如：生产数据库主库" />
          </el-form-item>
          <el-form-item label="分类" prop="category_id">
            <el-select v-model="form.category_id" placeholder="选择分类" style="width: 100%">
              <el-option v-for="cat in categories" :key="cat.id" :label="cat.name" :value="cat.id" />
            </el-select>
          </el-form-item>
          <el-row :gutter="12">
            <el-col :span="16">
              <el-form-item label="URL/IP">
                <el-input v-model="form.url" placeholder="URL 或 IP 地址" />
              </el-form-item>
            </el-col>
            <el-col :span="8">
              <el-form-item label="端口">
                <el-input v-model.number="form.port" placeholder="端口"/>
              </el-form-item>
            </el-col>
          </el-row>
          <el-form-item label="备注">
            <el-input v-model="form.notes" type="textarea" :rows="2" placeholder="备注信息" />
          </el-form-item>
          <el-form-item label="可查看用户">
            <el-select v-model="form.viewers" multiple filterable remote reserve-keyword :remote-method="searchUsers" placeholder="搜索 LDAP 用户" style="width: 100%" :disabled="!isCreator">
              <el-option v-for="u in ldapUsers" :key="u.sAMAccountName" :label="u.display_name + ' (' + u.sAMAccountName + ')'" :value="u.sAMAccountName" />
            </el-select>
          </el-form-item>

          <!-- 账号列表 -->
          <el-form-item label="账号" prop="accounts">
            <div class="account-editor">
              <div class="account-card" v-for="(acc, index) in form.accounts" :key="index">
                <div class="account-card-header">
                  <span class="account-card-title">账号 {{ index + 1 }}</span>
                  <el-button type="text" size="mini" icon="el-icon-delete" class="remove-account-btn" :disabled="form.accounts.length <= 1" @click="removeAccount(index)">移除</el-button>
                </div>
                <el-row :gutter="10">
                  <el-col :span="9">
                    <el-input v-model="acc.label" placeholder="标签（如 root / 业务账号）" />
                  </el-col>
                  <el-col :span="15">
                    <el-input v-model="acc.username" placeholder="用户名/账号（必填）" />
                  </el-col>
                </el-row>
                <div class="password-input-group">
                  <el-input v-model="acc.password" show-password :placeholder="acc.id ? '密码（留空则不修改）' : '密码（必填）'" />
                  <el-button size="small" type="text" @click="toggleGenerator(index)">
                    <i class="el-icon-magic-stick" /> 生成
                  </el-button>
                </div>
                <div v-if="generatorIndex === index" class="generator-wrapper">
                  <PasswordGenerator @apply="applyGeneratedPassword" />
                </div>
                <el-row :gutter="10">
                  <el-col :span="16">
                    <el-input v-model="acc.url" placeholder="URL/IP（可覆盖条目级）" />
                  </el-col>
                  <el-col :span="8">
                    <el-input v-model.number="acc.port" placeholder="端口" />
                  </el-col>
                </el-row>
                <el-input v-model="acc.notes" placeholder="账号备注" />
              </div>
              <el-button type="dashed" size="small" icon="el-icon-plus" class="add-account-btn" @click="addAccount">添加账号</el-button>
            </div>
          </el-form-item>
        </el-form>
      </div>
    </div>
    <span slot="footer">
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSave">保存</el-button>
    </span>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </el-dialog>
</template>

<script>
import IconPicker from './IconPicker.vue'
import PasswordGenerator from './PasswordGenerator.vue'
import DualControlDialog from '@/components/DualControlDialog.vue'
import { createPasswordEntry, updatePasswordEntry } from '@/api/password_vault'
import { encryptPassword } from '@/utils/rsa'
import request from '@/api/request'

const emptyAccount = () => ({ id: null, label: '', username: '', password: '', url: '', port: null, notes: '' })

export default {
  name: 'PasswordEntryDialog',
  components: { IconPicker, PasswordGenerator, DualControlDialog },
  props: {
    categories: { type: Array, default: () => [] }
  },
  data() {
    const validateAccounts = (rule, value, callback) => {
      if (!value || value.length === 0) {
        return callback(new Error('请至少添加一个账号'))
      }
      for (let i = 0; i < value.length; i++) {
        const acc = value[i]
        if (!acc.username) {
          return callback(new Error(`账号 ${i + 1}：请填写账号名`))
        }
        if (!acc.password && !acc.id) {
          return callback(new Error(`账号 ${i + 1}：请填写密码`))
        }
      }
      callback()
    }
    return {
      visible: false,
      isEdit: false,
      editId: null,
      loading: false,
      generatorIndex: null,
      isCreator: true,
      ldapUsers: [],
      form: {
        icon: 'server',
        name: '',
        category_id: null,
        url: '',
        port: null,
        notes: '',
        viewers: [],
        accounts: [emptyAccount()]
      },
      rules: {
        name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
        category_id: [{ required: true, message: '请选择分类', trigger: 'change' }],
        accounts: [{ validator: validateAccounts, trigger: 'change' }]
      }
    }
  },
  methods: {
    open(entry) {
      this.generatorIndex = null
      this.ldapUsers = []
      if (entry) {
        this.isEdit = true
        this.editId = entry.id
        this.isCreator = entry.is_creator !== false
        const accounts = (entry.accounts || []).map(a => ({
          id: a.id,
          label: a.label || '',
          username: a.username,
          password: '',
          url: a.url || '',
          port: a.port || null,
          notes: a.notes || ''
        }))
        this.form = {
          icon: entry.icon || 'server',
          name: entry.name,
          category_id: entry.category_id,
          url: entry.url || '',
          port: entry.port || null,
          notes: entry.notes || '',
          viewers: entry.viewers ? [...entry.viewers] : [],
          accounts: accounts.length ? accounts : [emptyAccount()]
        }
      } else {
        this.isEdit = false
        this.editId = null
        this.isCreator = true
        this.form = { icon: 'server', name: '', category_id: null, url: '', port: null, notes: '', viewers: [], accounts: [emptyAccount()] }
      }
      this.visible = true
      this.$nextTick(() => {
        this.$refs.form?.clearValidate()
        this.loadAllUsers()
      })
    },
    addAccount() {
      this.form.accounts.push(emptyAccount())
      // 新增账号后自动滚动到右侧面板底部，便于填写新卡片
      this.$nextTick(() => {
        const panel = this.$refs.rightPanel
        if (panel) panel.scrollTop = panel.scrollHeight
      })
    },
    removeAccount(index) {
      if (this.form.accounts.length <= 1) return
      this.form.accounts.splice(index, 1)
      if (this.generatorIndex === index) this.generatorIndex = null
    },
    toggleGenerator(index) {
      this.generatorIndex = this.generatorIndex === index ? null : index
    },
    applyGeneratedPassword(password) {
      if (this.generatorIndex === null) return
      this.$set(this.form.accounts[this.generatorIndex], 'password', password)
      this.generatorIndex = null
    },
    async loadAllUsers() {
      try {
        const res = await request.get('/ldap/users')
        if (res && res.code === 200) {
          this.ldapUsers = res.data || []
        }
      } catch (e) { /* ignore */ }
    },
    searchUsers(query) {
      if (!query) {
        this.loadAllUsers()
        return
      }
      const q = query.toLowerCase()
      this.ldapUsers = (this.ldapUsers || []).filter(u =>
        u.sAMAccountName?.toLowerCase().includes(q) || u.display_name?.toLowerCase().includes(q)
      )
    },
    async handleSave() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        this.loading = true
        try {
          const dualToken = await this.$refs.dualControl.open()

          const data = {
            icon: this.form.icon,
            name: this.form.name,
            category_id: this.form.category_id,
            url: this.form.url,
            port: this.form.port,
            notes: this.form.notes,
            viewers: this.form.viewers,
            accounts: []
          }
          // 逐个账号 RSA 加密密码
          for (let i = 0; i < this.form.accounts.length; i++) {
            const acc = this.form.accounts[i]
            const item = {
              id: acc.id || undefined,
              label: acc.label,
              username: acc.username,
              url: acc.url,
              port: acc.port,
              notes: acc.notes,
              sort_order: i
            }
            if (acc.password) {
              item.encrypted_password = await encryptPassword(acc.password)
            }
            data.accounts.push(item)
          }

          if (this.isEdit) {
            await updatePasswordEntry(this.editId, data, dualToken)
            this.$message.success('更新成功')
          } else {
            await createPasswordEntry(data, dualToken)
            this.$message.success('创建成功')
          }
          this.visible = false
          this.$emit('saved')
        } catch (err) {
          if (err.message !== 'canceled') {
            this.$message.error(err.response?.data?.message || '操作失败')
          }
        } finally {
          this.loading = false
        }
      })
    }
  }
}
</script>

<style scoped>
.dialog-body {
  display: flex;
  gap: 24px;
  height: 65vh;
  overflow: hidden;
}
.left-panel {
  width: 260px;
  flex-shrink: 0;
  align-self: stretch;
  display: flex;
  flex-direction: column;
  padding-right: 5px;
  border-right: 1px solid #e2e8f0;
  overflow: hidden;
}
/* 图标选择区撑满左侧栏剩余高度，独立滚动 */
.left-panel ::v-deep .icon-picker {
  max-height: none;
  flex: 1;
  min-height: 0;
}
.panel-title {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 12px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.right-panel {
  flex: 1;
  min-width: 0;
  overflow-y: auto;
  padding-right: 6px;
}
.password-input-group {
  display: flex;
  align-items: center;
  gap: 10px;
}
.password-input-group .el-input {
  flex: 1;
}
.password-input-group .el-button {
  color: #475569;
}
.password-input-group .el-button:hover {
  color: #2563eb;
}
.generator-wrapper {
  margin: 10px 0;
  padding: 16px;
  background: #f1f5f9;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}

/* 账号编辑器 */
.account-editor {
  width: 100%;
}
.account-card {
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 12px 14px;
  margin-bottom: 10px;
  background: #fbfcfe;
}
.account-card .el-row {
  margin-bottom: 8px;
}
.account-card .el-row:last-child {
  margin-bottom: 0;
}
.account-card .password-input-group {
  margin-bottom: 8px;
}
.account-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 10px;
}
.account-card-title {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
}
.remove-account-btn {
  color: #cf222e;
  padding: 0;
}
.remove-account-btn.is-disabled {
  color: #cbd5e1;
}
.add-account-btn {
  width: 100%;
  border-radius: 10px;
  color: #475569;
}
.add-account-btn:hover {
  color: #2563eb;
  border-color: #93c5fd;
}
</style>
