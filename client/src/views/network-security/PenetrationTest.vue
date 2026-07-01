<template>
  <div class="penetration-test">
    <el-card>
      <!-- Tabs -->
      <el-tabs v-model="activeTab" @tab-click="handleTabChange" style="margin-bottom: 16px">
        <el-tab-pane label="内部渗透测试" name="internal" />
        <el-tab-pane label="外部渗透测试" name="external" />
      </el-tabs>
      <div slot="header" class="page-header">
        <span>渗透测试</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-upload2" @click="openUpload">上传报告</el-button>
          <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="全部年份" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="y + '年'" :value="y" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索文件名/日期/描述..." size="small" clearable @keyup.enter.native="handleFilterChange" @clear="handleFilterChange" style="width: 220px" />
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleFilterChange">搜索</el-button>
      </div>

      <!-- 数据表格 -->
      <el-table :data="records" border stripe v-loading="loading" style="margin-top: 12px">
        <el-table-column type="index" label="#" width="50" align="center" />
        <el-table-column prop="year" label="年份" width="70" align="center" />
        <el-table-column prop="report_date" label="报告日期" width="110" align="center" />
        <el-table-column label="可渗透漏洞数" width="110" align="center">
          <template slot-scope="{ row }">
            <span style="color: #F56C6C; font-weight: bold">{{ row.vuln_count }}</span>
          </template>
        </el-table-column>
        <el-table-column prop="description" label="结果描述" min-width="200" show-overflow-tooltip />
        <el-table-column label="关联漏洞扫描报告" width="320" align="center">
          <template slot-scope="{ row }">
            <template v-if="row.vulnerability_scans && row.vulnerability_scans.length > 0">
              <el-tooltip v-for="vs in row.vulnerability_scans" :key="vs.id" placement="top" effect="dark">
                <div slot="content">{{ formatVulnScanTooltip(vs) }}</div>
                <el-tag size="mini" style="margin: 2px">
                  {{ vs.year }}-Q{{ vs.quarter }}-{{ vs.scan_type === 'internal' ? '内部' : '外部' }}
                </el-tag>
              </el-tooltip>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="180" fixed="right">
          <template slot-scope="{ row }">
            <div class="op-btns">
              <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreview(row)">预览</el-button>
              <el-button size="mini" type="text" icon="el-icon-edit" @click="handleEdit(row)">编辑</el-button>
              <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="handleDelete(row)">删除</el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        style="margin-top: 16px; text-align: right"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size.sync="pageSize"
        :current-page.sync="page"
        :page-sizes="[10, 20, 50]"
        @size-change="handleSizeChange"
        @current-change="fetchData"
      />
    </el-card>

    <!-- 上传/编辑弹窗 -->
    <el-dialog :title="isEdit ? '编辑渗透测试报告' : '上传渗透测试报告'" :visible.sync="showUpload" width="680px" :close-on-click-modal="false">
      <el-form :model="uploadForm" ref="uploadFormRef" :rules="uploadRules" label-width="120px">
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="年份" prop="year">
              <el-input-number v-model="uploadForm.year" :min="2020" :max="2100" :step="1" controls-position="right" style="width: 100%" />
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="报告日期" prop="report_date">
              <el-date-picker v-model="uploadForm.report_date" type="date" value-format="yyyy-MM-dd" placeholder="选择日期" style="width: 100%" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="可渗透漏洞数">
          <el-input-number v-model="uploadForm.vuln_count" :min="0" controls-position="right" style="width: 100%" />
        </el-form-item>
        <el-form-item label="结果描述">
          <el-input v-model="uploadForm.description" type="textarea" :rows="3" placeholder="描述渗透测试结果" />
        </el-form-item>
        <el-form-item label="关联漏洞扫描报告">
          <el-select v-model="uploadForm.vulnerability_scan_ids" multiple collapse-tags placeholder="选择关联的漏洞扫描报告" style="width: 100%" :popper-append-to-body="false">
            <el-option v-for="vs in vulnScanOptions" :key="vs.id" :label="formatVulnScanTag(vs)" :value="vs.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="文件" v-if="!isEdit">
          <el-upload
            ref="uploader"
            action=""
            :auto-upload="false"
            :limit="1"
            accept=".pdf,.docx"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            drag
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
            <div slot="tip" class="el-upload__tip">支持 PDF / DOCX 格式文件</div>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可更换文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showUpload = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">{{ isEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- 预览弹窗 -->
    <el-dialog title="文件预览" :visible.sync="previewVisible" width="80%" top="3vh" :close-on-click-modal="true">
      <iframe v-if="previewUrl && isPdf" :src="previewUrl" style="width: 100%; height: 70vh; border: none;" />
      <div v-else-if="!isPdf" style="text-align: center; padding: 40px;">
        <i class="el-icon-document" style="font-size: 64px; color: #409EFF;"></i>
        <p style="margin-top: 16px; font-size: 16px;">DOCX 文件无法直接预览，请下载后查看</p>
      </div>
      <span slot="footer">
        <el-button type="primary" size="small" icon="el-icon-download" @click="handleDownloadFromPreview">下载</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getPenetrationTests, createPenetrationTest, updatePenetrationTest, deletePenetrationTest,
  getPenetrationTestPreviewUrl, getPenetrationTestDownloadUrl
} from '@/api/penetration_test'
import { getVulnerabilityScans } from '@/api/vulnerability_scan'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'PenetrationTest',
  components: { DualControlDialog },
  data() {
    const now = new Date()
    return {
      activeTab: 'internal',
      records: [],
      loading: false,
      page: 1,
      pageSize: 10,
      total: 0,
      filterYear: '',
      keyword: '',
      yearOptions: Array.from({ length: 10 }, (_, i) => now.getFullYear() - i),
      // 上传/编辑
      showUpload: false,
      isEdit: false,
      editingId: null,
      uploading: false,
      uploadForm: {
        test_type: 'internal',
        year: now.getFullYear(),
        report_date: '',
        vuln_count: 0,
        description: '',
        vulnerability_scan_ids: []
      },
      uploadRules: {
        year: [{ required: true, message: '请选择年份', trigger: 'change' }]
      },
      selectedFile: null,
      fileList: [],
      vulnScanOptions: [],
      // 预览
      previewVisible: false,
      previewUrl: '',
      previewDownloadUrl: '',
      previewFileName: '',
      isPdf: true
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const params = { page: this.page, page_size: this.pageSize, test_type: this.activeTab }
        if (this.filterYear) params.year = this.filterYear
        if (this.keyword) params.keyword = this.keyword
        const res = await getPenetrationTests(params)
        this.records = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    async fetchVulnScans() {
      try {
        const res = await getVulnerabilityScans({ page: 1, page_size: 200, scan_type: this.activeTab })
        this.vulnScanOptions = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    formatVulnScanLabel(vs) {
      const base = `${vs.year}-Q${vs.quarter}-${vs.scan_type === 'internal' ? '内部' : '外部'} (${vs.report_date || '无日期'})`
      if (vs.scan_type === 'internal') {
        const regions = (vs.regions || []).map(r => r.name).join(', ')
        return regions ? `${base} | ${regions}` : `${base} | 无区域`
      } else {
        return `${base} | ${vs.external_ip || '无IP'}`
      }
    },
    formatVulnScanTag(vs) {
      // 简化显示：只显示年份-季度-类型
      return `${vs.year}-Q${vs.quarter}-${vs.scan_type === 'internal' ? '内部' : '外部'}`
    },
    formatVulnScanTooltip(vs) {
      const lines = [
        `年份: ${vs.year}`,
        `季度: Q${vs.quarter}`,
        `类型: ${vs.scan_type === 'internal' ? '内部' : '外部'}`,
        `报告日期: ${vs.report_date || '无'}`
      ]
      if (vs.scan_type === 'internal') {
        const regions = (vs.regions || []).map(r => r.name).join(', ')
        lines.push(`扫描区域: ${regions || '无'}`)
      } else {
        lines.push(`对外IP: ${vs.external_ip || '无'}`)
      }
      return lines.join('\n')
    },
    handleSizeChange() {
      this.page = 1
      this.fetchData()
    },
    handleFilterChange() {
      this.page = 1
      this.fetchData()
    },
    handleTabChange() {
      this.page = 1
      this.filterYear = ''
      this.keyword = ''
      this.fetchData()
    },
    // 上传/编辑
    openUpload() {
      this.resetUploadForm()
      this.uploadForm.test_type = this.activeTab
      this.showUpload = true
      this.fetchVulnScans()
    },
    resetUploadForm() {
      const now = new Date()
      this.isEdit = false
      this.editingId = null
      this.selectedFile = null
      this.fileList = []
      this.uploadForm = {
        test_type: this.activeTab,
        year: now.getFullYear(),
        report_date: '',
        vuln_count: 0,
        description: '',
        vulnerability_scan_ids: []
      }
    },
    handleFileChange(file) {
      this.selectedFile = file.raw
    },
    handleFileRemove() {
      this.selectedFile = null
    },
    handleEdit(row) {
      this.isEdit = true
      this.editingId = row.id
      this.uploadForm = {
        test_type: row.test_type || 'internal',
        year: row.year,
        report_date: row.report_date || '',
        vuln_count: row.vuln_count || 0,
        description: row.description || '',
        vulnerability_scan_ids: (row.vulnerability_scans || []).map(vs => vs.id)
      }
      this.showUpload = true
      this.fetchVulnScans()
    },
    async handleUpload() {
      this.$refs.uploadFormRef.validate(async valid => {
        if (!valid) return
        if (!this.isEdit && !this.selectedFile) {
          this.$message.warning('请选择文件')
          return
        }
        this.uploading = true
        try {
          const formData = new FormData()
          formData.append('test_type', this.uploadForm.test_type)
          formData.append('year', this.uploadForm.year)
          formData.append('report_date', this.uploadForm.report_date || '')
          formData.append('vuln_count', this.uploadForm.vuln_count)
          formData.append('description', this.uploadForm.description || '')
          formData.append('vulnerability_scan_ids', this.uploadForm.vulnerability_scan_ids.join(','))
          if (!this.isEdit && this.selectedFile) {
            formData.append('file', this.selectedFile)
          }

          if (this.isEdit) {
            await updatePenetrationTest(this.editingId, formData, await this.$refs.dualControl.open())
            this.$message.success('更新成功')
          } else {
            await createPenetrationTest(formData, await this.$refs.dualControl.open())
            this.$message.success('上传成功')
          }
          this.showUpload = false
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') {
            console.error(e)
            this.$message.error(this.isEdit ? '更新失败' : '上传失败')
          }
        } finally {
          this.uploading = false
        }
      })
    },
    // 预览
    handlePreview(row) {
      this.previewUrl = getPenetrationTestPreviewUrl(row.id)
      this.previewDownloadUrl = getPenetrationTestDownloadUrl(row.id)
      this.previewFileName = row.file_name
      this.isPdf = row.file_name && row.file_name.toLowerCase().endsWith('.pdf')
      this.previewVisible = true
    },
    handleDownloadFromPreview() {
      if (this.previewDownloadUrl) {
        window.open(this.previewDownloadUrl, '_blank')
      }
    },
    // 删除
    async handleDelete(row) {
      try {
        await this.$confirm('确定要删除该渗透测试报告吗？此操作不可恢复。', '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deletePenetrationTest(row.id, dualToken)
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
.penetration-test {
  padding: 0;
}
.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.page-header-right {
  display: flex;
  gap: 8px;
}
.filter-bar {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}
.op-btns {
  display: flex;
  gap: 4px;
  justify-content: center;
}
</style>
