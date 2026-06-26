<template>
  <div class="region-list">
    <el-card>
      <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
        <span>区域列表</span>
        <el-button type="primary" size="small" icon="el-icon-plus" @click="handleAdd">新增区域</el-button>
      </div>
      <el-table :data="regions" border stripe>
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="name" label="区域名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="操作" width="200">
          <template slot-scope="scope">
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-dialog :title="dialogTitle" :visible.sync="dialogVisible" width="500px">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="区域名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入区域名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import { getRegions, createRegion, updateRegion, deleteRegion } from '@/api/region'

export default {
  name: 'RegionList',
  data() {
    return {
      regions: [],
      dialogVisible: false,
      dialogTitle: '新增区域',
      form: { name: '', description: '' },
      rules: {
        name: [{ required: true, message: '请输入区域名称', trigger: 'blur' }]
      }
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      try {
        const res = await getRegions()
        this.regions = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    handleAdd() {
      this.dialogTitle = '新增区域'
      this.form = { name: '', description: '' }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogTitle = '编辑区域'
      this.form = { name: row.name, description: row.description }
      this.form.id = row.id
      this.dialogVisible = true
    },
    handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        try {
          if (this.form.id) {
            await updateRegion(this.form.id, { name: this.form.name, description: this.form.description })
            this.$message.success('更新成功')
          } else {
            await createRegion({ name: this.form.name, description: this.form.description })
            this.$message.success('创建成功')
          }
          this.dialogVisible = false
          this.fetchData()
        } catch (e) {
          console.error(e)
        }
      })
    },
    handleDelete(row) {
      this.$confirm('确定要删除该区域吗？', '提示', { type: 'warning' }).then(async () => {
        try {
          await deleteRegion(row.id)
          this.$message.success('删除成功')
          this.fetchData()
        } catch (e) {
          console.error(e)
        }
      }).catch(() => {})
    }
  }
}
</script>
