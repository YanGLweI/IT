<template>
  <div class="it-guide-page">
    <div class="page-card">
    <!-- 页面标题 -->
    <div class="page-header">
      <h1 class="page-title">
        <svg viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="#409EFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
          <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
        </svg>
        IT指南管理
      </h1>
      <el-button type="primary" icon="el-icon-plus" @click="showCreateDialog">新建指南</el-button>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <div class="search-wrap">
        <el-input v-model="filters.keyword" placeholder="搜索指南名称..." prefix-icon="el-icon-search" clearable @input="handleSearch" />
      </div>
      <el-select v-model="filters.guide_type" placeholder="全部类型" clearable @change="handleSearch" class="filter-select">
        <el-option label="步骤指南" value="step" />
        <el-option label="视频指南" value="video" />
      </el-select>
      <el-select v-model="filters.is_published" placeholder="全部状态" clearable @change="handleSearch" class="filter-select">
        <el-option label="已发布" value="true" />
        <el-option label="草稿" value="false" />
      </el-select>
      <el-select v-model="filters.category" placeholder="全部分类" clearable @change="handleSearch" class="filter-select">
        <el-option v-for="cat in categories" :key="cat" :label="cat" :value="cat" />
      </el-select>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-wrap">
      <el-skeleton :rows="4" animated /><el-skeleton :rows="4" animated /><el-skeleton :rows="4" animated />
    </div>

    <!-- 卡片网格 -->
    <div v-else-if="items.length" class="guide-grid">
      <div v-for="item in items" :key="item.id" class="guide-card">
        <div class="card-icon" :class="item.guide_type === 'step' ? 'icon-step' : 'icon-video'">
          <svg v-if="item.guide_type === 'step'" viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
          <svg v-else viewBox="0 0 24 24" width="28" height="28" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
        </div>
        <div class="card-body">
          <h3 class="card-title" :title="item.title">{{ item.title }}</h3>
          <p class="card-desc" :title="item.description">{{ item.description || '暂无描述' }}</p>
          <div class="card-meta">
            <span class="meta-tag" :class="item.guide_type === 'step' ? 'tag-step' : 'tag-video'">
              {{ item.guide_type === 'step' ? '步骤指南' : '视频指南' }}
            </span>
            <span class="meta-tag tag-category" v-if="item.category">{{ item.category }}</span>
          </div>
          <div class="card-meta" style="margin-top: 6px;">
            <el-tag :type="item.is_published ? 'success' : 'warning'" size="mini" effect="plain">
              {{ item.is_published ? '已发布' : '草稿' }}
            </el-tag>
          </div>
        </div>
        <div class="card-footer">
          <a class="action-link" @click.stop="handleEdit(item)">编辑</a>
          <a class="action-link" @click.stop="openAttachDialog(item)">附件</a>
          <a class="action-link" @click.stop="handleTogglePublish(item)" :style="{ color: item.is_published ? '#D97706' : '#16A34A' }">
            {{ item.is_published ? '取消发布' : '发布' }}
          </a>
          <a class="action-link action-danger" @click.stop="handleDelete(item)">删除</a>
        </div>
      </div>
    </div>

    <!-- 空状态 -->
    <div v-else class="empty-state">
      <svg viewBox="0 0 24 24" width="64" height="64" fill="none" stroke="#CBD5E1" stroke-width="1.5"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/></svg>
      <p class="empty-title">暂无IT指南</p>
      <p class="empty-desc">点击上方按钮创建第一个指南</p>
    </div>

    <!-- 创建/编辑弹窗 -->
    <el-dialog class="vault-dialog" :title="dialogMode === 'create' ? '新建IT指南' : '编辑IT指南'" :visible.sync="dialogVisible" width="720px" :close-on-click-modal="false" @closed="resetDialog">
      <!-- 步骤1: 基本信息 -->
      <div v-if="dialogStep === 1">
        <div class="type-selector">
          <div class="type-card" :class="{ active: form.guide_type === 'step' }" @click="form.guide_type = 'step'">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="#2563EB" stroke-width="1.5"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/><line x1="8" y1="7" x2="16" y2="7"/><line x1="8" y1="11" x2="14" y2="11"/></svg>
            <div class="type-name">步骤指南</div>
            <div class="type-desc">按步骤分解操作指引，每步可添加图文视频</div>
          </div>
          <div class="type-card" :class="{ active: form.guide_type === 'video' }" @click="form.guide_type = 'video'">
            <svg viewBox="0 0 24 24" width="48" height="48" fill="none" stroke="#D97706" stroke-width="1.5"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            <div class="type-name">视频指南</div>
            <div class="type-desc">文字说明配合教程视频，直观易懂</div>
          </div>
        </div>
        <el-form :model="form" label-width="80px" class="dialog-form">
          <el-form-item label="指南标题" required>
            <el-input v-model="form.title" placeholder="请输入指南标题" maxlength="200" show-word-limit />
          </el-form-item>
          <el-form-item label="分类">
            <el-input v-model="form.category" placeholder="如：系统操作、网络配置" />
          </el-form-item>
          <el-form-item label="描述" v-if="form.guide_type !== 'video'">
            <el-input v-model="form.description" type="textarea" :rows="3" placeholder="简要描述指南内容" />
          </el-form-item>
        </el-form>
        <div class="dialog-footer">
          <el-button @click="dialogVisible = false">取消</el-button>
          <el-button type="primary" @click="goToStep2" :disabled="!form.title || !form.guide_type">下一步</el-button>
        </div>
      </div>

      <!-- 步骤2: 内容编辑 -->
      <div v-if="dialogStep === 2">
        <div class="step-header">
          <span class="step-label">{{ form.guide_type === 'step' ? '步骤编辑器' : '视频指南编辑' }}</span>
          <el-button v-if="form.guide_type === 'step'" type="text" icon="el-icon-plus" @click="addStep" class="add-step-btn">添加步骤</el-button>
        </div>

        <!-- 步骤指南编辑器 -->
        <div v-if="form.guide_type === 'step'">
          <div v-for="(step, idx) in form.steps" :key="idx" class="step-card" :class="{ dragging: dragIndex === idx, 'is-focused': focusedStepIdx === idx }" draggable="true" @dragstart="onDragStart(idx)" @dragover.prevent="onDragOver(idx)" @drop="onDrop(idx)" @dragend="onDragEnd" @mousedown="focusedStepIdx = idx">
            <div class="step-drag-handle">
              <svg viewBox="0 0 24 24" width="20" height="20" fill="#94A3B8"><circle cx="9" cy="6" r="1.5"/><circle cx="15" cy="6" r="1.5"/><circle cx="9" cy="12" r="1.5"/><circle cx="15" cy="12" r="1.5"/><circle cx="9" cy="18" r="1.5"/><circle cx="15" cy="18" r="1.5"/></svg>
            </div>
            <div class="step-content">
              <div class="step-title-row">
                <span class="step-number">步骤 {{ idx + 1 }}</span>
                <el-popconfirm title="确定删除此步骤？" @confirm="removeStep(idx)">
                  <el-button slot="reference" type="text" icon="el-icon-close" class="step-delete-btn" />
                </el-popconfirm>
              </div>
              <el-input v-model="step.title" placeholder="步骤标题" style="margin-bottom: 10px;" />
              <el-input v-model="step.description" type="textarea" :rows="2" placeholder="步骤描述" style="margin-bottom: 10px;" />

              <!-- 图片上传 -->
              <div class="media-section">
                <div class="media-label">图片指引 <span class="paste-hint">支持 Ctrl+V 粘贴</span></div>
                <el-upload action="#" :http-request="noopUpload" :file-list="step.images" list-type="picture-card" :on-change="(f, fl) => onImageChange(idx, fl)" :on-remove="(f, fl) => onImageRemove(idx, fl)" :before-upload="beforeImageUpload" accept="image/*" multiple>
                  <i class="el-icon-plus"></i>
                </el-upload>
              </div>

              <!-- 视频上传 -->
              <div class="media-section">
                <div class="media-label">视频指引（可选）</div>
                <!-- 模式切换 -->
                <div style="margin-bottom: 10px;">
                  <el-radio-group v-model="step.videoMode" size="small">
                    <el-radio-button label="upload">上传视频</el-radio-button>
                    <el-radio-button label="embed">嵌入视频</el-radio-button>
                  </el-radio-group>
                </div>
                <!-- 上传模式 -->
                <template v-if="step.videoMode === 'upload'">
                  <div v-if="step.videoFile" class="video-preview-wrap">
                    <video :src="step.videoUrl" controls style="width: 100%; max-height: 200px; border-radius: 12px; border: 1px solid #E2E8F0;"></video>
                    <el-button class="remove-video-btn" @click="removeStepVideo(idx)" title="移除视频">
                      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    </el-button>
                  </div>
                  <el-upload v-else action="#" :http-request="noopUpload" :show-file-list="false" :before-upload="(f) => beforeVideoUpload(f, idx)" accept="video/mp4,video/webm">
                    <el-button size="small" icon="el-icon-video-camera">上传视频</el-button>
                  </el-upload>
                </template>
                <!-- 嵌入模式 -->
                <template v-else>
                  <div v-if="step.embedUrl" class="video-preview-wrap">
                    <iframe :src="step.embedUrl" style="width: 100%; height: 200px; border-radius: 12px; border: 1px solid #E2E8F0;" frameborder="0" allowfullscreen></iframe>
                    <el-button class="remove-video-btn" @click="removeStepEmbedVideo(idx)" title="移除嵌入视频">
                      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                    </el-button>
                  </div>
                  <div v-else>
                    <el-input v-model="step.embedInput" type="textarea" :rows="3" placeholder="粘贴iframe嵌入代码或视频URL..." style="margin-bottom: 8px;" />
                    <el-button size="small" type="primary" @click="applyStepEmbed(idx)">应用嵌入视频</el-button>
                  </div>
                </template>
              </div>
            </div>
          </div>
          <div v-if="!form.steps.length" class="empty-steps">
            <p>点击上方"添加步骤"开始创建指南内容</p>
          </div>
        </div>

        <!-- 视频指南编辑器 -->
        <div v-else>
          <el-input v-model="form.videoDescription" type="textarea" :rows="8" placeholder="请输入指南说明..." class="video-desc-input" />
          <div class="media-section" style="margin-top: 20px;">
            <div class="media-label">教程视频</div>
            <!-- 模式切换 -->
            <div style="margin-bottom: 10px;">
              <el-radio-group v-model="form.videoMode" size="small">
                <el-radio-button label="upload">上传视频</el-radio-button>
                <el-radio-button label="embed">嵌入视频</el-radio-button>
              </el-radio-group>
            </div>
            <!-- 上传模式 -->
            <template v-if="form.videoMode === 'upload'">
              <div v-if="form.videoFile" class="video-preview-wrap">
                <video :src="form.videoUrl" controls style="width: 100%; max-height: 360px; border-radius: 12px; border: 1px solid #E2E8F0;"></video>
                <el-button class="remove-video-btn" @click="removeGuideVideo" title="移除视频">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </el-button>
              </div>
              <el-upload v-else drag action="#" :http-request="noopUpload" :show-file-list="false" :before-upload="beforeGuideVideoUpload" accept="video/mp4,video/webm" class="video-drag-upload">
                <i class="el-icon-upload2" style="font-size: 40px; color: #94A3B8;"></i>
                <div class="el-upload__text">将视频文件拖到此处，或<em>点击上传</em></div>
                <div slot="tip" class="el-upload__tip">仅支持 MP4/WebM 格式，最大 200MB</div>
              </el-upload>
            </template>
            <!-- 嵌入模式 -->
            <template v-else>
              <div v-if="form.embedUrl" class="video-preview-wrap">
                <iframe :src="form.embedUrl" style="width: 100%; height: 360px; border-radius: 12px; border: 1px solid #E2E8F0;" frameborder="0" allowfullscreen></iframe>
                <el-button class="remove-video-btn" @click="removeGuideEmbedVideo" title="移除嵌入视频">
                  <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                </el-button>
              </div>
              <div v-else>
                <el-input v-model="form.embedInput" type="textarea" :rows="4" placeholder="粘贴iframe嵌入代码或视频URL..." style="margin-bottom: 8px;" />
                <el-button size="small" type="primary" @click="applyGuideEmbed">应用嵌入视频</el-button>
              </div>
            </template>
          </div>
        </div>

        <div class="dialog-footer">
          <el-button @click="dialogStep = 1">上一步</el-button>
          <el-button type="primary" @click="handleSave" :loading="saving">保存指南</el-button>
        </div>
      </div>
    </el-dialog>

    <!-- 附件管理弹窗 -->
    <el-dialog class="vault-dialog" :title="'附件关联 - ' + attachGuideTitle" :visible.sync="attachDialogVisible" width="600px" :close-on-click-modal="false">
      <div v-loading="attachLoading">
        <!-- 下载链接区域 -->
        <div class="attach-section">
          <div class="attach-section-title">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
            下载链接
          </div>
          <div class="attach-link-form">
            <el-input v-model="attachLinkForm.label" placeholder="标签名称" size="small" style="width: 140px;" />
            <el-input v-model="attachLinkForm.url" placeholder="https://..." size="small" style="flex: 1;" />
            <el-button type="primary" size="small" icon="el-icon-plus" @click="handleAddLink" :loading="attachAddingLink">添加</el-button>
          </div>
          <div v-if="attachLinks.length" class="attach-list">
            <div v-for="item in attachLinks" :key="item.id" class="attach-item">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="#409EFF" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
              <span class="attach-item-label" :title="item.url">{{ item.label }}</span>
              <span class="attach-item-url">{{ item.url }}</span>
              <el-popconfirm title="确定删除此链接？" @confirm="handleDeleteAttach(item)">
                <el-button slot="reference" type="text" icon="el-icon-close" class="attach-delete-btn" />
              </el-popconfirm>
            </div>
          </div>
          <div v-else class="attach-empty">暂无下载链接</div>
        </div>

        <!-- 上传附件区域 -->
        <div class="attach-section">
          <div class="attach-section-title">
            <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="17 8 12 3 7 8"/><line x1="12" y1="3" x2="12" y2="15"/></svg>
            上传附件
          </div>
          <el-upload
            drag
            action="#"
            :http-request="noopUpload"
            :show-file-list="false"
            :before-upload="beforeAttachUpload"
            accept=".exe,.msi,.zip,.rar,.7z,.tar,.gz,.pdf,.docx,.xlsx,.pptx,.txt,.bat,.sh,.dmg,.iso"
            class="attach-upload"
          >
            <i class="el-icon-upload2" style="font-size: 32px; color: #94A3B8;"></i>
            <div class="el-upload__text">将文件拖到此处，或<em>点击上传</em></div>
            <div slot="tip" class="el-upload__tip">支持 exe/msi/zip/rar/7z/tar/gz/pdf/docx/xlsx/pptx/txt/bat/sh/dmg，最大 100MB</div>
          </el-upload>
          <div v-if="attachFiles.length" class="attach-list">
            <div v-for="item in attachFiles" :key="item.id" class="attach-item">
              <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="#64748B" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>
              <span class="attach-item-label">{{ item.label || item.file_name }}</span>
              <span class="attach-item-size">{{ formatFileSize(item.file_size) }}</span>
              <el-popconfirm title="确定删除此附件？" @confirm="handleDeleteAttach(item)">
                <el-button slot="reference" type="text" icon="el-icon-close" class="attach-delete-btn" />
              </el-popconfirm>
            </div>
          </div>
          <div v-else class="attach-empty">暂无上传附件</div>
        </div>
      </div>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
    </div>
  </div>
