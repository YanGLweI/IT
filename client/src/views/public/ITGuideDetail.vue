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
        <!-- 点赞按钮 -->
        <button class="like-btn" :class="{ liked: isLiked }"
                ref="likeBtn"
                @click="handleLike"
                @mouseenter="onLikeBtnEnter"
                @mousemove="onLikeBtnMove"
                @mouseleave="onLikeBtnLeave">
          <span class="leftContainer">
            <svg viewBox="0 0 512 512" xmlns="http://www.w3.org/2000/svg"><path d="M47.6 300.4L228.3 469.1c7.5 7 17.4 10.9 27.7 10.9s20.2-3.9 27.7-10.9L464.4 300.4c30.4-28.3 47.6-68 47.6-109.5v-5.8c0-69.9-50.5-129.5-119.4-141C347 36.5 300.6 51.4 268 84L256 96 244 84c-32.6-32.6-79-47.5-124.6-39.9C50.5 55.6 0 115.2 0 185.1v5.8c0 41.5 17.2 81.2 47.6 109.5z"></path></svg>
            <span class="likeText">Like</span>
          </span>
          <span class="likeCount"><span class="likeCountNum" :key="likeAnimKey">{{ formatLikeCount(likeCount) }}</span></span>
        </button>
        <div class="detail-meta">
          <span class="meta-tag" :class="guide.guide_type === 'step' ? 'tag-step' : 'tag-video'">
            {{ guide.guide_type === 'step' ? '步骤指南' : '视频指南' }}
          </span>
          <span class="meta-tag tag-category" v-if="guide.category">{{ guide.category }}</span>
          <span class="meta-date" v-if="guide.published_at">
            发布于 {{ formatDate(guide.published_at) }}
          </span>
        </div>
        <p class="detail-desc" v-if="guide.description && guide.guide_type === 'step'">{{ guide.description }}</p>
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
            <el-carousel v-if="getStepImages(step.id).length > 1" :interval="0" trigger="click" indicator-position="outside">
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
              <iframe v-if="getStepVideo(step.id).embed_url" :src="getStepVideo(step.id).embed_url" scrolling="no" border="0" frameborder="0" allowfullscreen="true" sandbox="allow-scripts allow-same-origin allow-presentation" class="step-video-embed"></iframe>
              <video v-else controls preload="metadata" :src="getFileUrl(getStepVideo(step.id).file_path)" class="step-video" loading="lazy"></video>
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
          <iframe v-if="guideVideo.embed_url" :src="guideVideo.embed_url" scrolling="no" border="0" frameborder="0" allowfullscreen="true" sandbox="allow-scripts allow-same-origin allow-presentation" class="guide-video-embed"></iframe>
          <video v-else controls preload="metadata" :src="getFileUrl(guideVideo.file_path)" class="guide-video" loading="lazy"></video>
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
import { getPublicITGuideDetail, recordITGuideView, toggleITGuideLike } from '@/api/public_form'
import { animate, utils } from 'animejs'

