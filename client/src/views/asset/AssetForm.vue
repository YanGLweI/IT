<template>
  <el-dialog :title="isEdit ? '编辑资产' : '新增资产'" :visible.sync="dialogVisible" width="600px" @close="handleClose">
    <DualControlDialog ref="dualControl" />
    <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
      <el-form-item label="计算机名" prop="computer_name">
        <el-input v-model="form.computer_name" placeholder="请输入计算机名" />
      </el-form-item>
      <el-form-item label="所属区域" prop="region_id">
        <el-select v-model="form.region_id" placeholder="请选择区域" style="width: 100%">
          <el-option v-for="r in regions" :key="r.id" :label="r.name" :value="r.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="IP地址">
        <el-input v-model="form.ip_address" placeholder="请输入IP地址" />
      </el-form-item>
      <el-form-item label="操作系统" prop="os_type_id">
        <el-select v-model="form.os_type_id" placeholder="请选择操作系统" filterable style="width: 100%">
          <el-option v-for="os in osTypes" :key="os.id" :label="os.name" :value="os.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="用途">
        <el-input v-model="form.purpose" placeholder="请输入用途" />
      </el-form-item>
      <el-form-item label="资产等级">
        <el-select v-model="form.asset_level" placeholder="请选择资产等级" style="width: 100%">
          <el-option label="一级" value="一级" />
          <el-option label="二级" value="二级" />
          <el-option label="三级" value="三级" />
        </el-select>
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="form.status" style="width: 100%">
          <el-option label="在用" value="在用" />
          <el-option label="闲置" value="闲置" />
          <el-option label="报废" value="报废" />
        </el-select>
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.remark" type="textarea" :rows="3" placeholder="请输入备注" />
      </el-form-item>
    </el-form>
    <span slot="footer">
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="handleSubmit">确定</el-button>
    </span>
  </el-dialog>
</template>

<script>
import { createAsset, updateAsset } from '@/api/asset'
import DualControlDialog from '@/components/DualControlDialog.vue'
import { getOSTypes } from '@/api/os_type'

export default {
  components: { DualControlDialog },
  name: 'AssetForm',
  props: {
    visible: { type: Boolean, default: false },
    editData: { type: Object, default: null },
    regions: { type: Array, default: () => [] }
  },
  data() {
    return {
      form: this.getDefaultForm(),
      osTypes: [],
      rules: {
        computer_name: [{ required: true, message: '请输入计算机名', trigger: 'blur' }],
        region_id: [{ required: true, message: '请选择区域', trigger: 'change' }],
        os_type_id: [{ required: true, message: '请选择操作系统', trigger: 'change' }]
      }
    }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(val) { this.$emit('update:visible', val) }
    },
    isEdit() {
      return this.editData !== null
    }
  },
  watch: {
    visible(val) {
      if (val) {
        this.fetchOSTypes()
        if (this.editData) {
          // 只提取需要的字段，避免嵌套对象干扰
          this.form = {
            id: this.editData.id,
            computer_name: this.editData.computer_name || '',
            region_id: this.editData.region_id || null,
            ip_address: this.editData.ip_address || '',
            os_type_id: this.editData.os_type_id || null,
            purpose: this.editData.purpose || '',
            asset_level: this.editData.asset_level || '',
            status: this.editData.status || '在用',
            remark: this.editData.remark || ''
          }
        } else {
          this.form = this.getDefaultForm()
        }
      }
    }
  },
  methods: {
    async fetchOSTypes() {
      try {
        const res = await getOSTypes()
        this.osTypes = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    getDefaultForm() {
      return {
        computer_name: '',
        region_id: null,
        ip_address: '',
        os_type_id: null,
        purpose: '',
        asset_level: '',
        status: '在用',
        remark: ''
      }
    },
    handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) {
          console.log('表单验证失败', this.form)
          return
        }
        try {
          const submitData = {
            computer_name: this.form.computer_name,
            region_id: this.form.region_id,
            ip_address: this.form.ip_address,
            os_type_id: this.form.os_type_id,
            purpose: this.form.purpose,
            asset_level: this.form.asset_level,
            status: this.form.status,
            remark: this.form.remark
          }
          
          const dualToken = await this.$refs.dualControl.open()

          if (this.isEdit) {
            await updateAsset(this.form.id, submitData, dualToken)
            this.$message.success('更新成功')
          } else {
            await createAsset(submitData, dualToken)
            this.$message.success('创建成功')
          }
          this.dialogVisible = false
          this.$emit('success')
        } catch (e) {
          if (e.message !== 'canceled') {
            console.error('提交失败:', e)
            this.$message.error(e.response?.data?.message || '操作失败')
          }
        }
      })
    },
    handleClose() {
      this.form = this.getDefaultForm()
    }
  }
}
</script>
