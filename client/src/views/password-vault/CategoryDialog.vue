<template>
  <el-dialog :visible.sync="visible" :title="isEdit ? '编辑分类' : '添加分类'" width="400px" :close-on-click-modal="false">
    <el-form :model="form" :rules="rules" ref="form" label-width="80px">
      <el-form-item label="分类名称" prop="name">
        <el-input v-model="form.name" placeholder="请输入分类名称" maxlength="50" />
      </el-form-item>
      <el-form-item label="图标" prop="icon">
        <IconPicker v-model="form.icon" />
      </el-form-item>
    </el-form>
    <span slot="footer">
      <el-button @click="visible = false">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSave">保存</el-button>
    </span>
  </el-dialog>
</template>

<script>
import IconPicker from './IconPicker.vue'
import { createPasswordCategory, updatePasswordCategory } from '@/api/password_vault'

export default {
  name: 'CategoryDialog',
  components: { IconPicker },
  props: {
    categories: { type: Array, default: () => [] }
  },
  data() {
    return {
      visible: false,
      isEdit: false,
      editId: null,
      loading: false,
      form: { name: '', icon: 'server' },
      rules: {
        name: [{ required: true, message: '请输入分类名称', trigger: 'blur' }],
        icon: [{ required: true, message: '请选择图标', trigger: 'change' }]
      }
    }
  },
  methods: {
    open(category) {
      if (category) {
        this.isEdit = true
        this.editId = category.id
        this.form = { name: category.name, icon: category.icon }
      } else {
        this.isEdit = false
        this.editId = null
        this.form = { name: '', icon: 'server' }
      }
      this.visible = true
      this.$nextTick(() => this.$refs.form?.clearValidate())
    },
    handleSave() {
      this.$refs.form.validate(async (valid) => {
        if (!valid) return
        this.loading = true
        try {
          if (this.isEdit) {
            await updatePasswordCategory(this.editId, this.form)
            this.$message.success('更新成功')
          } else {
            await createPasswordCategory(this.form)
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
