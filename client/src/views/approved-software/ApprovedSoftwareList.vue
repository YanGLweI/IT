<template>
  <div class="approved-software-list">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">核准软件目录</h2>
        <p class="page-subtitle">管理组织允许使用的软件清单</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-plus" @click="handleAdd">新增软件</el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-input v-model="keyword" placeholder="搜索软件名称..." size="small" clearable @keyup.enter.native="handleFilterChange" @clear="handleFilterChange" style="width: 200px" />
      <el-select v-model="filterLicense" placeholder="全部授权类型" size="small" clearable @change="handleFilterChange" style="width: 140px">
        <el-option label="商用" value="商用" />
        <el-option label="开源" value="开源" />
      </el-select>
      <el-select v-model="filterNeedUpdate" placeholder="是否更新" size="small" clearable @change="handleFilterChange" style="width: 120px">
        <el-option label="需要更新" value="true" />
        <el-option label="无需更新" value="false" />
      </el-select>
      <el-button size="small" type="primary" icon="el-icon-search" @click="handleFilterChange">搜索</el-button>
    </div>

    <!-- 数据表格 -->
    <div class="table-card" ref="tableCard">
    <div class="table-wrapper">
      <el-table :data="list" stripe v-loading="loading" :max-height="tableMaxHeight">
        <el-table-column type="index" label="序号" width="75" align="center" />
        <el-table-column prop="name" label="软件名称" min-width="150" show-overflow-tooltip />
        <el-table-column label="厂商官网" min-width="180" show-overflow-tooltip>
          <template slot-scope="{ row }">
            <a v-if="row.vendor_website" :href="row.vendor_website" target="_blank" class="link">{{ row.vendor_website }}</a>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="version" label="软件版本" width="180" show-overflow-tooltip />
        <el-table-column prop="latest_version" label="最新版本" width="180" show-overflow-tooltip />
        <el-table-column label="是否更新" width="100" align="center">
          <template slot-scope="scope">
            <el-tag :type="scope.row.need_update ? 'danger' : 'success'" size="small">
              {{ scope.row.need_update ? '是' : '否' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right" align="center">
          <template slot-scope="scope">
            <div class="op-btns">
              <el-button size="mini" type="info" @click="handleDetail(scope.row)">详情</el-button>
              <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
              <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size.sync="pageSize"
        :current-page.sync="currentPage"
        :page-sizes="[10, 20, 50]"
        @size-change="handleSizeChange"
        @current-change="fetchData"
      />
    </div>

    <!-- 新增/编辑弹窗 -->
    <el-dialog class="vault-dialog" :title="dialogTitle" :visible.sync="dialogVisible" width="680px" :close-on-click-modal="false">
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
      <span slot="footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">确定</el-button>
      </span>
    </el-dialog>

    <!-- 详情弹窗 -->
    <el-dialog class="vault-dialog" title="软件详情" :visible.sync="detailVisible" width="650px">
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
      <span slot="footer">
        <el-button @click="detailVisible = false">关闭</el-button>
      </span>
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
import tableHeightMixin from '@/mixins/table-height'

export default {
  name: 'ApprovedSoftwareList',
  components: { DualControlDialog },
  mixins: [tableHeightMixin],
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      pageSize: 10,
      currentPage: 1,
      keyword: '',
      filterLicense: '',
      filterNeedUpdate: '',
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
        const params = { page: this.currentPage, page_size: this.pageSize }
        if (this.keyword) params.keyword = this.keyword
        if (this.filterLicense) params.license_type = this.filterLicense
        if (this.filterNeedUpdate) params.need_update = this.filterNeedUpdate
        const res = await getApprovedSoftware(params)
        this.list = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
        this.$nextTick(() => this.calcTableHeight())
      }
    },
    handleFilterChange() {
      this.currentPage = 1
      this.fetchData()
    },
    handleSizeChange() {
      this.currentPage = 1
      this.fetchData()
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
.approved-software-list {
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

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.header-left {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
  margin: 0;
}

.header-actions {
  display: flex;
  gap: 10px;
  align-items: center;
}

.op-btns {
  display: flex;
  gap: 6px;
  flex-wrap: nowrap;
}

.link {
  color: #3b82f6;
  text-decoration: none;
}
.link:hover {
  text-decoration: underline;
}

/* 按钮样式 */
.header-actions .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  color: #fff;
}
.header-actions .el-button--primary:hover {
  background: #2563eb;
  color: #fff;
}

/* 筛选栏搜索按钮 */
.filter-bar .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  color: #fff;
}
.filter-bar .el-button--primary:hover {
  background: #2563eb;
  color: #fff;
}

/* 弹窗底部按钮 */
.el-dialog__footer .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  color: #fff;
}
.el-dialog__footer .el-button--primary:hover {
  background: #2563eb;
  color: #fff;
}
</style>

<style>
.detail-label {
  width: 100px !important;
  font-weight: 600;
  color: #64748b;
  background: #f8fafc;
}
.detail-content {
  color: #1e293b;
}
</style>
