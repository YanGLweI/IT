<template>
  <div class="backup-management">

    <!-- ==================== 区块一：模板管理 ==================== -->
    <el-card style="margin-bottom: 20px">
      <div slot="header" class="page-header">
        <span>备份与恢复记录表模板</span>
        <div class="page-header-right">
          <el-button type="primary" size="small" icon="el-icon-upload2" @click="showTemplateUpload = true">上传新版本</el-button>
        </div>
      </div>

      <!-- 当前版本信息 -->
      <div v-if="currentTemplate" class="current-template-info">
        <div class="current-template-row">
          <span class="label">当前版本：</span>
          <el-tag type="success" size="medium">{{ currentTemplate.version }}</el-tag>
          <span style="margin-left: 12px; color: #909399; font-size: 13px">
            {{ currentTemplate.file_name }} · {{ formatSize(currentTemplate.file_size) }} · {{ formatDate(currentTemplate.created_at) }}
          </span>
          <span v-if="currentTemplate.description" style="margin-left: 12px; color: #606266; font-size: 13px">
            （{{ currentTemplate.description }}）
          </span>
          <el-button type="primary" size="mini" icon="el-icon-view" style="margin-left: 16px" @click="previewTemplate(currentTemplate)">预览</el-button>
          <el-button type="default" size="mini" icon="el-icon-download" @click="downloadTemplate(currentTemplate)">下载当前模板</el-button>
        </div>
      </div>
      <el-empty v-else description="暂无模板，请上传第一个版本" :image-size="60" style="padding: 16px 0" />

      <!-- 历史版本折叠面板 -->
      <el-collapse v-if="templateHistory.length > 0" v-model="templateCollapseActive" style="margin-top: 16px">
        <el-collapse-item title="历史版本" name="history">
          <el-table :data="templateHistory" border size="small">
            <el-table-column type="index" label="序号" width="56" align="center" />
            <el-table-column prop="version" label="版本号" width="110" align="center" />
            <el-table-column prop="description" label="版本说明" min-width="160" show-overflow-tooltip>
              <template slot-scope="{ row }">{{ row.description || '-' }}</template>
            </el-table-column>
            <el-table-column prop="file_name" label="文件名" min-width="180" show-overflow-tooltip />
            <el-table-column label="文件大小" width="100" align="center">
              <template slot-scope="{ row }">{{ formatSize(row.file_size) }}</template>
            </el-table-column>
            <el-table-column label="上传时间" width="180" align="center">
              <template slot-scope="{ row }">{{ formatDate(row.created_at) }}</template>
            </el-table-column>
            <el-table-column label="操作" width="200" align="center">
              <template slot-scope="{ row }">
                <el-button size="mini" type="text" icon="el-icon-view" @click="previewTemplate(row)">预览</el-button>
                <el-button size="mini" type="text" icon="el-icon-download" @click="downloadTemplate(row)">下载</el-button>
                <el-button size="mini" type="text" icon="el-icon-delete" style="color: #F56C6C" @click="deleteTemplate(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-collapse-item>
      </el-collapse>
    </el-card>

    <!-- ==================== 区块二：备份管理 ==================== -->
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
                <el-table-column prop="file_name" label="恢复与还原记录表" show-overflow-tooltip />
                <el-table-column label="操作" width="300" align="center">
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
        <el-table-column prop="file_name" label="申请表" min-width="150" show-overflow-tooltip />
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
          <el-select v-model="form.backup_source_asset_id" filterable remote :remote-method="remoteSearchSourceAsset" :loading="assetLoading" placeholder="输入关键词搜索资产" style="width: 100%">
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
          <el-select v-model="form.backup_medium_asset_id" filterable remote :remote-method="remoteSearchMediumAsset" :loading="assetLoading" placeholder="输入关键词搜索资产" style="width: 100%">
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

    <!-- 模板上传弹窗 -->
    <el-dialog title="上传新版本模板" :visible.sync="showTemplateUpload" width="520px" :close-on-click-modal="false">
      <el-form :model="templateForm" ref="templateFormRef" :rules="templateRules" label-width="90px">
        <el-form-item label="版本号" prop="version">
          <el-input v-model="templateForm.version" placeholder="如：IT03-1.0" />
        </el-form-item>
        <el-form-item label="版本说明">
          <el-input v-model="templateForm.description" type="textarea" :rows="2" placeholder="简要说明本次变更内容" />
        </el-form-item>
        <el-form-item label="模板文件" prop="file">
          <el-upload ref="templateUploader" action="" :auto-upload="false" :limit="1" accept=".docx,.pdf"
            :on-change="handleTemplateFileChange" :on-remove="handleTemplateFileRemove" :file-list="templateFileList" drag>
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">拖拽文件到此处，或<em>点击选择</em></div>
            <div slot="tip" class="el-upload__tip">支持 DOCX、PDF 格式</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="showTemplateUpload = false">取消</el-button>
        <el-button type="primary" :loading="templateUploading" @click="submitTemplateUpload">确定上传</el-button>
      </span>
    </el-dialog>

    <!-- 模板预览弹窗 -->
    <el-dialog :visible.sync="templatePreviewVisible" width="80%" top="3vh" @closed="clearTemplatePreview">
      <div slot="title">
        <span>模板预览</span>
      </div>
      <iframe v-if="templatePreviewType === 'pdf'" :src="templatePreviewUrl" style="width: 100%; height: 70vh; border: none" />
      <div v-else-if="templatePreviewType === 'docx'" style="height: 70vh; overflow: auto; border: 1px solid #eee; padding: 20px">
        <div ref="templateDocxContainer" class="docx-preview-container"></div>
      </div>
      <div v-else style="text-align: center; padding: 40px">
        <p>该文件格式不支持在线预览</p>
      </div>
      <span slot="footer">
        <el-button type="primary" size="small" icon="el-icon-download" @click="downloadTemplate(templatePreviewRow)">下载</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import {
  getBackups, createBackup, updateBackup, deleteBackup,
  getBackupPreviewUrl, getBackupDownloadUrl,
  createBackupRecovery, updateBackupRecovery, deleteBackupRecovery,
  getBackupRecoveryPreviewUrl, getBackupRecoveryDownloadUrl,
  getBackupTemplates, uploadBackupTemplate, deleteBackupTemplate,
  getBackupTemplateDownloadUrl, getBackupTemplatePreviewUrl
} from '@/api/backup'
import { getAssets } from '@/api/asset'
import { getDepartments } from '@/api/department'
import { renderAsync } from 'docx-preview'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'BackupManagement',
  components: { DualControlDialog },
  data() {
    const now = new Date()
    return {
      // 模板相关
      currentTemplate: null,
      templateHistory: [],
      templateCollapseActive: [],
      showTemplateUpload: false,
      templateUploading: false,
      templateForm: { version: '', description: '' },
      templateRules: {
        version: [{ required: true, message: '请输入版本号', trigger: 'blur' }]
      },
      templateSelectedFile: null,
      templateFileList: [],
      // 模板预览
      templatePreviewVisible: false,
      templatePreviewUrl: '',
      templatePreviewType: '',
      templatePreviewRow: null,
      // 备份记录
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
    this.fetchTemplates()
    this.fetchData()
    this.fetchDepartments()
  },
  methods: {
    // ============ 模板管理 ============
    async fetchTemplates() {
      try {
        const res = await getBackupTemplates()
        const list = res.data || []
        this.currentTemplate = list.find(t => t.is_current) || null
        this.templateHistory = list.filter(t => !t.is_current)
      } catch (e) {
        console.error(e)
      }
    },
    handleTemplateFileChange(file) {
      this.templateSelectedFile = file.raw
    },
    handleTemplateFileRemove() {
      this.templateSelectedFile = null
    },
    submitTemplateUpload() {
      this.$refs.templateFormRef.validate(async valid => {
        if (!valid) return
        if (!this.templateSelectedFile) {
          this.$message.warning('请选择模板文件')
          return
        }
        this.templateUploading = true
        try {
          const formData = new FormData()
          formData.append('version', this.templateForm.version)
          formData.append('description', this.templateForm.description || '')
          formData.append('file', this.templateSelectedFile)
          const dualToken = await this.$refs.dualControl.open()
          await uploadBackupTemplate(formData, dualToken)
          this.$message.success('上传成功')
          this.showTemplateUpload = false
          this.templateForm = { version: '', description: '' }
          this.templateSelectedFile = null
          this.templateFileList = []
          if (this.$refs.templateUploader) this.$refs.templateUploader.clearFiles()
          this.fetchTemplates()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        } finally {
          this.templateUploading = false
        }
      })
    },
    async downloadTemplate(row) {
      const url = getBackupTemplateDownloadUrl(row.id)
      try {
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('下载失败')
        const blob = await response.blob()
        const downloadUrl = URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = downloadUrl
        link.download = row.file_name
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        URL.revokeObjectURL(downloadUrl)
      } catch (e) {
        console.error('下载失败:', e)
        this.$message.error('下载失败')
      }
    },
    async previewTemplate(row) {
      const url = getBackupTemplatePreviewUrl(row.id)
      const fileName = (row.file_name || '').toLowerCase()
      this.templatePreviewRow = row
      this.templatePreviewUrl = ''

      if (fileName.endsWith('.pdf')) {
        this.templatePreviewType = 'pdf'
      } else if (fileName.endsWith('.docx')) {
        this.templatePreviewType = 'docx'
      } else {
        this.templatePreviewType = 'other'
      }

      this.templatePreviewVisible = true

      try {
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.templatePreviewUrl = URL.createObjectURL(blob)

        if (this.templatePreviewType === 'docx') {
          this.$nextTick(() => {
            this.renderTemplateDocx(blob)
          })
        }
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('模板预览失败')
      }
    },
    async renderTemplateDocx(blob) {
      try {
        const arrayBuffer = await blob.arrayBuffer()
        const container = this.$refs.templateDocxContainer
        if (container) {
          container.innerHTML = ''
          await renderAsync(arrayBuffer, container)
        }
      } catch (e) {
        console.error('docx渲染失败:', e)
        this.$message.error('文件预览失败，请尝试下载后查看')
      }
    },
    clearTemplatePreview() {
      if (this.$refs.templateDocxContainer) {
        this.$refs.templateDocxContainer.innerHTML = ''
      }
      if (this.templatePreviewUrl) {
        URL.revokeObjectURL(this.templatePreviewUrl)
        this.templatePreviewUrl = ''
      }
      this.templatePreviewRow = null
    },
    async deleteTemplate(row) {
      try {
        await this.$confirm(`确定要删除版本 ${row.version} 吗？`, '删除确认', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteBackupTemplate(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchTemplates()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    formatSize(bytes) {
      if (!bytes) return '-'
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const d = new Date(dateStr)
      return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0') + ' ' + String(d.getHours()).padStart(2, '0') + ':' + String(d.getMinutes()).padStart(2, '0')
    },
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
    async remoteSearchSourceAsset(query) {
      if (!query) {
        this.sourceAssetOptions = []
        return
      }
      this.assetLoading = true
      try {
        const res = await getAssets({ search: query, page: 1, page_size: 20 })
        this.sourceAssetOptions = res.data || []
      } catch (e) {
        console.error(e)
      } finally {
        this.assetLoading = false
      }
    },
    async remoteSearchMediumAsset(query) {
      if (!query) {
        this.mediumAssetOptions = []
        return
      }
      this.assetLoading = true
      try {
        const res = await getAssets({ search: query, page: 1, page_size: 20 })
        this.mediumAssetOptions = res.data || []
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
