<template>
  <el-dialog
    title="双控验证"
    :visible.sync="visible"
    width="420px"
    :close-on-click-modal="false"
    :append-to-body="true"
    @close="handleCancel"
  >
    <div class="dual-control-hint">
      <i class="el-icon-warning-outline"></i>
      此操作需要另一名IT组成员进行双控验证确认
    </div>
    <el-form :model="form" :rules="rules" ref="form" label-width="100px" @submit.native.prevent="handleConfirm">
      <el-form-item label="审批人账号" prop="username">
        <el-input
          ref="usernameInput"
          v-model="form.username"
          placeholder="请输入另一成员的域账号"
          @keyup.enter="$refs.password.focus()"
        />
      </el-form-item>
      <el-form-item label="密码" prop="password">
        <el-input
          ref="password"
          v-model="form.password"
          type="password"
          placeholder="请输入密码"
          @keyup.enter="handleConfirm"
        />
      </el-form-item>
      <!-- 隐藏的提交按钮：使回车键触发表单提交 -->
      <button type="submit" style="display:none"></button>
    </el-form>
    <span slot="footer">
      <el-button @click="handleCancel">取消</el-button>
      <el-button type="danger" :loading="verifying" @click="handleConfirm">
        验证并执行
      </el-button>
    </span>
  </el-dialog>
</template>

<script>
import { verifyDualControl } from '@/api/dualControl'

export default {
  name: 'DualControlDialog',
  data() {
    return {
      visible: false,
      verifying: false,
      form: { username: '', password: '' },
      resolvePromise: null,
      rejectPromise: null,
      rules: {
        username: [{ required: true, message: '请输入审批人账号', trigger: 'blur' }],
        password: [{ required: true, message: '请输入密码', trigger: 'blur' }]
      }
    }
  },
  methods: {
    // 打开双控验证弹窗，返回 Promise<string>（验证成功后resolve token）
    open() {
      this.visible = true
      this.form = { username: '', password: '' }
      this.$nextTick(() => {
        if (this.$refs.form) this.$refs.form.clearValidate()
        if (this.$refs.usernameInput) this.$refs.usernameInput.focus()
      })
      return new Promise((resolve, reject) => {
        this.resolvePromise = resolve
        this.rejectPromise = reject
      })
    },
    async handleConfirm() {
      try {
        await this.$refs.form.validate()
      } catch {
        return
      }
      this.verifying = true
      try {
        const res = await verifyDualControl({
          username: this.form.username.trim(),
          password: this.form.password
        })
        const token = res.data?.token || res.token
        this.$message.success('双控验证通过')
        this.visible = false
        if (this.resolvePromise) {
          this.resolvePromise(token)
          this.resolvePromise = null
          this.rejectPromise = null
        }
      } catch (e) {
        // request.js拦截器已统一显示错误提示，此处无需重复弹出
      } finally {
        this.verifying = false
      }
    },
    handleCancel() {
      this.visible = false
      if (this.rejectPromise) {
        this.rejectPromise(new Error('canceled'))
        this.resolvePromise = null
        this.rejectPromise = null
      }
    }
  }
}
</script>

<style scoped>
.dual-control-hint {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-bottom: 16px;
  padding: 10px 14px;
  background: #fef0f0;
  border: 1px solid #fde2e2;
  border-radius: 4px;
  color: #f56c6c;
  font-size: 13px;
}
.dual-control-hint i {
  font-size: 18px;
}
</style>
