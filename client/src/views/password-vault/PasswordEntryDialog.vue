<template>
  <el-dialog :visible.sync="visible" :title="isEdit ? '编辑密码条目' : '新增密码条目'" width="880px" :close-on-click-modal="false" custom-class="password-entry-dialog">
    <div class="dialog-body">
      <!-- 左列：图标选择器 -->
      <div class="left-panel">
        <div class="panel-title">选择图标</div>
        <IconPicker v-model="form.icon" />
      </div>
      <!-- 右列：表单 -->
      <div class="right-panel">
        <el-form :model="form" :rules="rules" ref="form" label-width="80px" size="small">
          <el-form-item label="名称" prop="name">
            <el-input v-model="form.name" placeholder="如：生产数据库主库" />
          </el-form-item>
          <el-form-item label="账号" prop="username">
            <el-input v-model="form.username" placeholder="用户名/账号" />
          </el-form-item>
          <el-form-item label="密码" prop="password">
            <div class="password-input-group">
              <el-input v-model="form.password" show-password placeholder="密码">
              </el-input>
              <el-button size="small" type="text" @click="showGenerator = !showGenerator">
                <i class="el-icon-magic-stick" /> 生成
              </el-button>
            </div>
            <div v-if="showGenerator" class="generator-wrapper">
              <PasswordGenerator @apply="applyGeneratedPassword" />
            </div>
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
            <el-select v-model="form.viewers" multiple filterable remote reserve-keyword :remote-method="searchUsers" placeholder="搜索 LDAP 用户" style="width: 100%">
              <el-option v-for="u in ldapUsers" :key="u.sAMAccountName" :label="u.display_name + ' (' + u.sAMAccountName + ')'" :value="u.sAMAccountName" />
            </el-select>
          </el-form-item>
        </el-form>
      </div>
    </div>
    <span slot="footer">
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSave">保存</el-button>
    </span>
  </el-dialog>
</template>

<script>
import IconPicker from './IconPicker.vue'
import PasswordGenerator from './PasswordGenerator.vue'
import { createPasswordEntry, updatePasswordEntry } from '@/api/password_vault'
import { encryptPassword } from '@/utils/rsa'
import request from '@/api/request'

export default {
  name: 'PasswordEntryDialog',
  components: { IconPicker, PasswordGenerator },
  props: {
    categories: { type: Array, default: () => [] }
  },
  data() {
    return {
      visible: false,
      isEdit: false,
      editId: null,
      loading: false,
      showPassword: false,
      showGenerator: false,
      ldapUsers: [],
      form: {
        icon: 'server',
        name: '',
        username: '',
        password: '',
        category_id: null,
        url: '',
        port: null,
        notes: '',
        viewers: []
      },
      rules: {
        name: [{ required: true, message: '请输入名称', trigger: 'blur' }],
        username: [{ required: true, message: '请输入账号', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
        category_id: [{ required: true, message: '请选择分类', trigger: 'change' }]
      }
    }
  },
  methods: {
    open(entry) {
      this.showPassword = false
      this.showGenerator = false
      this.ldapUsers = []
      if (entry) {
        this.isEdit = true
        this.editId = entry.id
        this.form = {
          icon: entry.icon || 'server',
          name: entry.name,
          username: entry.username,
          password: '',
          category_id: entry.category_id,
          url: entry.url || '',
          port: entry.port || null,
          notes: entry.notes || '',
          viewers: entry.viewers ? [...entry.viewers] : []
        }
        // 编辑时密码非必填
        this.rules.password = []
      } else {
        this.isEdit = false
        this.editId = null
        this.form = { icon: 'server', name: '', username: '', password: '', category_id: null, url: '', port: null, notes: '', viewers: [] }
        this.rules.password = [{ required: true, message: '请输入密码', trigger: 'blur' }]
      }
      this.visible = true
      this.$nextTick(() => {
        this.$refs.form?.clearValidate()
        this.loadAllUsers()
      })
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
    applyGeneratedPassword(password) {
      this.form.password = password
      this.showGenerator = false
    },
    handleSave() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        this.loading = true
        try {
          const data = { ...this.form }
          // RSA 加密密码
          if (data.password) {
            data.encrypted_password = await encryptPassword(data.password)
          }
          delete data.password

          if (this.isEdit) {
            await updatePasswordEntry(this.editId, data)
            this.$message.success('更新成功')
          } else {
            await createPasswordEntry(data)
            this.$message.success('创建成功')
          }
          this.visible = false
          this.$emit('saved')
        } catch (err) {
          this.$message.error(err.response?.data?.message || '操作失败')
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
}
.left-panel {
  width: 260px;
  flex-shrink: 0;
  padding-right: 5px;
  border-right: 1px solid #e2e8f0;
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
  margin-top: 12px;
  padding: 16px;
  background: #f1f5f9;
  border-radius: 10px;
  border: 1px solid #e2e8f0;
}
</style>
