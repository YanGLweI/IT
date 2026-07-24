<template>
  <div class="ostype-list">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">操作系统管理</h2>
        <p class="page-subtitle">管理资产支持的操作系统类型</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-plus" @click="handleAdd">新增类型</el-button>
      </div>
    </div>

    <div class="table-card" ref="tableCard">
      <div class="table-wrapper">
      <el-table :data="osTypes" stripe :max-height="tableMaxHeight">
        <el-table-column type="index" label="#" width="70" align="center" />
        <el-table-column prop="name" label="操作系统类型" />
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      </div>
    </div>

    <el-dialog class="vault-dialog" :title="dialogTitle" :visible.sync="dialogVisible" width="500px">
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
import tableHeightMixin from '@/mixins/table-height'

export default {
  name: 'OSTypeList',
  components: { DualControlDialog },
  mixins: [tableHeightMixin],
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
      } finally {
        this.$nextTick(() => this.calcTableHeight())
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

<style scoped>
.ostype-list {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

/* --- 页面头部 --- */
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.page-title {
  font-size: 20px;
  font-weight: 600;
  color: #1e293b;
  margin: 0;
}
.page-subtitle {
  font-size: 13px;
  color: #64748b;
  margin: 4px 0 0;
}
.header-actions {
  display: flex;
  align-items: center;
  gap: 10px;
}

.table-card {
}

.table-wrapper {
}

/* --- 主按钮 --- */
.header-actions .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 500;
}
.header-actions .el-button--primary:hover {
  background: #2563eb;
}
</style>
