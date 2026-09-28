/**
 * 文件预览预检 mixin
 *
 * 背景：@file-viewer/vue2.7 适配器把 coreOptions 硬编码为 { registry }（见
 * node_modules/@file-viewer/vue2.7/dist/index.js 的 viewerCoreOptions），
 * 因此通过 :options 传入的 fetchFile 永远不会被调用，FileViewer 内部始终使用
 * 库自带的 defaultFetchFile 发起裸 fetch。当后端文件被手动删除时，接口返回 404，
 * FileViewer 抛出 "Failed to fetch file: 404 Not Found" 并向上 re-throw，
 * 形成未捕获的 Promise rejection，触发 webpack-dev-server 全屏报错浮层。
 *
 * 用法：
 *   import previewGuardMixin from '@/mixins/preview-guard'
 *   mixins: [previewGuardMixin]
 *
 *   async handlePreview(row) {
 *     const url = getXxxPreviewUrl(row.id)
 *     if (!(await this.checkFileExists(url))) return   // 已弹框提示，阻止打开预览
 *     // ...原有赋值与 previewVisible 打开逻辑
 *   }
 */
export default {
  methods: {
    /**
     * 打开 FileViewer 前探测文件是否可访问
     * @param {string} url 预览接口地址（相对路径亦可）
     * @param {Object} [opts]
     * @param {string} [opts.label] 展示在提示中的文件名，便于定位问题记录
     * @param {boolean} [opts.silent] 只返回结果，不弹出任何提示
     * @returns {Promise<boolean>} true 表示可预览；false 表示已提示并应阻止打开
     */
    async checkFileExists(url, opts = {}) {
      let response
      try {
        response = await fetch(url, { headers: this.buildPreviewHeaders() })
      } catch (e) {
        // 网络异常无法判定文件状态，交由 FileViewer 自行处理
        console.error('预览预检请求失败:', e)
        return true
      }

      if (response.ok) {
        // 响应体被 FileViewer 复用或已锁定时 cancel() 会抛错，必须就地吞掉，否则会冒泡成未捕获拒绝
        try {
          if (response.body) await response.body.cancel()
        } catch (e) { /* ignore */ }
        return true
      }

      if (response.status === 401 || response.status === 403) {
        if (!opts.silent) this.$message.warning('登录状态已失效，请重新登录后再预览')
        return false
      }

      if (response.status === 404) {
        if (!opts.silent) this.showFileMissingAlert(opts.label)
        return false
      }

      if (!opts.silent) this.$message.error('文件预览失败，请稍后重试')
      return false
    },

    showFileMissingAlert(label) {
      const suffix = label ? `（${label}）` : ''
      return this.$alert(
        `文件不存在或已丢失${suffix}，请联系管理员确认文件状态。`,
        '预览失败',
        { confirmButtonText: '确定', type: 'warning' }
      ).catch(() => {})
    },

    /**
     * 预检请求的认证头：Authorization 为主，token 缺失时由后端 access_token Cookie 兜底
     */
    buildPreviewHeaders() {
      const token = localStorage.getItem('token')
      return token ? { 'Authorization': `Bearer ${token}` } : {}
    }
  }
}
