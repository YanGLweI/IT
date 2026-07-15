<template>
  <div class="it-guide-detail-page">
    <!-- 返回链接 -->
    <div class="back-bar">
      <a class="back-link" @click="$router.push('/public/it-guides')">
        <i class="el-icon-arrow-left"></i>
        返回指南列表
      </a>
    </div>

    <!-- 加载中 -->
    <div v-if="loading" class="loading-wrap" v-loading="true" style="min-height: 300px;"></div>

    <template v-else-if="guide">
      <!-- 页面头部 -->
      <div class="detail-header">
        <h1 class="detail-title">{{ guide.title }}</h1>
        <div class="detail-meta">
          <span class="meta-tag" :class="guide.guide_type === 'step' ? 'tag-step' : 'tag-video'">
            {{ guide.guide_type === 'step' ? '步骤指南' : '视频指南' }}
          </span>
          <span class="meta-tag tag-category" v-if="guide.category">{{ guide.category }}</span>
          <span class="meta-date" v-if="guide.published_at">
            发布于 {{ formatDate(guide.published_at) }}
          </span>
        </div>
        <p class="detail-desc" v-if="guide.description">{{ guide.description }}</p>
      </div>

      <!-- 步骤指南 -->
      <div v-if="guide.guide_type === 'step'" class="steps-timeline">
        <div v-for="(step, idx) in steps" :key="step.id" class="timeline-item">
          <div class="timeline-dot">
            <span class="dot-badge">{{ idx + 1 }}</span>
          </div>
          <div class="timeline-content">
            <h3 class="step-title">步骤 {{ idx + 1 }}: {{ step.title }}</h3>
            <p class="step-desc" v-if="step.description">{{ step.description }}</p>

            <!-- 图片走马灯 -->
            <el-carousel v-if="getStepImages(step.id).length > 1" :interval="0" trigger="click" class="step-carousel" arrow="always" indicator-position="bottom">
              <el-carousel-item v-for="img in getStepImages(step.id)" :key="img.id">
                <div class="carousel-image-wrap" @click="previewImage(img)">
                  <img :src="getFileUrl(img.file_path)" :alt="img.file_name" class="carousel-image" loading="lazy" />
                </div>
              </el-carousel-item>
            </el-carousel>
            <!-- 单张图片直接展示 -->
            <div v-else-if="getStepImages(step.id).length === 1" class="step-images">
              <div class="step-image-wrap" @click="previewImage(getStepImages(step.id)[0])">
                <img :src="getFileUrl(getStepImages(step.id)[0].file_path)" :alt="getStepImages(step.id)[0].file_name" class="step-image" loading="lazy" />
              </div>
            </div>

            <!-- 步骤视频 -->
            <div v-if="getStepVideo(step.id)" class="step-video-wrap">
              <video controls preload="metadata" :src="getFileUrl(getStepVideo(step.id).file_path)" style="width: 100%; border-radius: 12px; border: 1px solid #E2E8F0;"></video>
            </div>
          </div>
        </div>
      </div>

      <!-- 视频指南 -->
      <div v-if="guide.guide_type === 'video'" class="video-guide">
        <div class="video-guide-text" v-if="guide.description">
          <p style="white-space: pre-wrap;">{{ guide.description }}</p>
        </div>
        <div v-if="guideVideo" class="video-player-wrap">
          <video controls preload="metadata" :src="getFileUrl(guideVideo.file_path)" style="width: 100%; border-radius: 12px; border: 1px solid #E2E8F0;"></video>
        </div>
      </div>
    </template>

    <!-- 未找到 -->
    <div v-else class="empty-state">
      <p class="empty-title">指南不存在或已下架</p>
      <a class="back-link-inline" @click="$router.push('/public/it-guides')">返回指南列表</a>
    </div>

    <!-- 图片全屏预览 -->
    <div v-if="imageViewer.visible" class="image-viewer-overlay" @click="imageViewer.visible = false">
      <img :src="getFileUrl(imageViewer.url)" class="viewer-image" />
    </div>
  </div>
</template>

<script>
import { getPublicITGuideDetail } from '@/api/public_form'

export default {
  name: 'PublicITGuideDetail',
  data() {
    return {
      guide: null,
      steps: [],
      media: [],
      loading: false,
      imageViewer: { visible: false, url: '' }
    }
  },
  computed: {
    guideVideo() {
      return this.media.find(m => m.step_id === 0 && m.media_type === 'video') || null
    }
  },
  created() {
    this.fetchDetail()
  },
  methods: {
    async fetchDetail() {
      this.loading = true
      try {
        const res = await getPublicITGuideDetail(this.$route.params.id)
        const body = res.data
        this.guide = body.data
        this.steps = body.steps || []
        // 媒体数据：优先从步骤嵌套中提取，兼容顶层 media
        const stepMedia = []
        ;(body.steps || []).forEach(s => {
          if (s.media && s.media.length) stepMedia.push(...s.media)
        })
        this.media = stepMedia.length ? stepMedia : (body.media || [])
      } catch (e) {
        console.error('获取指南详情失败:', e)
        this.guide = null
      } finally {
        this.loading = false
      }
    },
    getStepImages(stepId) {
      return this.media.filter(m => m.step_id === stepId && m.media_type === 'image')
    },
    getStepVideo(stepId) {
      return this.media.find(m => m.step_id === stepId && m.media_type === 'video') || null
    },
    getFileUrl(path) {
      if (!path) return ''
      return path.startsWith('/') ? path : '/' + path
    },
    previewImage(img) {
      this.imageViewer.url = img.file_path
      this.imageViewer.visible = true
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      const d = new Date(dateStr)
      return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
    }
  }
}
</script>

