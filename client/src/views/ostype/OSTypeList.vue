<template>
  <div class="ostype-list">
    <el-card>
      <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
        <span>操作系统类型管理</span>
        <el-button type="primary" size="small" icon="el-icon-plus" @click="handleAdd">新增类型</el-button>
      </div>
      <el-table :data="osTypes" border stripe>
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="name" label="操作系统类型" />
        <el-table-column label="操作" width="200" align="center">
          <template slot-scope="scope">
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="类型名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入操作系统类型名称，如 Windows 11" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getOSTypes, createOSType, updateOSType, deleteOSType } from '@/api/os_type'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'OSTypeList',
  components: { DualControlDialog },
  data() {
    return {
      osTypes: [],
      dialogVisible: false,
      dialogTitle: '新增操作系统类型',
      form: { name: '' },
      rules: {
        name: [{ required: true, message: '请输入类型名称', trigger: 'blur' }]
      }
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      try {
        const res = await getOSTypes()
        this.osTypes = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    handleAdd() {
      this.dialogTitle = '新增操作系统类型'
      this.form = { name: '' }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogTitle = '编辑操作系统类型'
      this.form = { id: row.id, name: row.name }
      this.dialogVisible = true
    },
    handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        try {
          const dualToken = await this.$refs.dualControl.open()
          if (this.form.id) {
            await updateOSType(this.form.id, { name: this.form.name }, dualToken)
            this.$message.success('更新成功')
          } else {
            await createOSType({ name: this.form.name }, dualToken)
            this.$message.success('创建成功')
          }
          this.dialogVisible = false
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        }
      })
    },
    handleDelete(row) {
      this.$confirm('确定要删除该操作系统类型吗？', '提示', { type: 'warning' }).then(async () => {
        try {
          const dualToken = await this.$refs.dualControl.open()
          await deleteOSType(row.id, dualToken)
          this.$message.success('删除成功')
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        }
      }).catch(() => {})
    }
  }
}
</script>