</template>

<script>
import { getITGuides, getITGuide, createITGuide, updateITGuide, deleteITGuide, publishITGuide, unpublishITGuide, uploadITGuideMedia, deleteITGuideMedia, deleteITGuideStep, createITGuideStep, updateITGuideStep, getITGuideAttachments, uploadITGuideAttachment, deleteITGuideAttachment } from '@/api/it_guide'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'ITGuideList',
  components: { DualControlDialog },
  data() {
    return {
      items: [],
      categories: [],
      loading: false,
      filters: { keyword: '', guide_type: '', is_published: '', category: '' },
      // 弹窗
      dialogVisible: false,
      dialogMode: 'create',
      dialogStep: 1,
      editingId: null,
      saving: false,
      form: {
        title: '', description: '', guide_type: '', category: '',
        steps: [], videoDescription: '', videoFile: null, videoUrl: '', _hasNewVideo: false,
        videoMode: 'upload', embedUrl: '', embedInput: ''
      },
      // 拖拽
      dragIndex: -1,
      focusedStepIdx: 0,
      contentLoaded: false,
      _serverImageMap: {},
      // 附件管理弹窗
      attachDialogVisible: false,
      attachGuideId: null,
      attachGuideTitle: '',
      attachList: [],
      attachLoading: false,
      attachLinkForm: { label: '', url: '' },
      attachAddingLink: false,
      attachUploading: false
    }
  },
  computed: {
    attachLinks() { return this.attachList.filter(a => a.attachment_type === 'link') },
    attachFiles() { return this.attachList.filter(a => a.attachment_type === 'file') }
  },
  created() { this.fetchItems() },
  mounted() {
    this._pasteHandler = (e) => this.onGlobalPaste(e)
    document.addEventListener('paste', this._pasteHandler)
  },
  beforeDestroy() {
    document.removeEventListener('paste', this._pasteHandler)
  },
  methods: {
    async fetchItems() {
      this.loading = true
      try {
        const params = {}
        if (this.filters.keyword) params.keyword = this.filters.keyword
        if (this.filters.guide_type) params.guide_type = this.filters.guide_type
        if (this.filters.is_published) params.is_published = this.filters.is_published
        if (this.filters.category) params.category = this.filters.category
        const res = await getITGuides(params)
        this.items = res.data || []
        this.categories = res.categories || []
      } catch (e) { console.error(e) } finally { this.loading = false }
    },
    handleSearch() { this.fetchItems() },
    showCreateDialog() {
      this.dialogMode = 'create'
      this.editingId = null
      this._serverImageMap = {}
      this.resetForm()
      this.dialogVisible = true
      this.dialogStep = 1
    },
    handleEdit(item) {
      this.dialogMode = 'edit'
      this.editingId = item.id
      this._serverImageMap = {}
      this.form.title = item.title
      this.form.description = item.description
      this.form.guide_type = item.guide_type
      this.form.category = item.category
      this.form.steps = []
      this.form.videoDescription = ''
      this.form.videoFile = null
      this.form.videoUrl = ''
      this.form._hasNewVideo = false
      this.form.videoMode = 'upload'
      this.form.embedUrl = ''
      this.form.embedInput = ''
      this.dialogStep = 1
      this.dialogVisible = true
      this.contentLoaded = false
    },
    resetForm() {
      this.form = { title: '', description: '', guide_type: '', category: '', steps: [], videoDescription: '', videoFile: null, videoUrl: '', _hasNewVideo: false, videoMode: 'upload', embedUrl: '', embedInput: '' }
    },
    resetDialog() { this.resetForm(); this.dialogStep = 1; this.contentLoaded = false; this._serverImageMap = {} },
    goToStep2() {
      if (!this.form.title) { this.$message.warning('请输入指南标题'); return }
      if (!this.form.guide_type) { this.$message.warning('请选择指南类型'); return }
      this.dialogStep = 2
      // 编辑模式：加载已有步骤和媒体（仅加载一次，避免覆盖用户修改）
      if (this.dialogMode === 'edit' && this.editingId && !this.contentLoaded) {
        this.loadGuideContent()
      }
    },
    async loadGuideContent() {
      try {
        const res = await getITGuide(this.editingId)
        const { steps, media } = res
        // 构建服务器图片映射（key=url, value={id, url, name}）
        this._serverImageMap = {}
        for (const m of (media || [])) {
          if (m.media_type === 'image') {
            this._serverImageMap[m.file_path] = { id: m.id, url: m.file_path, name: m.file_name }
          }
        }
        if (this.form.guide_type === 'step') {
          this.form.steps = (steps || []).map(s => {
            const stepVideo = (media || []).find(m => m.step_id === s.id && m.media_type === 'video')
            return {
              id: s.id, title: s.title, description: s.description, sort_order: s.sort_order,
              images: (media || []).filter(m => m.step_id === s.id && m.media_type === 'image').map(m => ({ name: m.file_name, url: m.file_path, id: m.id })),
              videoFile: stepVideo && stepVideo.file_path ? stepVideo : null,
              videoUrl: stepVideo?.file_path || '',
              videoMode: stepVideo?.embed_url ? 'embed' : 'upload',
              embedUrl: stepVideo?.embed_url || '',
              embedInput: ''
            }
          })
        } else {
          // 视频指南：用 description 作为指南说明
          this.form.videoDescription = this.form.description || ''
          const guideMedia = (media || []).filter(m => m.step_id === 0)
          const video = guideMedia.find(m => m.media_type === 'video')
          if (video) {
            if (video.embed_url) {
              // 嵌入视频
              this.form.videoMode = 'embed'
              this.form.embedUrl = video.embed_url
              this.form.embedInput = ''
            } else {
              // 文件视频
              this.form.videoFile = { name: video.file_name, id: video.id }
              this.form.videoUrl = video.file_path
              this.form._hasNewVideo = false
            }
          }
        }
      } catch (e) { console.error(e) } finally { this.contentLoaded = true }
    },
    noopUpload() {},
    // 步骤操作
    addStep() { this.form.steps.push({ title: '', description: '', images: [], videoFile: null, videoUrl: '', videoMode: 'upload', embedUrl: '', embedInput: '' }) },
    removeStep(idx) { this.form.steps.splice(idx, 1) },
    // 图片上传
    beforeImageUpload(file) {
      const isImage = file.type.startsWith('image/')
      const isLt5M = file.size / 1024 / 1024 < 5
      if (!isImage) { this.$message.error('只能上传图片文件'); return false }
      if (!isLt5M) { this.$message.error('图片大小不能超过 5MB'); return false }
      return true
    },
    // 恢复服务器图片属性（el-upload change 事件会替换文件对象，丢失 id）
    _restoreServerImageProps(fileList) {
      return fileList.map(f => {
        if (f.url && this._serverImageMap[f.url]) {
          const serverImg = this._serverImageMap[f.url]
          return Object.assign({}, f, { id: serverImg.id })
        }
        return f
      })
    },
    onImageChange(stepIdx, fileList) {
      this.form.steps[stepIdx].images = this._restoreServerImageProps(fileList)
    },
    onImageRemove(stepIdx, fileList) {
      this.form.steps[stepIdx].images = this._restoreServerImageProps(fileList)
    },
    // 全局粘贴监听（弹窗打开时生效）
    onGlobalPaste(e) {
      if (!this.dialogVisible || this.dialogStep !== 2 || this.form.guide_type !== 'step') return
      const items = (e.clipboardData || window.clipboardData).items
      if (!items) return
      for (const item of items) {
        if (item.type.startsWith('image/')) {
          const file = item.getAsFile()
          if (!file) continue
          const isLt5M = file.size / 1024 / 1024 < 5
          if (!isLt5M) { this.$message.error('图片大小不能超过 5MB'); return }
          // 使用最后点击的步骤卡片
          const stepIdx = this.focusedStepIdx
          const url = URL.createObjectURL(file)
          const uid = Date.now() + Math.random()
          // File 对象属性只读，需包装为新对象
          const wrapped = {
            uid: uid,
            name: file.name || `pasted-${uid}.png`,
            status: 'success',
            url: url,
            raw: file
          }
          this.form.steps[stepIdx].images.push(wrapped)
          this.$message.success('图片已粘贴到步骤 ' + (stepIdx + 1))
          break
        }
      }
    },
    // 视频上传
    beforeVideoUpload(file, stepIdx) {
      const isVideo = file.type.startsWith('video/')
      const isLt100M = file.size / 1024 / 1024 < 100
      if (!isVideo) { this.$message.error('只能上传视频文件'); return false }
      if (!isLt100M) { this.$message.error('视频大小不能超过 100MB'); return false }
      this.form.steps[stepIdx].videoFile = file
      this.form.steps[stepIdx].videoUrl = URL.createObjectURL(file)
      return false
    },
    removeStepVideo(idx) {
      if (this.form.steps[idx].videoUrl) URL.revokeObjectURL(this.form.steps[idx].videoUrl)
      this.form.steps[idx].videoFile = null
      this.form.steps[idx].videoUrl = ''
    },
    beforeGuideVideoUpload(file) {
      const isVideo = file.type.startsWith('video/')
      const isLt200M = file.size / 1024 / 1024 < 200
      if (!isVideo) { this.$message.error('只能上传视频文件'); return false }
      if (!isLt200M) { this.$message.error('视频大小不能超过 200MB'); return false }
      this.form.videoFile = file
      this.form.videoUrl = URL.createObjectURL(file)
      this.form._hasNewVideo = true
      return false
    },
    removeGuideVideo() {
      if (this.form.videoUrl) URL.revokeObjectURL(this.form.videoUrl)
      this.form.videoFile = null
      this.form.videoUrl = ''
      this.form._hasNewVideo = false
    },
    // 嵌入视频处理
    _extractEmbedSrc(input) {
      if (!input) return ''
      // 尝试从 iframe 代码中提取 src
      const iframeMatch = input.match(/src=["']([^"']+)["']/)
      if (iframeMatch) {
        const src = iframeMatch[1]
        return src.startsWith('//') ? 'https:' + src : src
      }
      // 如果是直接 URL，确保有协议
      if (input.startsWith('//')) return 'https:' + input
      if (input.startsWith('http')) return input
      // 拒绝非安全协议（如 javascript:、data: 等）
      return ''
    },
    applyStepEmbed(idx) {
      const src = this._extractEmbedSrc(this.form.steps[idx].embedInput)
      if (!src) { this.$message.warning('请输入有效的嵌入代码或URL'); return }
      this.form.steps[idx].embedUrl = src
      this.form.steps[idx].embedInput = ''
    },
    removeStepEmbedVideo(idx) {
      this.form.steps[idx].embedUrl = ''
      this.form.steps[idx].embedInput = ''
    },
    applyGuideEmbed() {
      const src = this._extractEmbedSrc(this.form.embedInput)
      if (!src) { this.$message.warning('请输入有效的嵌入代码或URL'); return }
      this.form.embedUrl = src
      this.form.embedInput = ''
    },
    removeGuideEmbedVideo() {
      this.form.embedUrl = ''
      this.form.embedInput = ''
    },
    // 拖拽排序
    onDragStart(idx) { this.dragIndex = idx },
    onDragOver(idx) { this.dragOverIndex = idx },
    onDrop(idx) {
      if (this.dragIndex === -1 || this.dragIndex === idx) return
      const item = this.form.steps.splice(this.dragIndex, 1)[0]
      this.form.steps.splice(idx, 0, item)
    },
    onDragEnd() { this.dragIndex = -1; this.dragOverIndex = -1 },
    // 保存
    async handleSave() {
      if (this.form.guide_type === 'step' && !this.form.steps.length) {
        this.$message.warning('请至少添加一个步骤'); return
      }
      if (this.form.guide_type === 'video') {
        if (this.form.videoMode === 'upload' && !this.form.videoFile) {
          this.$message.warning('请上传教程视频'); return
        }
        if (this.form.videoMode === 'embed' && !this.form.embedUrl) {
          this.$message.warning('请输入嵌入视频URL'); return
        }
      }
      this.saving = true
      try {
        const dualToken = await this.$refs.dualControl.open()
        // 视频指南：将 videoDescription 同步到 description
        if (this.form.guide_type === 'video') {
          this.form.description = this.form.videoDescription || ''
        }
        let guideId = this.editingId
        if (this.dialogMode === 'create') {
          const res = await createITGuide(this.form, dualToken)
          guideId = res.data.id
          // 创建模式：创建步骤和上传媒体
          if (this.form.guide_type === 'step') {
            for (let i = 0; i < this.form.steps.length; i++) {
              const step = this.form.steps[i]
              const stepRes = await createITGuideStep(guideId, {
                title: step.title, description: step.description, sort_order: i
              }, dualToken)
              const newStepId = stepRes.data.id
              // 上传图片
              for (const img of (step.images || [])) {
                if (img.raw) {
                  const fd = new FormData()
                  fd.append('file', img.raw)
                  fd.append('media_type', 'image')
                  fd.append('step_id', newStepId)
                  await uploadITGuideMedia(guideId, fd, dualToken)
                }
              }
              // 上传视频（按当前模式提交）
              if (step.videoMode === 'upload' && step.videoFile) {
                const fd = new FormData()
                fd.append('file', step.videoFile)
                fd.append('media_type', 'video')
                fd.append('step_id', newStepId)
                await uploadITGuideMedia(guideId, fd, dualToken)
              } else if (step.videoMode === 'embed' && step.embedUrl) {
                const fd = new FormData()
                fd.append('embed_url', step.embedUrl)
                fd.append('media_type', 'video')
                fd.append('step_id', newStepId)
                await uploadITGuideMedia(guideId, fd, dualToken)
              }
            }
          } else {
            // 视频指南：上传视频（按当前模式提交）
            if (this.form.videoMode === 'upload' && this.form.videoFile) {
              const fd = new FormData()
              fd.append('file', this.form.videoFile)
              fd.append('media_type', 'video')
              fd.append('step_id', '0')
              await uploadITGuideMedia(guideId, fd, dualToken)
            } else if (this.form.videoMode === 'embed' && this.form.embedUrl) {
              const fd = new FormData()
              fd.append('embed_url', this.form.embedUrl)
              fd.append('media_type', 'video')
              fd.append('step_id', '0')
              await uploadITGuideMedia(guideId, fd, dualToken)
            }
          }
        } else {
          await updateITGuide(this.editingId, this.form, dualToken)
          // 获取当前服务器上的步骤和媒体
          const old = await getITGuide(this.editingId)
          const oldSteps = old.steps || []
          const oldMedia = old.media || []

          if (this.form.guide_type === 'step') {
            // 构建旧步骤 ID 映射
            const oldStepMap = {}
            for (const s of oldSteps) { oldStepMap[s.id] = s }
            const processedOldStepIds = new Set()

            for (let i = 0; i < this.form.steps.length; i++) {
              const step = this.form.steps[i]
              let stepId

              if (step.id && oldStepMap[step.id]) {
                // 按 ID 匹配已有步骤
                stepId = step.id
                processedOldStepIds.add(stepId)
                await updateITGuideStep(this.editingId, stepId, {
                  title: step.title, description: step.description, sort_order: i
                }, dualToken)
              } else {
                // 新步骤（无 id 或 id 不在旧列表中）
                const stepRes = await createITGuideStep(this.editingId, {
                  title: step.title, description: step.description, sort_order: i
                }, dualToken)
                stepId = stepRes.data.id
              }

              // 找出该步骤在服务器上的图片
              const serverStepImages = oldMedia.filter(m => m.step_id === stepId && m.media_type === 'image')

              // 删除被移除的图片（在服务器列表中但不在当前列表中）
              const currentImageIds = new Set(
                (step.images || []).filter(img => img.id).map(img => img.id)
              )
              for (const m of serverStepImages) {
                if (!currentImageIds.has(m.id)) {
                  try {
                    await deleteITGuideMedia(this.editingId, m.id, dualToken)
                  } catch (e) {
                    console.warn('删除媒体失败:', e)
                  }
                }
              }

              // 只上传新图片（没有 id 属性 = 用户新上传的）
              for (const img of (step.images || [])) {
                if (!img.id && img.raw) {
                  const fd = new FormData()
                  fd.append('file', img.raw)
                  fd.append('media_type', 'image')
                  fd.append('step_id', stepId)
                  await uploadITGuideMedia(this.editingId, fd, dualToken)
                }
              }

              // 处理视频：按当前模式提交
              const oldVideo = oldMedia.find(m => m.step_id === stepId && m.media_type === 'video')
              if (step.videoMode === 'upload' && step.videoFile) {
                // 上传文件视频
                if (oldVideo) {
                  try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
                }
                const fd = new FormData()
                fd.append('file', step.videoFile)
                fd.append('media_type', 'video')
                fd.append('step_id', stepId)
                await uploadITGuideMedia(this.editingId, fd, dualToken)
              } else if (step.videoMode === 'embed' && step.embedUrl) {
                // 嵌入视频
                if (oldVideo) {
                  try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
                }
                const fd = new FormData()
                fd.append('embed_url', step.embedUrl)
                fd.append('media_type', 'video')
                fd.append('step_id', stepId)
                await uploadITGuideMedia(this.editingId, fd, dualToken)
              } else if (step.videoMode === 'upload' && !step.videoFile && !step.videoUrl && oldVideo) {
                // 上传模式下用户移除了视频
                try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
              } else if (step.videoMode === 'embed' && !step.embedUrl && oldVideo) {
                // 嵌入模式下用户移除了视频
                try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
              }
            }

            // 删除未被匹配的旧步骤
            for (const s of oldSteps) {
              if (!processedOldStepIds.has(s.id)) {
                try { await deleteITGuideStep(this.editingId, s.id, dualToken) } catch (e) {}
              }
            }
          } else {
            // 视频指南：处理视频更新（按当前模式提交）
            const oldVideo = oldMedia.find(m => m.media_type === 'video' && m.step_id === 0)
            if (this.form.videoMode === 'upload' && this.form.videoFile && this.form._hasNewVideo) {
              // 上传模式：有新视频文件
              if (oldVideo) {
                try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
              }
              const fd = new FormData()
              fd.append('file', this.form.videoFile)
              fd.append('media_type', 'video')
              fd.append('step_id', '0')
              await uploadITGuideMedia(this.editingId, fd, dualToken)
            } else if (this.form.videoMode === 'embed' && this.form.embedUrl) {
              // 嵌入模式
              if (oldVideo) {
                try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
              }
              const fd = new FormData()
              fd.append('embed_url', this.form.embedUrl)
              fd.append('media_type', 'video')
              fd.append('step_id', '0')
              await uploadITGuideMedia(this.editingId, fd, dualToken)
            } else if (this.form.videoMode === 'upload' && !this.form.videoFile && !this.form.videoUrl && oldVideo) {
              // 上传模式下用户移除了视频
              try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
            } else if (this.form.videoMode === 'embed' && !this.form.embedUrl && oldVideo) {
              // 嵌入模式下用户移除了视频
              try { await deleteITGuideMedia(this.editingId, oldVideo.id, dualToken) } catch (e) {}
            }
          }
        }
        this.$message.success(this.dialogMode === 'create' ? '创建成功' : '保存成功')
        this.dialogVisible = false
        this.fetchItems()
      } catch (e) {
        if (e.message !== 'canceled') {
          console.error(e)
          this.$message.error('保存失败')
        }
      } finally { this.saving = false }
    },
    // 发布/取消发布
    async handleTogglePublish(item) {
      try {
        const dualToken = await this.$refs.dualControl.open()
        if (item.is_published) {
          await unpublishITGuide(item.id, dualToken)
          this.$message.success('已取消发布')
        } else {
          await publishITGuide(item.id, dualToken)
          this.$message.success('发布成功')
        }
        this.fetchItems()
      } catch (e) {
        if (e.message !== 'canceled') this.$message.error('操作失败')
      }
    },
    // 删除
    handleDelete(item) {
      this.$confirm(`确定删除指南“${item.title}”？`, '提示', { type: 'warning' }).then(async () => {
        try {
          const dualToken = await this.$refs.dualControl.open()
          await deleteITGuide(item.id, dualToken)
          this.$message.success('删除成功')
          this.fetchItems()
        } catch (e) {
          if (e.message !== 'canceled') this.$message.error('删除失败')
        }
      }).catch(() => {})
    },
    // ============ 附件管理 ============
    openAttachDialog(item) {
      this.attachGuideId = item.id
      this.attachGuideTitle = item.title
      this.attachList = []
      this.attachLinkForm = { label: '', url: '' }
      this.attachDialogVisible = true
      this.fetchAttachments()
    },
    async fetchAttachments() {
      this.attachLoading = true
      try {
        const res = await getITGuideAttachments(this.attachGuideId)
        this.attachList = res.data || []
      } catch (e) { console.error(e) } finally { this.attachLoading = false }
    },
    async handleAddLink() {
      if (!this.attachLinkForm.label || !this.attachLinkForm.url) {
        this.$message.warning('请输入标签和URL'); return
      }
      if (!this.attachLinkForm.url.startsWith('http://') && !this.attachLinkForm.url.startsWith('https://')) {
        this.$message.warning('URL必须以 http:// 或 https:// 开头'); return
      }
      this.attachAddingLink = true
      try {
        const dualToken = await this.$refs.dualControl.open()
        const fd = new FormData()
        fd.append('attachment_type', 'link')
        fd.append('label', this.attachLinkForm.label)
        fd.append('url', this.attachLinkForm.url)
        await uploadITGuideAttachment(this.attachGuideId, fd, dualToken)
        this.$message.success('链接添加成功')
        this.attachLinkForm = { label: '', url: '' }
        this.fetchAttachments()
      } catch (e) {
        if (e.message !== 'canceled') this.$message.error('添加失败')
      } finally { this.attachAddingLink = false }
    },
    beforeAttachUpload(file) {
      const allowedExts = ['.exe', '.msi', '.zip', '.rar', '.7z', '.tar', '.gz', '.pdf', '.docx', '.xlsx', '.pptx', '.txt', '.bat', '.sh', '.dmg', '.iso']
      const ext = '.' + file.name.split('.').pop().toLowerCase()
      if (!allowedExts.includes(ext)) {
        this.$message.error('不支持的文件格式'); return false
      }
      if (file.size / 1024 / 1024 > 100) {
        this.$message.error('文件大小不能超过 100MB'); return false
      }
      this.doUploadFile(file)
      return false
    },
    async doUploadFile(file) {
      this.attachUploading = true
      try {
        const dualToken = await this.$refs.dualControl.open()
        const fd = new FormData()
        fd.append('attachment_type', 'file')
        fd.append('file', file)
        await uploadITGuideAttachment(this.attachGuideId, fd, dualToken)
        this.$message.success('上传成功')
        this.fetchAttachments()
      } catch (e) {
        // 具体错误信息由请求拦截器统一提示，此处不重复弹窗
        if (e.message !== 'canceled') console.error('上传附件失败:', e)
      } finally { this.attachUploading = false }
    },
    async handleDeleteAttach(item) {
      try {
        const dualToken = await this.$refs.dualControl.open()
        await deleteITGuideAttachment(this.attachGuideId, item.id, dualToken)
        this.$message.success('删除成功')
        this.fetchAttachments()
      } catch (e) {
        if (e.message !== 'canceled') this.$message.error('删除失败')
      }
    },
    formatFileSize(bytes) {
      if (!bytes) return ''
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
    }
  }
}
</script>

