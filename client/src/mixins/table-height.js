/**
 * 表格高度动态适配 mixin
 * 用法：
 *   import tableHeightMixin from '@/mixins/table-height'
 *   mixins: [tableHeightMixin]
 *   模板中：el-table :max-height="tableMaxHeight"，外层 .table-card[ref=tableCard] > .table-wrapper
 */
export default {
  data() {
    return {
      tableMaxHeight: 600
    }
  },
  mounted() {
    this.$nextTick(() => this.calcTableHeight())
    window.addEventListener('resize', this.calcTableHeight)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.calcTableHeight)
  },
  methods: {
    calcTableHeight() {
      this.$nextTick(() => {
        const container = this.$el
        if (!container) return
        const containerStyle = getComputedStyle(container)
        const containerH = container.clientHeight
          - parseInt(containerStyle.paddingTop)
          - parseInt(containerStyle.paddingBottom)
        // 通用计算：遍历所有直接子元素，减去非 table-card 的元素高度
        let usedH = 0
        Array.from(container.children).forEach(child => {
          if (child === this.$refs.tableCard) return
          const style = getComputedStyle(child)
          usedH += child.offsetHeight + parseInt(style.marginTop) + parseInt(style.marginBottom)
        })
        const available = containerH - usedH
        // 减去 table-card 自身的 margin（margin 是卡片盒模型的一部分，不是表格可用空间）
        const cardStyle = this.$refs.tableCard ? getComputedStyle(this.$refs.tableCard) : null
        const cardMarginH = cardStyle ? parseInt(cardStyle.marginTop) + parseInt(cardStyle.marginBottom) : 0
        this.tableMaxHeight = Math.max(available - cardMarginH, 200)
        const wrapper = this.$refs.tableCard ? this.$refs.tableCard.querySelector('.table-wrapper') : null
        if (wrapper) {
          wrapper.style.setProperty('--table-body-max-height', (this.tableMaxHeight - 48) + 'px')
        }
      })
    }
  }
}
