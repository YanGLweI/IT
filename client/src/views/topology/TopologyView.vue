<template>
  <div class="topology-view">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">网络拓扑图</h2>
        <p class="page-subtitle">管理公司网络拓扑图，支持图片及PDF格式在线预览与下载</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="uploadVisible = true">上传拓扑图</el-button>
      </div>
    </div>

    <el-row :gutter="20">
        <el-col :span="6" v-for="item in topologies" :key="item.id" style="margin-bottom: 20px">
          <el-card shadow="hover">
            <div class="topo-card">
              <div class="topo-thumb" @click="handlePreview(item)">
                <div v-if="isPdfFile(item.file_name)" class="pdf-thumb">
                  <i class="el-icon-document"></i>
                  <span>PDF</span>
                </div>
                <img v-else :src="getThumbUrl(item.id)" :alt="item.name" />
              </div>
              <div class="topo-info">
                <h4>{{ item.name }}</h4>
                <p>{{ item.description || '暂无描述' }}</p>
                <div class="topo-meta">
                  <span>{{ formatSize(item.file_size) }}</span>
                  <span>{{ formatDate(item.created_at) }}</span>
                </div>
                <div class="topo-actions">
                  <el-button size="mini" @click="handlePreview(item)">预览</el-button>
                  <el-button size="mini" @click="handleEdit(item)">编辑</el-button>
                  <el-button size="mini" type="danger" @click="handleDelete(item)">删除</el-button>
                </div>
              </div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <el-empty v-if="topologies.length === 0" description="暂无拓扑图" />

    <!-- 上传弹窗 -->
    <el-dialog class="vault-dialog" title="上传拓扑图" :visible.sync="uploadVisible" width="550px">
      <el-form :model="uploadForm" :rules="uploadRules" ref="uploadFormRef" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="uploadForm.name" placeholder="请输入拓扑图名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="uploadForm.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="文件">
          <el-upload
            drag
            action=""
            :auto-upload="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            accept=".png,.jpg,.jpeg,.gif,.svg,.pdf"
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">将文件拖到此处，或<em>点击选择</em></div>
            <div class="el-upload__tip" slot="tip">支持 PNG、JPG、GIF、SVG、PDF 格式</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </span>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog class="vault-dialog" title="编辑拓扑图" :visible.sync="editVisible" width="500px">
      <el-form :model="editForm" label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="editForm.name" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="editForm.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="editVisible = false">取消</el-button>
        <el-button type="primary" @click="handleEditSubmit">保存</el-button>
      </span>
    </el-dialog>

    <!-- 预览弹窗 - 使用 file-viewer -->
    <el-dialog class="vault-dialog preview-dialog fv-preview-dialog" title="拓扑图预览" :visible.sync="previewVisible" width="90%" top="2vh">
      <div class="fv-container" style="height: 80vh">
        <div v-if="fvFeature.enabled">
          <FileViewer 
            v-if="previewVisible"
            :url="currentFileUrl"
            :name="currentFileName"
            :type="currentFileType"
            :options="fvOptions"
          />
        </div>
        <div v-else class="fv-fallback">
          <p>文件预览功能暂不可用，请下载后查看。</p>
          <el-button type="primary" @click="downloadFile">下载文件</el-button>
        </div>
      </div>
    </el-dialog>
    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getTopologies, createTopology, updateTopology, deleteTopology, getTopologyPreviewUrl, getTopologyDownloadUrl } from '@/api/topology'
import { FileViewer } from '@file-viewer/vue2.7'
import officePreset from '@file-viewer/preset-office'
import DualControlDialog from '@/components/DualControlDialog.vue'
import fvFeature from '@/config/fv-feature'

