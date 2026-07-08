<template>
  <div class="approved-software-list">
    <el-card>
      <template #header><div style="display: flex; justify-content: space-between; align-items: center">
        <span>核准软件目录</span>
        <el-button type="primary" size="small" :icon="Plus" @click="handleAdd">新增软件</el-button>
      </div>
      </template>
      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column type="index" label="序号" width="60" align="center" />
        <el-table-column prop="name" label="软件名称" min-width="150" show-overflow-tooltip />
        <el-table-column label="厂商官网" min-width="180" show-overflow-tooltip>
          <template v-slot="{ row }">
            <a v-if="row.vendor_website" :href="row.vendor_website" target="_blank" class="link">{{ row.vendor_website }}</a>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="软件版本" width="150" show-overflow-tooltip />
        <el-table-column prop="latest_version" label="最新版本" width="150" show-overflow-tooltip />
        <el-table-column label="是否更新" width="90" align="center">
          <template v-slot="scope">
            <el-tag :type="scope.row.need_update ? 'danger' : 'success'" size="small">
              {{ scope.row.need_update ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right" align="center">
          <template v-slot="scope">
            <div class="op-btns">
              <el-button size="small" type="info" @click="handleDetail(scope.row)">详情</el-button>
              <el-button size="small" @click="handleEdit(scope.row)">编辑</el-button>
              <el-button size="small" type="danger" @click="handleDelete(scope.row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- 新增/编辑弹窗 -->
    <el-dialog :title="dialogTitle" v-model="dialogVisible" width="680px" :close-on-click-modal="false">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="软件名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入软件名称" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="软件版本" prop="version">
              <el-input v-model="form.version" placeholder="如 1.0.0" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="最新版本" prop="latest_version">
              <el-input v-model="form.latest_version" placeholder="如 2.0.0" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="是否更新" prop="need_update">
          <el-radio-group v-model="form.need_update">
            <el-radio :label="true">是</el-radio>
            <el-radio :label="false">否</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.need_update" label="更新原因" prop="update_reason">
          <el-input v-model="form.update_reason" type="textarea" :rows="2" placeholder="请输入需要更新的原因" />
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="厂商" prop="vendor">
              <el-input v-model="form.vendor" placeholder="请输入厂商名称" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="厂商官网" prop="vendor_website">
              <el-input v-model="form.vendor_website" placeholder="https://example.com" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="授权类型" prop="license_type">
          <el-radio-group v-model="form.license_type">
            <el-radio label="商用">商用</el-radio>
            <el-radio label="开源">开源</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="用途" prop="purpose">
          <el-input v-model="form.purpose" type="textarea" :rows="2" placeholder="请输入软件用途说明" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog title="软件详情" v-model="detailVisible" width="650px">
      <el-descriptions :column="2" border v-if="detailRow" label-class-name="detail-label" content-class-name="detail-content">
        <el-descriptions-item label="软件名称" :span="2">{{ detailRow.name }}</el-descriptions-item>
        <el-descriptions-item label="软件版本">{{ detailRow.version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="最新版本">{{ detailRow.latest_version || '-' }}</el-descriptions-item>
        <el-descriptions-item label="是否更新" :span="2">
          <el-tag :type="detailRow.need_update ? 'danger' : 'success'" size="small">
            {{ detailRow.need_update ? '是' : '否' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="更新原因" :span="2">{{ detailRow.update_reason || '-' }}</el-descriptions-item>
        <el-descriptions-item label="厂商" :span="2">{{ detailRow.vendor || '-' }}</el-descriptions-item>
        <el-descriptions-item label="厂商官网" :span="2">
          <a v-if="detailRow.vendor_website" :href="detailRow.vendor_website" target="_blank" class="link">{{ detailRow.vendor_website }}</a>
          <span v-else>-</span>
        </el-descriptions-item>
        <el-descriptions-item label="授权类型" :span="2">
          <el-tag :type="detailRow.license_type === '开源' ? 'success' : 'warning'" size="small">
            {{ detailRow.license_type || '商用' }}
          </el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="用途" :span="2">{{ detailRow.purpose || '-' }}</el-descriptions-item>
        <el-descriptions-item label="创建时间" :span="2">{{ formatTime(detailRow.created_at) }}</el-descriptions-item>
        <el-descriptions-item label="修改时间" :span="2">{{ formatTime(detailRow.updated_at) }}</el-descriptions-item>
      </el-descriptions>
      <template #footer>
        <el-button @click="detailVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getApprovedSoftware,
  createApprovedSoftware,
  updateApprovedSoftware,
  deleteApprovedSoftware
} from '@/api/approved_software'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'ApprovedSoftwareList',
  components: { DualControlDialog },
  data() {
    return {
      list: [],
      loading: false,
      detailVisible: false,
      detailRow: null,
      dialogVisible: false,
      dialogTitle: '新增核准软件',
      submitting: false,
      form: this.getDefaultForm(),
      rules: {
        name: [{ required: true, message: '请输入软件名称', trigger: 'blur' }],
        license_type: [{ required: true, message: '请选择授权类型', trigger: 'change' }]
      }
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    getDefaultForm() {
      return {
        name: '',
        version: '',
        latest_version: '',
        need_update: false,
        update_reason: '',
        vendor: '',
        vendor_website: '',
        license_type: '商用',
        purpose: ''
      }
    },
    formatTime(t) {
      if (!t) return '-'
      return t.replace('T', ' ').substring(0, 19)
    },
    async fetchData() {
      this.loading = true
      try {
        const res = await getApprovedSoftware()
        this.list = res.data || []
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    handleDetail(row) {
      this.detailRow = row
      this.detailVisible = true
    },
    handleAdd() {
      this.dialogTitle = '新增核准软件'
      this.form = this.getDefaultForm()
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogTitle = '编辑核准软件'
      this.form = {
        id: row.id,
        name: row.name,
        version: row.version || '',
        latest_version: row.latest_version || '',
        need_update: row.need_update || false,
        update_reason: row.update_reason || '',
        vendor: row.vendor || '',
        vendor_website: row.vendor_website || '',
        license_type: row.license_type || '商用',
        purpose: row.purpose || ''
      }
      this.dialogVisible = true
    },
    handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        this.submitting = true
        try {
          const dualToken = await this.$refs.dualControl.open()
          if (this.form.id) {
            await updateApprovedSoftware(this.form.id, this.form, dualToken)
            this.$message.success('更新成功')
          } else {
            await createApprovedSoftware(this.form, dualToken)
            this.$message.success('创建成功')
          }
          this.dialogVisible = false
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        } finally {
          this.submitting = false
        }
      })
    },
    async handleDelete(row) {
      try {
        await this.$confirm('确定要删除该核准软件吗？', '提示', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteApprovedSoftware(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    }
  }
}
</script>

<style scoped>
.op-btns {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
}
.link {
  color: #409EFF;
  text-decoration: none;
}
.link:hover {
  text-decoration: underline;
}
</style>

<style>
.detail-label {
  width: 100px !important;
  font-weight: 600;
  color: #606266;
  background: #f5f7fa;
}
.detail-content {
  color: #303133;
}
</style>
