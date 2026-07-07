<template>
  <div class="backup-management">
    <el-card>
      <div slot="header" class="page-header">
        <span>备份管理</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-plus" @click="openCreate">新增备份记录</el-button>
          <el-button type="default" size="small" icon="el-icon-refresh" @click="fetchData" :loading="loading">刷新</el-button>
        </div>
      </div>

      <!-- 筛选栏 -->
      <div class="filter-bar">
        <el-select v-model="filterYear" placeholder="全部年份" size="small" clearable @change="handleFilterChange" style="width: 120px">
          <el-option v-for="y in yearOptions" :key="y" :label="y + '年'" :value="y" />
        </el-select>
        <el-select v-model="filterDept" placeholder="全部部门" size="small" clearable @change="handleFilterChange" style="width: 150px">
          <el-option v-for="d in departmentOptions" :key="d.id" :label="d.name" :value="d.id" />
        </el-select>
        <el-select v-model="filterTool" placeholder="全部备份工具" size="small" clearable @change="handleFilterChange" style="width: 150px">
          <el-option v-for="t in backupToolOptions" :key="t" :label="t" :value="t" />
        </el-select>
        <el-input v-model="keyword" placeholder="搜索备份对象/工具/资产名" size="small" clearable style="width: 220px" @clear="handleFilterChange" @keyup.enter.native="handleFilterChange" />
        <el-button size="small" type="primary" icon="el-icon-search" @click="handleFilterChange">搜索</el-button>
      </div>

      <!-- 数据表格 -->
      <el-table :data="records" border stripe v-loading="loading" style="margin-top: 12px">
        <el-table-column type="expand">
          <template slot-scope="{ row }">
            <div style="padding: 8px 20px">
              <div style="margin-bottom: 8px">
                <el-button size="mini" type="primary" icon="el-icon-plus" @click="openRecoveryForm(row)">添加恢复记录</el-button>
              </div>
              <el-table :data="row.recoveries || []" border size="small">
                <el-table-column type="index" label="#" width="50" align="center" />
                <el-table-column prop="recovery_type" label="恢复类型" width="100" align="center" />
                <el-table-column label="恢复结果" width="80" align="center">
                  <template slot-scope="{ row: r }">
                    <el-tag :type="r.recovery_result === '成功' ? 'success' : 'danger'" size="small">{{ r.recovery_result }}</el-tag>
                  </template>
                </el-table-column>
                <el-table-column prop="recovery_date" label="恢复日期" width="110" align="center" />
                <el-table-column prop="file_name" label="上传文件" show-overflow-tooltip />
                <el-table-column label="操作" width="320" align="center">
                  <template slot-scope="{ row: r }">
                    <div class="op-btns">
                      <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreviewRecovery(r)">预览</el-button>
                      <el-button size="mini" type="text" icon="el-icon-download" @click="handleDownloadRecovery(r)">下载</el-button>
                      <el-button size="mini" type="text" icon="el-icon-edit" @click="openRecoveryEdit(r, row)">编辑</el-button>
                      <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="handleDeleteRecovery(r)">删除</el-button>
                    </div>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </template>
        </el-table-column>
        <el-table-column type="index" label="#" width="50" align="center" />
        <el-table-column prop="application_date" label="申请日期" width="110" align="center" />
        <el-table-column label="备份源" width="150" show-overflow-tooltip>
          <template slot-scope="{ row }">{{ row.backup_source_asset ? row.backup_source_asset.computer_name : '-' }}</template>
        </el-table-column>
        <el-table-column prop="backup_target" label="备份对象" width="150" show-overflow-tooltip />
        <el-table-column prop="backup_tool" label="备份工具" width="120" align="center" />
        <el-table-column label="备份介质" width="150" show-overflow-tooltip>
          <template slot-scope="{ row }">{{ row.backup_medium_asset ? row.backup_medium_asset.computer_name : '-' }}</template>
        </el-table-column>
        <el-table-column prop="backup_frequency" label="备份频率" width="80" align="center" />
        <el-table-column prop="retention_policy" label="保留策略" width="90" align="center" show-overflow-tooltip />
        <el-table-column prop="full_backup_strategy" label="全量备份" width="80" align="center" />
        <el-table-column label="所属部门" width="100" align="center">
          <template slot-scope="{ row }">{{ row.department ? row.department.name : '-' }}</template>
        </el-table-column>
        <el-table-column prop="file_name" label="申请表" width="150" show-overflow-tooltip />
        <el-table-column label="操作" width="320" fixed="right" align="center">
          <template slot-scope="{ row }">
            <div class="op-btns">
              <el-button size="mini" type="text" icon="el-icon-view" @click="handlePreview(row)">预览</el-button>
              <el-button size="mini" type="text" icon="el-icon-download" @click="handleDownload(row)">下载</el-button>
              <el-button size="mini" type="text" icon="el-icon-edit" @click="openEdit(row)">编辑</el-button>
              <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="handleDelete(row)">删除</el-button>
              <el-button size="mini" type="text" icon="el-icon-refresh-right" style="color: #409EFF" @click="openRecoveryForm(row)">恢复</el-button>
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

    <!-- 新增/编辑备份记录弹窗 -->
    <el-dialog :title="isEdit ? '编辑备份记录' : '新增备份记录'" :visible.sync="showForm" width="650px" :close-on-click-modal="false">
      <el-form :model="form" ref="formRef" :rules="formRules" label-width="110px">
        <el-form-item label="申请日期" prop="application_date">
          <el-date-picker v-model="form.application_date" type="date" value-format="yyyy-MM-dd" placeholder="选择日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="备份源" prop="backup_source_asset_id">
          <el-select v-model="form.backup_source_asset_id" filterable remote :remote-method="remoteSearchAsset" :loading="assetLoading" placeholder="输入关键词搜索资产" style="width: 100%">
            <el-option v-for="a in sourceAssetOptions" :key="a.id" :label="`${a.computer_name} (${a.ip_address || '无IP'})`" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-row :gutter="16">
          <el-col :span="12">
            <el-form-item label="备份对象类型" prop="backup_target_type">
              <el-select v-model="form.backup_target_type" placeholder="选择类型" style="width: 100%" @change="handleTargetTypeChange">
                <el-option label="系统" value="系统" />
                <el-option label="磁盘分区" value="磁盘分区" />
                <el-option label="目录" value="目录" />
                <el-option label="配置文件" value="配置文件" />
                <el-option label="其他" value="其他" />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="12">
            <el-form-item label="备份对象" prop="backup_target">
              <el-input v-if="form.backup_target_type === '系统' || form.backup_target_type === '配置文件'" :value="form.backup_target_type" disabled />
              <el-input v-else v-model="form.backup_target" :placeholder="backupTargetPlaceholder" />
            </el-form-item>
          </el-col>
        </el-row>
        <el-form-item label="备份工具" prop="backup_tool">
          <el-radio-group v-model="form.backup_tool">
            <el-radio label="Veeam Backup" />
            <el-radio label="FortiConfBak" />
            <el-radio label="HuaweiConfBak" />
          </el-radio-group>
        </el-form-item>
        <el-form-item label="备份介质" prop="backup_medium_asset_id">
          <el-select v-model="form.backup_medium_asset_id" filterable remote :remote-method="remoteSearchAsset" :loading="assetLoading" placeholder="输入关键词搜索资产" style="width: 100%">
            <el-option v-for="a in mediumAssetOptions" :key="a.id" :label="`${a.computer_name} (${a.ip_address || '无IP'})`" :value="a.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="备份频率" prop="backup_frequency">
          <el-radio-group v-model="form.backup_frequency">
            <el-radio label="每天" />
            <el-radio label="每周" />
            <el-radio label="每月" />
          </el-radio-group>
        </el-form-item>
        <el-form-item label="保留策略" prop="retention_policy">
          <el-radio-group v-model="retentionRadio" @change="handleRetentionChange">
            <el-radio label="7天" />
            <el-radio label="15天" />
            <el-radio label="1月" />
            <el-radio label="3月" />
            <el-radio label="1年" />
            <el-radio label="其它" />
          </el-radio-group>
          <el-input v-if="retentionRadio === '其它'" v-model="retentionOtherText" placeholder="请输入自定义保留策略" size="small" style="margin-top: 8px" />
        </el-form-item>
        <el-form-item label="全量备份策略" prop="full_backup_strategy">
          <el-radio-group v-model="form.full_backup_strategy">
            <el-radio label="每天" />
            <el-radio label="每周" />
          </el-radio-group>
        </el-form-item>
        <el-form-item label="所属部门" prop="department_id">
          <el-select v-model="form.department_id" placeholder="选择部门" style="width: 100%">
            <el-option v-for="d in departmentOptions" :key="d.id" :label="d.name" :value="d.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="申请表" v-if="!isEdit">
          <el-upload ref="uploader" action="" :auto-upload="false" :limit="1" accept=".pdf" :on-change="handleFileChange" :on-remove="handleFileRemove" :file-list="fileList" drag>
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
            <div slot="tip" class="el-upload__tip">仅支持 PDF 格式文件</div>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可更换文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showForm = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSubmit">{{ isEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- 恢复记录弹窗 -->
    <el-dialog :title="isRecoveryEdit ? '编辑恢复记录' : '新增恢复记录'" :visible.sync="showRecoveryForm" width="500px" :close-on-click-modal="false">
      <el-form :model="recoveryForm" ref="recoveryFormRef" :rules="recoveryFormRules" label-width="90px">
        <el-form-item label="恢复类型" prop="recovery_type">
          <el-radio-group v-model="recoveryForm.recovery_type">
            <el-radio label="恢复测试" />
            <el-radio label="故障恢复" />
          </el-radio-group>
        </el-form-item>
        <el-form-item label="恢复结果" prop="recovery_result">
          <el-radio-group v-model="recoveryForm.recovery_result">
            <el-radio label="成功" />
            <el-radio label="失败" />
          </el-radio-group>
        </el-form-item>
        <el-form-item label="恢复日期" prop="recovery_date">
          <el-date-picker v-model="recoveryForm.recovery_date" type="date" value-format="yyyy-MM-dd" placeholder="选择日期" style="width: 100%" />
        </el-form-item>
        <el-form-item label="上传记录" v-if="!isRecoveryEdit">
          <el-upload ref="recoveryUploader" action="" :auto-upload="false" :limit="1" accept=".pdf" :on-change="handleRecoveryFileChange" :on-remove="handleRecoveryFileRemove" :file-list="recoveryFileList" drag>
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击上传</em></div>
            <div slot="tip" class="el-upload__tip">仅支持 PDF 格式文件</div>
          </el-upload>
        </el-form-item>
        <el-alert v-else title="编辑模式下不可更换文件" type="info" :closable="false" show-icon />
      </el-form>
      <span slot="footer">
        <el-button @click="showRecoveryForm = false">取消</el-button>
        <el-button type="primary" :loading="recoverySubmitting" @click="handleRecoverySubmit">{{ isRecoveryEdit ? '保存' : '确定上传' }}</el-button>
      </span>
    </el-dialog>

    <!-- 申请表预览弹窗 -->
    <el-dialog title="申请表预览" :visible.sync="previewVisible" width="80%" top="3vh" @close="clearPreview">
      <iframe v-if="pdfBlobUrl" :src="pdfBlobUrl" style="width: 100%; height: 70vh; border: none;" />
      <span slot="footer">
        <el-button type="primary" size="small" icon="el-icon-download" @click="handleDownloadFromPreview">下载</el-button>
      </span>
    </el-dialog>

    <!-- 恢复记录预览弹窗 -->
    <el-dialog title="恢复记录预览" :visible.sync="recoveryPreviewVisible" width="80%" top="3vh" @close="clearRecoveryPreview">
      <iframe v-if="recoveryPdfBlobUrl" :src="recoveryPdfBlobUrl" style="width: 100%; height: 70vh; border: none;" />
      <span slot="footer">
        <el-button type="primary" size="small" icon="el-icon-download" @click="handleDownloadRecoveryFromPreview">下载</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import {
  getBackups, createBackup, updateBackup, deleteBackup,
  getBackupPreviewUrl, getBackupDownloadUrl,
  createBackupRecovery, updateBackupRecovery, deleteBackupRecovery,
  getBackupRecoveryPreviewUrl, getBackupRecoveryDownloadUrl
} from '@/api/backup'
import { getAssets } from '@/api/asset'
import { getDepartments } from '@/api/department'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'BackupManagement',
  components: { DualControlDialog },
  data() {
    const now = new Date()
    return {
      records: [],
      loading: false,
      page: 1,
      pageSize: 10,
      total: 0,
      filterYear: '',
      filterDept: '',
      filterTool: '',
      keyword: '',
      yearOptions: Array.from({ length: 10 }, (_, i) => now.getFullYear() - i),
      departmentOptions: [],
      backupToolOptions: ['Veeam Backup', 'FortiConfBak', 'HuaweiConfBak'],
      // 资产搜索
      sourceAssetOptions: [],
      mediumAssetOptions: [],
      assetLoading: false,
      // 新增/编辑备份
      showForm: false,
      isEdit: false,
      editingId: null,
      submitting: false,
      form: {
        application_date: '',
        backup_source_asset_id: null,
        backup_target_type: '',
        backup_target: '',
        backup_tool: '',
        backup_medium_asset_id: null,
        backup_frequency: '',
        retention_policy: '',
        full_backup_strategy: '',
        department_id: null
      },
      formRules: {
        application_date: [{ required: true, message: '请选择申请日期', trigger: 'change' }],
        backup_source_asset_id: [{ required: true, message: '请选择备份源', trigger: 'change' }],
        backup_target_type: [{ required: true, message: '请选择备份对象类型', trigger: 'change' }],
        backup_tool: [{ required: true, message: '请选择备份工具', trigger: 'change' }],
        backup_medium_asset_id: [{ required: true, message: '请选择备份介质', trigger: 'change' }],
        backup_frequency: [{ required: true, message: '请选择备份频率', trigger: 'change' }],
        retention_policy: [{ required: true, message: '请选择保留策略', trigger: 'change' }],
        full_backup_strategy: [{ required: true, message: '请选择全量备份策略', trigger: 'change' }],
        department_id: [{ required: true, message: '请选择所属部门', trigger: 'change' }]
      },
      retentionRadio: '',
      retentionOtherText: '',
      selectedFile: null,
      fileList: [],
      // 恢复记录
      showRecoveryForm: false,
      isRecoveryEdit: false,
      recoveryEditingId: null,
      recoveryBackupId: null,
      recoverySubmitting: false,
      recoveryForm: {
        recovery_type: '',
        recovery_result: '',
        recovery_date: ''
      },
      recoveryFormRules: {
        recovery_type: [{ required: true, message: '请选择恢复类型', trigger: 'change' }],
        recovery_result: [{ required: true, message: '请选择恢复结果', trigger: 'change' }],
        recovery_date: [{ required: true, message: '请选择恢复日期', trigger: 'change' }]
      },
      recoverySelectedFile: null,
      recoveryFileList: [],
      // 申请表预览
      previewVisible: false,
      previewUrl: '',
      previewDownloadUrl: '',
      previewFileName: '',
      pdfBlobUrl: '',
      // 恢复记录预览
      recoveryPreviewVisible: false,
      recoveryPreviewUrl: '',
      recoveryDownloadUrl: '',
      recoveryPreviewFileName: '',
      recoveryPdfBlobUrl: ''
    }
  },
  computed: {
    backupTargetPlaceholder() {
      const map = {
        '磁盘分区': '请输入磁盘分区',
        '目录': '请输入目录路径',
        '其他': '请输入备份对象'
      }
      return map[this.form.backup_target_type] || '请输入备份对象'
    }
  },
  mounted() {
    this.fetchData()
    this.fetchDepartments()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const params = { page: this.page, page_size: this.pageSize }
        if (this.filterYear) params.year = this.filterYear
        if (this.filterDept) params.department_id = this.filterDept
        if (this.filterTool) params.backup_tool = this.filterTool
        if (this.keyword) params.keyword = this.keyword
        const res = await getBackups(params)
        this.records = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    async fetchDepartments() {
      try {
        const res = await getDepartments()
        this.departmentOptions = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    async remoteSearchAsset(query) {
      if (!query) {
        this.sourceAssetOptions = []
        this.mediumAssetOptions = []
        return
      }
      this.assetLoading = true
      try {
        const res = await getAssets({ search: query, page: 1, page_size: 20 })
        const assets = res.data || []
        this.sourceAssetOptions = assets
        this.mediumAssetOptions = assets
      } catch (e) {
        console.error(e)
      } finally {
        this.assetLoading = false
      }
    },
    handleSizeChange() {
      this.page = 1
      this.fetchData()
    },
    handleFilterChange() {
      this.page = 1
      this.fetchData()
    },
    handleTargetTypeChange(val) {
      if (val === '系统') {
        this.form.backup_target = '系统'
      } else if (val === '配置文件') {
        this.form.backup_target = '配置文件'
      } else {
        this.form.backup_target = ''
      }
    },
    handleRetentionChange(val) {
      if (val !== '其它') {
        this.form.retention_policy = val
        this.retentionOtherText = ''
      } else {
        this.form.retention_policy = ''
      }
      this.$nextTick(() => {
        if (this.$refs.formRef) this.$refs.formRef.clearValidate(['retention_policy'])
      })
    },
    // 新增
    openCreate() {
      this.resetForm()
      this.showForm = true
      this.$nextTick(() => { if (this.$refs.formRef) this.$refs.formRef.clearValidate() })
    },
    // 编辑
    openEdit(row) {
      this.isEdit = true
      this.editingId = row.id
      this.form = {
        application_date: row.application_date || '',
        backup_source_asset_id: row.backup_source_asset_id,
        backup_target_type: row.backup_target_type || '',
        backup_target: row.backup_target || '',
        backup_tool: row.backup_tool || '',
        backup_medium_asset_id: row.backup_medium_asset_id,
        backup_frequency: row.backup_frequency || '',
        retention_policy: row.retention_policy || '',
        full_backup_strategy: row.full_backup_strategy || '',
        department_id: row.department_id
      }
      // 回显保留策略radio
      const stdOptions = ['7天', '15天', '1月', '3月', '1年']
      if (stdOptions.includes(row.retention_policy)) {
        this.retentionRadio = row.retention_policy
        this.retentionOtherText = ''
      } else if (row.retention_policy) {
        this.retentionRadio = '其它'
        this.retentionOtherText = row.retention_policy
      } else {
        this.retentionRadio = ''
        this.retentionOtherText = ''
      }
      // 回显资产选项
      if (row.backup_source_asset) {
        this.sourceAssetOptions = [row.backup_source_asset]
      }
      if (row.backup_medium_asset) {
        this.mediumAssetOptions = [row.backup_medium_asset]
      }
      this.showForm = true
      this.$nextTick(() => { if (this.$refs.formRef) this.$refs.formRef.clearValidate() })
    },
    resetForm() {
      this.isEdit = false
      this.editingId = null
      this.selectedFile = null
      this.fileList = []
      this.retentionRadio = ''
      this.retentionOtherText = ''
      this.sourceAssetOptions = []
      this.mediumAssetOptions = []
      this.form = {
        application_date: '',
        backup_source_asset_id: null,
        backup_target_type: '',
        backup_target: '',
        backup_tool: '',
        backup_medium_asset_id: null,
        backup_frequency: '',
        retention_policy: '',
        full_backup_strategy: '',
        department_id: null
      }
    },
    handleFileChange(file) {
      this.selectedFile = file.raw
    },
    handleFileRemove() {
      this.selectedFile = null
    },
    async handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        // 保留策略"其它"处理
        if (this.retentionRadio === '其它') {
          if (!this.retentionOtherText.trim()) {
            this.$message.warning('请输入自定义保留策略')
            return
          }
          this.form.retention_policy = this.retentionOtherText.trim()
        }
        // 备份对象验证
        if (this.form.backup_target_type !== '系统' && this.form.backup_target_type !== '配置文件' && !this.form.backup_target) {
          this.$message.warning('请输入备份对象名称')
          return
        }
        if (!this.isEdit && !this.selectedFile) {
          this.$message.warning('请选择申请表文件')
          return
        }
        this.submitting = true
        try {
          const dualToken = await this.$refs.dualControl.open()
          const formData = new FormData()
          formData.append('application_date', this.form.application_date || '')
          formData.append('backup_source_asset_id', this.form.backup_source_asset_id || '')
          formData.append('backup_target_type', this.form.backup_target_type || '')
          formData.append('backup_target', this.form.backup_target || '')
          formData.append('backup_tool', this.form.backup_tool || '')
          formData.append('backup_medium_asset_id', this.form.backup_medium_asset_id || '')
          formData.append('backup_frequency', this.form.backup_frequency || '')
          formData.append('retention_policy', this.form.retention_policy || '')
          formData.append('full_backup_strategy', this.form.full_backup_strategy || '')
          formData.append('department_id', this.form.department_id || '')
          if (this.isEdit) {
            await updateBackup(this.editingId, formData, dualToken)
            this.$message.success('更新成功')
          } else {
            formData.append('file', this.selectedFile)
            await createBackup(formData, dualToken)
            this.$message.success('创建成功')
          }
          this.showForm = false
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') {
            console.error(e)
            this.$message.error(this.isEdit ? '更新失败' : '创建失败')
          }
        } finally {
          this.submitting = false
        }
      })
    },
    // 删除备份记录
    async handleDelete(row) {
      try {
        await this.$confirm('确定要删除该备份记录吗？关联的恢复记录也将被删除。此操作不可恢复。', '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteBackup(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    // 申请表预览
    async handlePreview(row) {
      this.clearPreview()
      this.previewUrl = getBackupPreviewUrl(row.id)
      this.previewDownloadUrl = getBackupDownloadUrl(row.id)
      this.previewFileName = row.file_name
      this.previewVisible = true
      await this.fetchPdfAsBlob(this.previewUrl)
    },
    async fetchPdfAsBlob(url) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
        this.pdfBlobUrl = URL.createObjectURL(blob)
      } catch (e) {
        console.error('PDF加载失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    clearPreview() {
      if (this.pdfBlobUrl) {
        URL.revokeObjectURL(this.pdfBlobUrl)
        this.pdfBlobUrl = ''
      }
    },
    // 下载
    async handleDownload(row) {
      await this.downloadWithAuth(getBackupDownloadUrl(row.id), row.file_name)
    },
    async handleDownloadFromPreview() {
      await this.downloadWithAuth(this.previewDownloadUrl, this.previewFileName)
    },
    async downloadWithAuth(url, fileName) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
        const link = document.createElement('a')
        link.href = URL.createObjectURL(blob)
        link.download = fileName || '下载文件'
        link.click()
        URL.revokeObjectURL(link.href)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('文件下载失败')
      }
    },
    // === 恢复记录 ===
    openRecoveryForm(row) {
      this.isRecoveryEdit = false
      this.recoveryEditingId = null
      this.recoveryBackupId = row.id
      this.recoverySelectedFile = null
      this.recoveryFileList = []
      this.recoveryForm = { recovery_type: '', recovery_result: '', recovery_date: '' }
      this.showRecoveryForm = true
      this.$nextTick(() => { if (this.$refs.recoveryFormRef) this.$refs.recoveryFormRef.clearValidate() })
    },
    openRecoveryEdit(recovery, backupRow) {
      this.isRecoveryEdit = true
      this.recoveryEditingId = recovery.id
      this.recoveryBackupId = backupRow.id
      this.recoverySelectedFile = null
      this.recoveryFileList = []
      this.recoveryForm = {
        recovery_type: recovery.recovery_type || '',
        recovery_result: recovery.recovery_result || '',
        recovery_date: recovery.recovery_date || ''
      }
      this.showRecoveryForm = true
      this.$nextTick(() => { if (this.$refs.recoveryFormRef) this.$refs.recoveryFormRef.clearValidate() })
    },
    handleRecoveryFileChange(file) {
      this.recoverySelectedFile = file.raw
    },
    handleRecoveryFileRemove() {
      this.recoverySelectedFile = null
    },
    async handleRecoverySubmit() {
      this.$refs.recoveryFormRef.validate(async valid => {
        if (!valid) return
        if (!this.isRecoveryEdit && !this.recoverySelectedFile) {
          this.$message.warning('请上传恢复记录文件')
          return
        }
        this.recoverySubmitting = true
        try {
          const dualToken = await this.$refs.dualControl.open()
          const formData = new FormData()
          formData.append('recovery_type', this.recoveryForm.recovery_type || '')
          formData.append('recovery_result', this.recoveryForm.recovery_result || '')
          formData.append('recovery_date', this.recoveryForm.recovery_date || '')
          if (this.isRecoveryEdit) {
            await updateBackupRecovery(this.recoveryEditingId, formData, dualToken)
            this.$message.success('更新成功')
          } else {
            formData.append('file', this.recoverySelectedFile)
            await createBackupRecovery(this.recoveryBackupId, formData, dualToken)
            this.$message.success('创建成功')
          }
          this.showRecoveryForm = false
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') {
            console.error(e)
            this.$message.error(this.isRecoveryEdit ? '更新失败' : '创建失败')
          }
        } finally {
          this.recoverySubmitting = false
        }
      })
    },
    async handleDeleteRecovery(recovery) {
      try {
        await this.$confirm('确定要删除该恢复记录吗？此操作不可恢复。', '删除确认', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning'
        })
        const dualToken = await this.$refs.dualControl.open()
        await deleteBackupRecovery(recovery.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    // 恢复记录预览
    async handlePreviewRecovery(recovery) {
      this.clearRecoveryPreview()
      this.recoveryPreviewUrl = getBackupRecoveryPreviewUrl(recovery.id)
      this.recoveryDownloadUrl = getBackupRecoveryDownloadUrl(recovery.id)
      this.recoveryPreviewFileName = recovery.file_name
      this.recoveryPreviewVisible = true
      await this.fetchRecoveryPdfAsBlob(this.recoveryPreviewUrl)
    },
    async fetchRecoveryPdfAsBlob(url) {
      try {
        const token = localStorage.getItem('token')
        const response = await fetch(url, { headers: { 'Authorization': `Bearer ${token}` } })
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`)
        const blob = await response.blob()
        this.recoveryPdfBlobUrl = URL.createObjectURL(blob)
      } catch (e) {
        console.error('PDF加载失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    clearRecoveryPreview() {
      if (this.recoveryPdfBlobUrl) {
        URL.revokeObjectURL(this.recoveryPdfBlobUrl)
        this.recoveryPdfBlobUrl = ''
      }
    },
    async handleDownloadRecovery(recovery) {
      await this.downloadWithAuth(getBackupRecoveryDownloadUrl(recovery.id), recovery.file_name)
    },
    async handleDownloadRecoveryFromPreview() {
      await this.downloadWithAuth(this.recoveryDownloadUrl, this.recoveryPreviewFileName)
    }
  }
}
</script>

<style scoped>
.backup-management {
  margin: 20px;
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