export default {
  components: { DualControlDialog, FileViewer },
  name: 'TopologyView',
  data() {
    return {
      topologies: [],
      uploadVisible: false,
      uploading: false,
      uploadForm: { name: '', description: '' },
      uploadRules: {
        name: [{ required: true, message: '请输入名称', trigger: 'blur' }]
      },
      fileList: [],
      selectedFile: null,
      editVisible: false,
      editForm: { id: null, name: '', description: '' },
      previewVisible: false,
      currentFileUrl: '',
      currentFileName: '',
      currentFileType: '',
      previewId: null,
      previewFileName: '',
      fvFeature: fvFeature,
      thumbUrls: {} // 存储缩略图的 blob URL
    }
  },
  mounted() {
    this.fetchData()
  },
  computed: {
    fvOptions() {
      return {
        preset: officePreset,
        fetchFile: this.fetchFileWithAuth
      }
    }
  },
  methods: {
    async fetchData() {
      try {
        const res = await getTopologies()
        this.topologies = res.data || []
        // 加载所有缩略图（带 token）
        this.loadThumbnails()
      } catch (e) {
        console.error(e)
      }
    },
    async loadThumbnails() {
      // 清理旧的 blob URL
      Object.values(this.thumbUrls).forEach(url => URL.revokeObjectURL(url))
      this.thumbUrls = {}
      
      for (const item of this.topologies) {
        try {
          const url = getTopologyPreviewUrl(item.id)
          const response = await fetch(url, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
          })
          if (!response.ok) continue
          const blob = await response.blob()
          this.$set(this.thumbUrls, item.id, URL.createObjectURL(blob))
        } catch (e) {
          console.error(`加载缩略图失败 ID=${item.id}:`, e)
        }
      }
    },
    getThumbUrl(id) {
      return this.thumbUrls[id] || ''
    },
    handleFileChange(file) {
      this.selectedFile = file.raw
      this.fileList = [file]
    },
    handleFileRemove() {
      this.selectedFile = null
      this.fileList = []
    },
    async handleUpload() {
      if (!this.$refs.uploadFormRef) return
      this.$refs.uploadFormRef.validate(async valid => {
        if (!valid) {
          console.log('表单验证未通过')
          return
        }
        if (!this.selectedFile) {
          this.$message.warning('请选择文件')
          return
        }
        this.uploading = true
        try {
          const formData = new FormData()
          formData.append('name', this.uploadForm.name)
          formData.append('description', this.uploadForm.description)
          formData.append('file', this.selectedFile)
          await createTopology(formData)
          this.$message.success('上传成功')
          this.uploadVisible = false
          this.uploadForm = { name: '', description: '' }
          this.selectedFile = null
          this.fileList = []
          this.fetchData()
        } catch (e) {
          console.error('上传失败:', e)
          this.$message.error('上传失败')
        } finally {
          this.uploading = false
        }
      })
    },
    handleEdit(item) {
      this.editForm = { id: item.id, name: item.name, description: item.description }
      this.editVisible = true
    },
    async handleEditSubmit() {
      try {
        const dualToken = await this.$refs.dualControl.open()
        await updateTopology(this.editForm.id, {
          name: this.editForm.name,
          description: this.editForm.description
        }, dualToken)
        this.$message.success('更新成功')
        this.editVisible = false
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    async handleDelete(item) {
      try {
        await this.$confirm('确定要删除该拓扑图吗？', '提示', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteTopology(item.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    async handlePreview(item) {
      const url = getTopologyPreviewUrl(item.id)
      const fileExtension = item.file_name ? item.file_name.split('.').pop().toLowerCase() : ''
      
      this.currentFileUrl = url
      this.currentFileName = item.file_name || 'unknown_file'
      this.currentFileType = fileExtension
      this.previewFileName = item.file_name || '拓扑图'
      this.previewId = item.id
      
      // 先关闭再打开，确保 FileViewer 组件重新渲染
      this.previewVisible = false
      await this.$nextTick()
      this.previewVisible = true
      await this.$nextTick()
    },
    async fetchFileWithAuth({ url }) {
      const token = localStorage.getItem('token')
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (!response.ok) {
        throw new Error(`Failed to fetch file: ${response.status} ${response.statusText}`)
      }
      return response.arrayBuffer()
    },
    async downloadFile() {
      if (this.previewId) {
        const url = getTopologyDownloadUrl(this.previewId)
        try {
          const response = await fetch(url, {
            headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
          })
          if (!response.ok) throw new Error('下载失败')
          const blob = await response.blob()
          const downloadUrl = URL.createObjectURL(blob)
          const link = document.createElement('a')
          link.href = downloadUrl
          link.download = this.previewFileName
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          URL.revokeObjectURL(downloadUrl)
        } catch (e) {
          console.error('下载失败:', e)
          this.$message.error('下载失败')
        }
      }
    },
    formatSize(size) {
      if (!size) return '-'
      if (size < 1024) return size + ' B'
      if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
      return (size / (1024 * 1024)).toFixed(1) + ' MB'
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      return new Date(dateStr).toLocaleString('zh-CN')
    },
    isPdfFile(fileName) {
      if (!fileName) return false
      return fileName.toLowerCase().endsWith('.pdf')
    }
  }
}
</script>

<style scoped>
.topology-view {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow-y: auto;
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

/* --- 卡片 --- */
.topo-card {
  display: flex;
  flex-direction: column;
}
.topo-thumb {
  height: 180px;
  overflow: hidden;
  border-radius: 12px;
  cursor: pointer;
  background: #f8fafc;
  display: flex;
  align-items: center;
  justify-content: center;
}
.topo-thumb img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.pdf-thumb {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: #e74c3c;
}
.pdf-thumb i {
  font-size: 48px;
}
.pdf-thumb span {
  font-size: 14px;
  font-weight: 600;
  letter-spacing: 1px;
}
/* file-viewer 组件高度修复 - 强制填满容器 */
.fv-container > div {
  height: 100% !important;
}

.fv-container >>> .ff-file-viewer-vue27 {
  height: 100% !important;
}

/* file-viewer 内部工具栏样式覆盖 */
.fv-container >>> .fv-toolbar {
  border-bottom: 1px solid #e2e8f0;
  padding: 8px 12px;
  background: #ffffff;
}

/* 确保滚动正确工作 */
.fv-container >>> .fv-content-wrapper {
  height: calc(100% - 50px);
}

/* 降级方案样式 */
.fv-fallback {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 100%;
  gap: 16px;
  color: #64748b;
}

.topo-info {
  margin-top: 10px;
}
.topo-info h4 {
  margin: 0 0 5px 0;
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
}
.topo-info p {
  margin: 0 0 8px 0;
  color: #64748b;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.topo-meta {
  display: flex;
  justify-content: space-between;
  color: #94a3b8;
  font-size: 12px;
  margin-bottom: 8px;
}
.topo-actions {
  display: flex;
  justify-content: flex-end;
  gap: 6px;
}
/* 卡片操作按钮 — 符合设计规范 */
.topo-actions .el-button {
  border-radius: 10px;
  background: transparent;
  border: 1px solid #e2e8f0;
  color: #64748b;
  font-size: 12px;
  padding: 5px 12px;
}
.topo-actions .el-button:hover {
  border-color: #94a3b8;
  color: #1e293b;
}
.topo-actions .el-button--danger {
  background: rgba(239, 68, 68, 0.1);
  border-color: rgba(239, 68, 68, 0.3);
  color: #ef4444;
}
.topo-actions .el-button--danger:hover {
  background: #ef4444;
  border-color: #ef4444;
  color: #fff;
}
</style>