export default {
  name: 'PublicITGuideDetail',
  data() {
    return {
      guide: null,
      steps: [],
      media: [],
      loading: false,
      imageViewer: { visible: false, url: '' },
      isLiked: false,
      likeCount: 0,
      likeAnimKey: 0,
      _likeCooldown: false,
      heartParticles: [],
      heartPoolSize: 24,
      heartTimer: null,
      iconAnim: null,
      pointerX: 0,
      pointerY: 0,
      heartSpread: 16
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
  mounted() {
    if (this.guide) this.initHeartParticles()
  },
  beforeDestroy() {
    this.cleanupHeartParticles()
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
        this.likeCount = body.data.like_count || 0
        // 记录浏览量
        try { await recordITGuideView(this.$route.params.id) } catch (e) {}
      } catch (e) {
        console.error('获取指南详情失败:', e)
        this.guide = null
      } finally {
        this.loading = false
        this.$nextTick(() => { this.initHeartParticles() })
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
      if (!path.includes('uploads/it_guide_media/')) return ''
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
    },
    formatLikeCount(n) {
      if (!n) return '0'
      if (n >= 10000) return (n / 10000).toFixed(1) + 'W'
      if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
      return String(n)
    },
    async handleLike() {
      if (this._likeCooldown) return
      this._likeCooldown = true
      setTimeout(() => { this._likeCooldown = false }, 500)
      const nextLiked = !this.isLiked
      try {
        const res = await toggleITGuideLike(this.$route.params.id, { liked: nextLiked })
        const data = res.data.data
        this.isLiked = data.liked
        this.likeCount = data.like_count
        this.likeAnimKey++
      } catch (e) {
        console.error('点赞操作失败:', e)
      }
    },
    initHeartParticles() {
      if (this.heartParticles.length > 0) return
      const btn = this.$refs.likeBtn
      if (!btn) return
      const svg = btn.querySelector('svg')
      if (!svg) return
      for (let i = 0; i < this.heartPoolSize; i++) {
        const p = svg.cloneNode(true)
        p.style.cssText = 'position:absolute;top:0;left:0;mix-blend-mode:plus-lighter;visibility:hidden;pointer-events:none;z-index:10;'
        p.classList.add('heart-particle')
        btn.appendChild(p)
        this.heartParticles.push({ el: p, inUse: false })
      }
    },
    getHeartParticle() {
      const p = this.heartParticles.find(x => !x.inUse)
      if (p) { p.inUse = true; p.el.style.visibility = 'visible'; return p }
      return null
    },
    releaseHeartParticle(p) {
      p.inUse = false
      p.el.style.visibility = 'hidden'
    },
    emitHeartParticle() {
      const p = this.getHeartParticle()
      if (!p) return
      const prefixA = utils.randomPick(['-=', '+='])
      const prefixB = prefixA === '-=' ? '+=' : '-='
      animate(p.el, {
        translateX: [
          this.pointerX + utils.random(-this.heartSpread, this.heartSpread),
          prefixA + (5 + utils.random(-2, 2, 2)),
          prefixB + (6 + utils.random(-2, 2, 2)),
          prefixA + (4 + utils.random(-2, 2, 2))
        ],
        translateY: [
          { from: this.pointerY + utils.random(-5, 5), to: '-=' + utils.random(30, 50) }
        ],
        scale: [{ from: 0, to: 0.85 }, { to: 0 }],
        opacity: [{ from: 0, to: 1, duration: 150 }, { to: 0 }],
        duration: 1200,
        easing: 'easeOutQuad',
        onComplete: () => this.releaseHeartParticle(p)
      })
    },
    onLikeBtnEnter(e) {
      this.updatePointer(e)
      const svg = this.$refs.likeBtn.querySelector('svg')
      if (this.iconAnim) this.iconAnim.revert()
      this.iconAnim = animate(svg, {
        scale: [1, 1.25, 1],
        loop: true,
        duration: 900
      })
      this.heartTimer = setInterval(() => this.emitHeartParticle(), 100)
    },
    onLikeBtnMove(e) {
      this.updatePointer(e)
    },
    onLikeBtnLeave() {
      clearInterval(this.heartTimer)
      this.heartTimer = null
      const svg = this.$refs.likeBtn.querySelector('svg')
      if (this.iconAnim) this.iconAnim.revert()
      this.iconAnim = animate(svg, { scale: 1, duration: 500 })
    },
    updatePointer(e) {
      const rect = this.$refs.likeBtn.getBoundingClientRect()
      this.pointerX = utils.clamp(e.clientX - rect.left, this.heartSpread, rect.width - this.heartSpread)
      this.pointerY = utils.clamp(e.clientY - rect.top, 0, rect.height)
    },
    cleanupHeartParticles() {
      clearInterval(this.heartTimer)
      this.heartParticles.forEach(p => p.el.remove())
      this.heartParticles = []
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

/* 点赞按钮 */
.like-btn {
  position: absolute;
  top: 0;
  right: 0;
  width: 140px;
  height: 35px;
  display: flex;
  align-items: center;
  justify-content: flex-start;
  border: none;
  border-radius: 18px;
  overflow: visible;
  box-shadow: 5px 5px 10px rgba(0, 0, 0, 0.089);
  cursor: pointer;
  background-color: transparent;
  padding: 0;
}

.like-btn .leftContainer {
  width: 60%;
  height: 100%;
  background-color: #94A3B8;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  border-radius: 18px 0 0 18px;
  transition: background-color 0.25s ease, box-shadow 0.25s ease;
}

.like-btn.liked .leftContainer {
  background-color: #EE0000;
}

.like-btn .leftContainer svg {
  width: 1em;
  height: 1em;
  fill: white;
  transition: transform 0.2s ease, filter 0.25s ease;
}

.like-btn .likeText {
  color: white;
  font-weight: 600;
  font-size: 13px;
  transition: text-shadow 0.25s ease;
}

.like-btn .likeCount {
  width: 40%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #94A3B8;
  font-weight: 600;
  font-size: 13px;
  position: relative;
  background-color: white;
  border-radius: 0 18px 18px 0;
  transition: color 0.25s ease;
}

.like-btn .likeCountNum {
  animation: likeCountRise 0.35s ease-out;
}

@keyframes likeCountRise {
  0% {
    opacity: 0;
    transform: translateY(8px);
  }
  100% {
    opacity: 1;
    transform: translateY(0);
  }
}

.like-btn.liked .likeCount {
  color: #EE0000;
}

.like-btn .likeCount::before {
  height: 8px;
  width: 8px;
  position: absolute;
  content: "";
  background-color: white;
  transform: rotate(45deg);
  left: -4px;
}

.like-btn:hover:not(.liked) .leftContainer {
  background-color: #C0392B;
  box-shadow: 0 0 12px rgba(192, 57, 43, 0.5);
}

.like-btn:hover:not(.liked) .leftContainer svg {
  filter: drop-shadow(0 0 4px rgba(255, 107, 129, 0.8));
}

.like-btn:hover:not(.liked) .likeText {
  text-shadow: 0 0 6px rgba(255, 107, 129, 0.7);
}

.like-btn:active .leftContainer svg {
  transform: scale(1.15);
  transform-origin: top;
}

.heart-particle {
  width: 14px;
  height: 14px;
  fill: #FF6B81;
  filter: drop-shadow(0 0 4px rgba(255, 107, 129, 0.8));
}

/* 页面头部 */
.detail-header {
  margin-bottom: 32px;
  position: relative;
}

.detail-title {
  font-size: 24px;
  font-weight: 700;
  color: #1E293B;
  margin: 0 0 12px 0;
  line-height: 1.3;
  padding-right: 160px;
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
.el-carousel__container { height: auto; }
.el-carousel__item { display: flex; align-items: center; justify-content: center; background: #F8FAFC; }
.carousel-image-wrap {
  width: 100%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
}
.carousel-image { max-width: 100%; max-height: 480px; width: auto; height: auto; display: block; object-fit: contain; }

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
  display: flex;
  align-items: center;
  justify-content: center;
  max-height: 480px;
  background: #F8FAFC;
}

.step-image-wrap:hover {
  transform: scale(1.02);
  border-color: #BFDBFE;
}

.step-image {
  max-width: 100%;
  max-height: 480px;
  width: auto;
  height: auto;
  display: block;
  object-fit: contain;
}

/* 步骤视频 */
.step-video-wrap {
  margin-top: 16px;
  background: #F1F5F9;
  border-radius: 12px;
  padding: 8px;
}
.step-video {
  width: 100%;
  max-height: 480px;
  border-radius: 12px;
  border: 1px solid #E2E8F0;
  display: block;
  object-fit: contain;
}
.step-video-embed {
  width: 100%;
  height: 480px;
  border-radius: 12px;
  border: none;
  display: block;
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
.guide-video {
  width: 100%;
  max-height: 540px;
  border-radius: 12px;
  border: 1px solid #E2E8F0;
  display: block;
  object-fit: contain;
}
.guide-video-embed {
  width: 100%;
  height: 540px;
  border-radius: 12px;
  border: none;
  display: block;
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