<style scoped>
.it-guide-detail-page {
  max-width: 960px;
  margin: 0 auto;
}

/* 返回链接 */
.back-bar {
  margin-bottom: 20px;
}

.back-link {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  font-size: 13px;
  color: #64748B;
  text-decoration: none;
  padding: 6px 12px;
  border-radius: 6px;
  transition: all 0.25s ease;
  cursor: pointer;
}

.back-link:hover {
  color: #409EFF;
  background: #EFF6FF;
}

/* 页面头部 */
.detail-header {
  margin-bottom: 32px;
}

.detail-title {
  font-size: 24px;
  font-weight: 700;
  color: #1E293B;
  margin: 0 0 12px 0;
  line-height: 1.3;
}

.detail-meta {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.meta-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  font-weight: 500;
}

.tag-step {
  background: #EFF6FF;
  color: #2563EB;
}

.tag-video {
  background: #FFF7ED;
  color: #D97706;
}

.tag-category {
  background: #F1F5F9;
  color: #64748B;
}

.meta-date {
  font-size: 12px;
  color: #94A3B8;
}

.detail-desc {
  font-size: 14px;
  color: #64748B;
  line-height: 1.6;
  margin: 0;
}

/* 步骤时间线 */
.steps-timeline {
  position: relative;
  padding-left: 20px;
}

.steps-timeline::before {
  content: '';
  position: absolute;
  left: 5px;
  top: 14px;
  bottom: 14px;
  width: 2px;
  background: #E2E8F0;
  border-radius: 1px;
}

.timeline-item {
  position: relative;
  padding-left: 36px;
  padding-bottom: 32px;
}

.timeline-item:last-child {
  padding-bottom: 0;
}

.timeline-dot {
  position: absolute;
  left: -1px;
  top: 4px;
}

.dot-badge {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: #EFF6FF;
  color: #2563EB;
  font-size: 13px;
  font-weight: 600;
  border: 2px solid #fff;
  box-shadow: 0 0 0 2px #409EFF;
}

.timeline-content {
  background: #fff;
  border: 1px solid #E2E8F0;
  border-radius: 12px;
  padding: 20px;
}

.step-title {
  font-size: 16px;
  font-weight: 600;
  color: #1E293B;
  margin: 0 0 8px 0;
}

.step-desc {
  font-size: 14px;
  line-height: 1.8;
  color: #64748B;
  margin: 0 0 16px 0;
  white-space: pre-wrap;
}

/* 图片走马灯 */
.step-carousel {
  margin-top: 12px;
  border-radius: 12px;
  border: 1px solid #E2E8F0;
  overflow: hidden;
}
.step-carousel .el-carousel__container { height: auto; }
.step-carousel .el-carousel__item { display: flex; align-items: center; justify-content: center; background: #F8FAFC; }
.carousel-image-wrap { width: 100%; cursor: pointer; }
.carousel-image { width: 100%; height: auto; display: block; }

/* 单张图片 */
.step-images {
  display: grid;
  grid-template-columns: 1fr;
  gap: 12px;
  margin-top: 12px;
}

.step-image-wrap {
  border-radius: 12px;
  border: 1px solid #E2E8F0;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.25s ease;
}

.step-image-wrap:hover {
  transform: scale(1.02);
  border-color: #BFDBFE;
}

.step-image {
  width: 100%;
  height: auto;
  display: block;
  object-fit: cover;
}

/* 步骤视频 */
.step-video-wrap {
  margin-top: 16px;
  background: #F1F5F9;
  border-radius: 12px;
  padding: 8px;
}

/* 视频指南 */
.video-guide-text {
  max-width: 800px;
  margin: 0 auto 32px auto;
  font-size: 14px;
  line-height: 1.8;
  color: #475569;
}

.video-guide-text p {
  margin: 0;
  white-space: pre-wrap;
}

.video-player-wrap {
  max-width: 960px;
  margin: 0 auto;
  background: #F1F5F9;
  border-radius: 12px;
  padding: 8px;
}

/* 图片查看器 */
.image-viewer-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.8);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 9999;
  cursor: pointer;
}

.viewer-image {
  max-width: 90vw;
  max-height: 90vh;
  border-radius: 8px;
  object-fit: contain;
}

/* 空状态 */
.empty-state {
  text-align: center;
  padding: 80px 20px;
}

.empty-title {
  font-size: 16px;
  font-weight: 600;
  color: #64748B;
  margin: 0 0 12px 0;
}

.back-link-inline {
  font-size: 13px;
  color: #409EFF;
  cursor: pointer;
  text-decoration: underline;
}

.loading-wrap {
  min-height: 300px;
}

@media (max-width: 640px) {
  .it-guide-detail-page {
    max-width: 100%;
  }
  .detail-title {
    font-size: 20px;
  }
  .step-carousel {
    border-radius: 8px;
  }
  .timeline-item {
    padding-left: 28px;
  }
}
</style>
