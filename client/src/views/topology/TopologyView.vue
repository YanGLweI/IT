<template>
  <div class="topology-view">
    <el-card>
      <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
        <span>网络拓扑图</span>
        <el-button type="primary" size="small" icon="el-icon-upload2" @click="uploadVisible = true">上传拓扑图</el-button>
      </div>

      <el-row :gutter="20">
        <el-col :span="8" v-for="item in topologies" :key="item.id" style="margin-bottom: 20px">
          <el-card shadow="hover">
            <div class="topo-card">
              <div class="topo-thumb" @click="handlePreview(item)">
                <img :src="getThumbUrl(item.id)" :alt="item.name" />
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
    </el-card>

    <!-- 上传弹窗 -->
    <el-dialog title="上传拓扑图" :visible.sync="uploadVisible" width="550px">
      <el-form :model="uploadForm" :rules="uploadRules" ref="uploadFormRef" label-width="80px">
        <el-form-item label="名称" prop="name">
          <el-input v-model="uploadForm.name" placeholder="请输入拓扑图名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="uploadForm.description" type="textarea" :rows="2" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="图片">
          <el-upload
            drag
            :auto-upload="false"
            :on-change="handleFileChange"
            :on-remove="handleFileRemove"
            :file-list="fileList"
            accept=".png,.jpg,.jpeg,.gif,.svg"
          >
            <i class="el-icon-upload"></i>
            <div class="el-upload__text">将图片拖到此处，或<em>点击选择</em></div>
            <div class="el-upload__tip" slot="tip">支持 PNG、JPG、GIF、SVG 格式</div>
          </el-upload>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="uploadVisible = false">取消</el-button>
        <el-button type="primary" :loading="uploading" @click="handleUpload">上传</el-button>
      </span>
    </el-dialog>

    <!-- 编辑弹窗 -->
    <el-dialog title="编辑拓扑图" :visible.sync="editVisible" width="500px">
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

    <!-- 预览弹窗 -->
    <el-dialog :visible.sync="previewVisible" width="90%" top="3vh" @closed="clearPreview">
      <div slot="title" class="preview-toolbar">
        <span>拓扑图预览</span>
        <div class="preview-toolbar-right">
          <el-button type="primary" size="small" icon="el-icon-download" @click="downloadFile">下载</el-button>
        </div>
      </div>
      <div style="text-align: center; overflow: auto; max-height: 80vh">
        <img :src="previewUrl" style="max-width: 100%" />
      </div>
    </el-dialog>
    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getTopologies, createTopology, updateTopology, deleteTopology, getTopologyPreviewUrl, getTopologyDownloadUrl } from '@/api/topology'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  components: { DualControlDialog },
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
      previewUrl: '',
      previewId: null,
      previewFileName: '',
      thumbUrls: {} // 存储缩略图的 blob URL
    }
  },
  mounted() {
    this.fetchData()
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
          this.$message.warning('请选择图片')
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
      this.previewId = item.id
      this.previewFileName = item.file_name || '拓扑图'
      this.previewVisible = true
      
      // 使用 fetch 带 token 获取文件并创建 blob URL
      try {
        const url = getTopologyPreviewUrl(item.id)
        const response = await fetch(url, {
          headers: { 'Authorization': `Bearer ${localStorage.getItem('token')}` }
        })
        if (!response.ok) throw new Error('预览失败')
        const blob = await response.blob()
        this.previewUrl = URL.createObjectURL(blob)
      } catch (e) {
        console.error('预览失败:', e)
        this.$message.error('图片预览失败')
      }
    },
    clearPreview() {
      if (this.previewUrl) {
        URL.revokeObjectURL(this.previewUrl)
        this.previewUrl = ''
      }
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
    }
  }
}
</script>

<style scoped>
.topo-card {
  display: flex;
  flex-direction: column;
}
.topo-thumb {
  height: 180px;
  overflow: hidden;
  border-radius: 4px;
  cursor: pointer;
  background: #f5f7fa;
  display: flex;
  align-items: center;
  justify-content: center;
}
.topo-thumb img {
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
}
.topo-info {
  margin-top: 10px;
}
.topo-info h4 {
  margin: 0 0 5px 0;
  font-size: 16px;
}
.topo-info p {
  margin: 0 0 8px 0;
  color: #999;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.topo-meta {
  display: flex;
  justify-content: space-between;
  color: #bbb;
  font-size: 12px;
  margin-bottom: 8px;
}
.topo-actions {
  display: flex;
  gap: 5px;
}
.preview-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
}
.preview-toolbar-right {
  display: flex;
  align-items: center;
  margin-right: 30px;
}
</style>