<style scoped>
.it-guide-page {
  padding: 20px;
}

.page-card {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #E2E8F0;
}
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 28px; }
.page-title { display: flex; align-items: center; gap: 10px; font-size: 24px; font-weight: 700; color: #1E293B; margin: 0; }
.filter-bar { display: flex; gap: 12px; margin-bottom: 28px; flex-wrap: wrap; }
.search-wrap { width: 140px; flex: none; }
.search-wrap ::v-deep .el-input__inner { border-radius: 8px; height: 40px; line-height: 40px; }
.filter-select { width: 140px; }
.filter-select ::v-deep .el-input__inner { border-radius: 8px; height: 40px; line-height: 40px; }
.page-header .el-button { border-radius: 10px; }
.loading-wrap { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; }
.guide-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 20px; }
.guide-card { background: #fff; border: 1px solid #E2E8F0; border-radius: 12px; padding: 20px; display: flex; flex-direction: column; transition: all 0.25s ease; }
.guide-card:hover { transform: translateY(-3px); box-shadow: 0 8px 24px rgba(0,0,0,0.08), 0 2px 8px rgba(64,158,255,0.06); border-color: #BFDBFE; }
.card-icon { width: 56px; height: 56px; border-radius: 12px; display: flex; align-items: center; justify-content: center; margin-bottom: 14px; }
.icon-step { background: #EFF6FF; color: #2563EB; }
.icon-video { background: #FFF7ED; color: #D97706; }
.card-body { flex: 1; }
.card-title { font-size: 15px; font-weight: 600; color: #1E293B; margin: 0 0 8px 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.card-desc { font-size: 13px; color: #64748B; margin: 0 0 12px 0; line-height: 1.5; overflow: hidden; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; min-height: 40px; }
.card-meta { display: flex; align-items: center; gap: 8px; flex-wrap: wrap; }
.meta-tag { font-size: 11px; padding: 2px 8px; border-radius: 4px; font-weight: 500; }
.tag-step { background: #EFF6FF; color: #2563EB; }
.tag-video { background: #FFF7ED; color: #D97706; }
.tag-category { background: #F1F5F9; color: #64748B; }
.card-footer { margin-top: 16px; padding-top: 14px; border-top: 1px solid #F1F5F9; display: flex; gap: 8px; }
.action-link { font-size: 13px; font-weight: 500; color: #409EFF; text-decoration: none; padding: 6px 14px; border-radius: 6px; background: #EFF6FF; transition: all 0.25s ease; cursor: pointer; }
.action-link:hover { background: #DBEAFE; color: #2563EB; }
.action-danger { color: #DC2626; background: #FEF2F2; }
.action-danger:hover { background: #FEE2E2; color: #B91C1C; }
.empty-state { text-align: center; padding: 80px 20px; }
.empty-state svg { margin-bottom: 16px; }
.empty-title { font-size: 16px; font-weight: 600; color: #64748B; margin: 0 0 6px 0; }
.empty-desc { font-size: 13px; color: #94A3B8; margin: 0; }

/* 弹窗样式 */
.type-selector { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-bottom: 24px; }
.type-card { border: 1px solid #E2E8F0; border-radius: 12px; padding: 20px; text-align: center; cursor: pointer; transition: all 0.25s ease; background: #fff; }
.type-card:hover { border-color: #BFDBFE; }
.type-card.active { border: 2px solid #409EFF; background: #EFF6FF; }
.type-name { font-size: 15px; font-weight: 600; color: #1E293B; margin-top: 12px; }
.type-desc { font-size: 13px; color: #64748B; margin-top: 4px; }
.dialog-form { margin-top: 16px; }
.dialog-footer { display: flex; justify-content: flex-end; gap: 12px; margin-top: 24px; padding-top: 16px; border-top: 1px solid #F1F5F9; }

/* 步骤编辑器 */
.step-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.step-label { font-size: 15px; font-weight: 600; color: #1E293B; }
.add-step-btn { color: #409EFF; border: 1px dashed #BFDBFE; border-radius: 8px; padding: 6px 16px; background: #EFF6FF; }
.add-step-btn:hover { background: #DBEAFE; }
.step-card { display: flex; gap: 12px; border: 1px solid #E2E8F0; border-radius: 12px; padding: 16px; margin-bottom: 12px; background: #fff; transition: border-color 0.2s ease, box-shadow 0.2s ease, opacity 0.15s ease; }
.step-card.is-focused { border-color: #BFDBFE; box-shadow: 0 0 0 3px rgba(64, 158, 255, 0.1); }
.step-card.dragging { opacity: 0.6; }
.step-drag-handle { cursor: grab; display: flex; align-items: flex-start; padding-top: 4px; flex-shrink: 0; }
.step-drag-handle:active { cursor: grabbing; }
.step-content { flex: 1; }
.step-title-row { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.step-number { font-size: 14px; font-weight: 600; color: #2563EB; background: #EFF6FF; padding: 2px 10px; border-radius: 12px; }
.step-delete-btn { color: #94A3B8; }
.step-delete-btn:hover { color: #DC2626; }
.media-section { margin-top: 12px; }
.media-label { font-size: 13px; font-weight: 500; color: #64748B; margin-bottom: 8px; }
.paste-hint { font-size: 11px; color: #94A3B8; font-weight: 400; margin-left: 6px; }
.video-preview-wrap { position: relative; }
.remove-video-btn {
  position: absolute;
  top: 12px;
  right: 12px;
  width: 32px;
  height: 32px;
  padding: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
  background: rgba(0, 0, 0, 0.5);
  color: #fff;
  border: none;
  cursor: pointer;
  transition: all 0.2s ease;
  z-index: 2;
}
.remove-video-btn:hover {
  background: rgba(220, 38, 38, 0.85);
  transform: scale(1.1);
}
.remove-video-btn svg { display: block; }
.empty-steps { text-align: center; padding: 40px 20px; color: #94A3B8; font-size: 14px; }
.video-desc-input ::v-deep .el-textarea__inner { border-radius: 8px; }
.video-drag-upload ::v-deep .el-upload-dragger {
  border-radius: 12px;
  border: 1px dashed #BFDBFE;
  background: #FAFBFF;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 32px 20px;
}
@media (max-width: 640px) { .guide-grid { grid-template-columns: 1fr; } .page-title { font-size: 20px; } .type-selector { grid-template-columns: 1fr; } }

/* 附件管理弹窗 */
.attach-section { margin-bottom: 24px; }
.attach-section:last-child { margin-bottom: 0; }
.attach-section-title { display: flex; align-items: center; gap: 6px; font-size: 14px; font-weight: 600; color: #1E293B; margin-bottom: 12px; }
.attach-link-form { display: flex; gap: 8px; align-items: center; margin-bottom: 12px; }
.attach-list { border: 1px solid #F1F5F9; border-radius: 8px; overflow: hidden; }
.attach-item { display: flex; align-items: center; gap: 8px; padding: 10px 12px; border-bottom: 1px solid #F1F5F9; transition: background 0.15s ease; }
.attach-item:last-child { border-bottom: none; }
.attach-item:hover { background: #F8FAFC; }
.attach-item-label { font-size: 13px; font-weight: 500; color: #1E293B; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 160px; }
.attach-item-url { font-size: 12px; color: #94A3B8; flex: 1; white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.attach-item-size { font-size: 12px; color: #94A3B8; flex-shrink: 0; }
.attach-delete-btn { color: #94A3B8; flex-shrink: 0; }
.attach-delete-btn:hover { color: #DC2626; }
.attach-empty { font-size: 13px; color: #94A3B8; text-align: center; padding: 16px 0; }
.attach-upload ::v-deep .el-upload { width: 100%; }
.attach-upload ::v-deep .el-upload-dragger { width: 100%; border-radius: 12px; border: 1px dashed #BFDBFE; background: #FAFBFF; display: flex; flex-direction: column; align-items: center; justify-content: center; padding: 24px 20px; height: auto; }
.attach-upload ::v-deep .el-upload-dragger:hover { border-color: #409EFF; }
</style>
