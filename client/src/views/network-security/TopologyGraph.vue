<template>
  <div class="topology-graph" ref="pageRoot">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">拓扑图生成</h2>
        <p class="page-subtitle">根据防火墙树结构与区域挂靠关系自动生成网络拓扑图</p>
      </div>
      <div class="header-actions">
        <el-button size="small" icon="el-icon-rank" @click="expandAll">全部展开</el-button>
        <el-button size="small" icon="el-icon-refresh-left" @click="resetView">重置视图</el-button>
        <el-button size="small" icon="el-icon-refresh" @click="fetchAll" :loading="loading">刷新</el-button>
        <el-button size="small" icon="el-icon-download" @click="exportPDF" :loading="exporting">导出PDF</el-button>
      </div>
    </div>

    <!-- 图例 -->
    <div class="legend-bar" v-if="hasData">
      <span class="legend-item"><i class="legend-dot" style="background: #ffffff; border: 1.5px solid #334155"></i>互联网</span>
      <span class="legend-item"><i class="legend-dot" style="background: #f87171"></i>防火墙</span>
      <span class="legend-item"><i class="legend-dot" style="background: #86efac"></i>二级区域</span>
      <span class="legend-item"><i class="legend-dot" style="background: #fde047"></i>三级区域</span>
      <span class="legend-item"><i class="legend-dot" style="background: #fdba74"></i>四级区域</span>
      <span class="legend-item"><i class="legend-dot" style="background: #f87171"></i>高安全区域</span>
      <span class="legend-item"><i class="legend-dot" style="background: #64748b"></i>Linux资产</span>
      <span class="legend-item"><i class="legend-dot" style="background: #93c5fd"></i>Windows资产</span>
      <span class="legend-item"><i class="legend-dot" style="background: #d4a574"></i>其他资产</span>
      <span class="legend-tip">提示：点击区域节点可展开/收起其下资产，点击防火墙节点折叠子树，右侧按钮或滚轮缩放、拖动平移</span>
    </div>

    <div class="chart-wrap" v-show="hasData">
      <div ref="chartRef" class="chart-container"></div>
      <!-- 缩放控制：模拟滚轮事件，与滚轮缩放完全一致的整体缩放 -->
      <div class="zoom-controls">
        <el-tooltip content="放大" placement="left">
          <button class="zoom-btn" @click="zoomIn">+</button>
        </el-tooltip>
        <el-tooltip content="缩小" placement="left">
          <button class="zoom-btn" @click="zoomOut">−</button>
        </el-tooltip>
      </div>
    </div>
    <el-empty v-if="!loading && !hasData" description="暂无拓扑数据，请先在《拓扑关系配置》中配置防火墙与区域关系" />
  </div>
</template>

<script>
import * as echarts from 'echarts'
import { jsPDF } from 'jspdf'
import { getFirewallTopologyTree } from '@/api/firewall_topology'

// 互联网云朵形状：经典三瓣云（顶部大包 + 左右两包），Bootstrap Icons cloud 同款 path（16x16 viewBox）
const CLOUD_SYMBOL = 'path://M4.406 3.342A5.53 5.53 0 0 1 8 2c2.69 0 4.923 2 5.166 4.579C14.758 6.804 16 8.137 16 9.773 16 11.569 14.502 13 12.687 13H3.781C1.708 13 0 11.366 0 9.318c0-1.763 1.266-3.223 2.942-3.593.143-.863.698-1.723 1.464-2.383z'
// 防火墙砖墙图标：实心圆角方块（保证图标整体可点击展开/收起），砖缝以镂空方式露出页面白底（24x24 viewBox）
const FIREWALL_SYMBOL = 'path://M5 3h14a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2zM5 10h14v-2H5zM5 16h14v-2H5zM9 8h2V5H9zM15 14h2v-3h-2zM9 19h2v-3H9z'

export default {
  name: 'TopologyGraph',
  data() {
    return {
      loading: false,
      exporting: false,
      nodes: [],
      chart: null
    }
  },
  computed: {
    hasData() {
      return this.nodes.some(n => !n.invalid)
    }
  },
  mounted() {
    this.fetchAll()
    this.bindExternalRoam()
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    this.unbindExternalRoam()
    window.removeEventListener('resize', this.handleResize)
    if (this.chart) {
      this.chart.dispose()
      this.chart = null
    }
  },
  methods: {
    async fetchAll() {
      this.loading = true
      try {
        const res = await getFirewallTopologyTree()
        this.nodes = res.data || []
        if (this.hasData) {
          this.$nextTick(() => this.renderChart())
        }
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    // 组装 ECharts 树数据：互联网 -> 防火墙树 -> 区域 -> 资产
    buildTreeData() {
      const fwNode = (n) => {
        const name = n.asset.computer_name
        const ip = n.asset.ip_address
        const children = []
        // 挂靠区域
        ;(n.regions || []).forEach(r => {
          children.push({
            // 稳定 id 供 ECharts 数据 diff 精确匹配，避免更新时重建节点导致抖动
            id: 'region:' + r.id,
            name: r.name,
            nodeType: 'region',
            networkLevel: r.network_level || 0,
            tooltip: `区域：${r.name}<br/>资产数：${(r.assets || []).length}`,
            children: (r.assets || []).map(a => ({
              id: 'asset:' + a.id,
              name: a.computer_name,
              nodeType: 'asset',
              osName: a.os_name || '',
              tooltip: `资产：${a.computer_name}<br/>IP：${a.ip_address || '无'}`,
              children: []
            }))
          })
        })
        // 下级防火墙
        this.nodes
          .filter(c => c.parent_id === n.id && !c.invalid)
          .sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
          .forEach(c => children.push(fwNode(c)))
        return {
          id: 'fw:' + n.id,
          name,
          nodeType: 'firewall',
          tooltip: `防火墙：${name}<br/>IP：${ip || '无'}${n.remark ? '<br/>备注：' + n.remark : ''}`,
          children
        }
      }
      const roots = this.nodes
        .filter(n => n.parent_id === 0 && !n.invalid)
        .sort((a, b) => a.sort_order - b.sort_order || a.id - b.id)
        .map(fwNode)
      return {
        id: 'internet',
        name: '互联网',
        nodeType: 'internet',
        tooltip: '互联网（虚拟根节点）',
        children: roots
      }
    },
    // 根据节点类型设置外观（自定义形状 + 胶囊标签，保证文字与背景对比度）
    decorateNode(data) {
      if (data.nodeType === 'internet') {
        // 互联网：经典描边云朵（白底 + 深色轮廓），文字在云内
        data.symbol = CLOUD_SYMBOL
        data.symbolSize = [84, 60]
        data.itemStyle = { color: '#ffffff', borderColor: '#334155', borderWidth: 2 }
        data.label = {
          position: 'inside',
          color: '#334155',
          fontSize: 13,
          fontWeight: 600
        }
      } else if (data.nodeType === 'firewall') {
        // 防火墙：实心砖墙图标（淡红色，镂空砖缝），整块图标均可点击，淡红胶囊标签
        data.symbol = FIREWALL_SYMBOL
        data.symbolSize = [38, 38]
        data.itemStyle = { color: '#f87171', borderColor: '#ef4444' }
        data.label = {
          position: 'left',
          distance: 8,
          color: '#b91c1c',
          fontSize: 12,
          fontWeight: 600,
          backgroundColor: '#fee2e2',
          borderColor: '#fca5a5',
          borderWidth: 1,
          borderRadius: 4,
          padding: [4, 8]
        }
      } else if (data.nodeType === 'region') {
        // 区域：根据网络等级设置不同配色（二级淡绿、三级淡黄、四级淡橙、五级淡红）
        const levelColors = {
          2: { dot: '#86efac', border: '#4ade80', label: '#14532d', bg: '#dcfce7', labelBorder: '#86efac' },
          3: { dot: '#fde047', border: '#facc15', label: '#713f12', bg: '#fef9c3', labelBorder: '#fde047' },
          4: { dot: '#fdba74', border: '#fb923c', label: '#7c2d12', bg: '#ffedd5', labelBorder: '#fdba74' },
          5: { dot: '#f87171', border: '#ef4444', label: '#b91c1c', bg: '#fee2e2', labelBorder: '#fca5a5' }
        }
        const colors = levelColors[data.networkLevel] || levelColors[2]
        data.symbolSize = 22
        data.itemStyle = { color: colors.dot, borderColor: colors.border }
        data.label = {
          position: 'left',
          distance: 6,
          color: colors.label,
          fontSize: 12,
          fontWeight: 500,
          backgroundColor: colors.bg,
          borderColor: colors.labelBorder,
          borderWidth: 1,
          borderRadius: 4,
          padding: [3, 6]
        }
      } else {
        // 资产：根据操作系统类型设置不同颜色（RHEL=Linux 灰色、Windows 淡蓝色、其他淡灰色）
        const osName = (data.osName || '').toLowerCase()
        let assetColor = '#d4a574' // 默认淡棕色（其他）
        if (osName.startsWith('rhel')) {
          assetColor = '#64748b' // Linux 灰色
        } else if (osName.startsWith('windows')) {
          assetColor = '#93c5fd' // Windows 淡蓝色
        }
        data.symbolSize = 14
        data.itemStyle = { color: assetColor, borderColor: assetColor }
        data.label = {
          position: 'right',
          distance: 5,
          color: '#1e293b',
          fontSize: 10,
          lineHeight: 14,
          align: 'left',
          backgroundColor: 'rgba(241, 245, 249, 0.95)',
          borderColor: '#e2e8f0',
          borderWidth: 1,
          borderRadius: 3,
          padding: [2, 4],
          formatter: () => data.name
        }
        const m = data.tooltip && data.tooltip.match(/IP：(.+?)(<br\/\>|$)/)
        if (m && m[1] !== '无') {
          data.label.formatter = () => `${data.name}  ${m[1]}`
        }
      }
      ;(data.children || []).forEach(c => this.decorateNode(c))
      return data
    },
    // 导出拓扑图为 PDF（离屏渲染高清大图，确保全部资产完整显示不遮挡）
    async exportPDF() {
      if (!this.treeData || this.exporting) return
      this.exporting = true
      try {
        // 创建离屏容器
        const container = document.createElement('div')
        container.style.cssText = 'position:fixed;left:-9999px;top:0;width:3200px;height:2200px;'
        document.body.appendChild(container)

        // 初始化临时 ECharts 实例
        const tmpChart = echarts.init(container, null, { renderer: 'canvas' })

        // 深拷贝树数据并重新 decorate（避免修改主图表数据）
        const exportData = JSON.parse(JSON.stringify(this.treeData))
        this.decorateNode(exportData)

        // 渲染配置：展开全部层级、关闭动画、放大符号与字体
        tmpChart.setOption({
          graphic: {
            elements: [
              {
                type: 'group',
                left: 'center',
                top: 20,
                children: [
                  { type: 'rect', shape: { x: 0, y: 0, width: 12, height: 12 }, style: { fill: '#ffffff', stroke: '#334155', lineWidth: 1.5 } },
                  { type: 'text', style: { text: '互联网', x: 18, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 70, y: 0, width: 12, height: 12 }, style: { fill: '#f87171' } },
                  { type: 'text', style: { text: '防火墙', x: 88, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 140, y: 0, width: 12, height: 12 }, style: { fill: '#86efac' } },
                  { type: 'text', style: { text: '二级区域', x: 158, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 210, y: 0, width: 12, height: 12 }, style: { fill: '#fde047' } },
                  { type: 'text', style: { text: '三级区域', x: 228, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 280, y: 0, width: 12, height: 12 }, style: { fill: '#fdba74' } },
                  { type: 'text', style: { text: '四级区域', x: 298, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 350, y: 0, width: 12, height: 12 }, style: { fill: '#f87171' } },
                  { type: 'text', style: { text: '高安全区域', x: 368, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 440, y: 0, width: 12, height: 12 }, style: { fill: '#64748b' } },
                  { type: 'text', style: { text: 'Linux 资产', x: 458, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 530, y: 0, width: 12, height: 12 }, style: { fill: '#93c5fd' } },
                  { type: 'text', style: { text: 'Windows 资产', x: 548, y: 10, font: '12px sans-serif', fill: '#64748b' } },
                  { type: 'rect', shape: { x: 640, y: 0, width: 12, height: 12 }, style: { fill: '#d4a574' } },
                  { type: 'text', style: { text: '其他资产', x: 658, y: 10, font: '12px sans-serif', fill: '#64748b' } }
                ]
              }
            ]
          },
          series: [{
            type: 'tree',
            data: [exportData],
            orient: 'LR',
            top: '12%',
            bottom: '5%',
            left: '5%',
            right: '10%',
            roam: false,
            expandAndCollapse: false,
            initialTreeDepth: 10,
            symbol: 'circle',
            lineStyle: { color: '#cbd5e1', width: 1.5, curveness: 0.3 },
            emphasis: { focus: 'descendant' },
            animation: false
          }]
        }, true)

        // 等待渲染完成
        await new Promise(resolve => setTimeout(resolve, 500))

        // 获取高清 PNG（pixelRatio: 1.5 平衡清晰度与文件大小）
        const imgData = tmpChart.getDataURL({ type: 'png', pixelRatio: 1.5, backgroundColor: '#fff' })

        // 销毁临时图表和容器
        tmpChart.dispose()
        document.body.removeChild(container)

        // 创建横向 A4 PDF
        const pdf = new jsPDF({ orientation: 'landscape', unit: 'mm', format: 'a4' })
        const pageWidth = pdf.internal.pageSize.getWidth()
        const pageHeight = pdf.internal.pageSize.getHeight()
        const margin = 10
        const availWidth = pageWidth - margin * 2
        const availHeight = pageHeight - margin * 2

        // 解析图片尺寸，等比缩放适配页面
        const imgProps = pdf.getImageProperties(imgData)
        const imgRatio = imgProps.width / imgProps.height
        const pageRatio = availWidth / availHeight
        let drawWidth, drawHeight
        if (imgRatio > pageRatio) {
          drawWidth = availWidth
          drawHeight = availWidth / imgRatio
        } else {
          drawHeight = availHeight
          drawWidth = availHeight * imgRatio
        }
        const x = (pageWidth - drawWidth) / 2
        const y = (pageHeight - drawHeight) / 2

        pdf.addImage(imgData, 'PNG', x, y, drawWidth, drawHeight)

        // 生成文件名：网络拓扑图-YYYY-MM-DD.pdf
        const today = new Date()
        const dateStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`
        pdf.save(`网络拓扑图-${dateStr}.pdf`)
      } catch (e) {
        console.error('exportPDF error:', e)
        this.$message.error('导出失败，请重试')
      } finally {
        this.exporting = false
      }
    },
    renderChart() {
      const el = this.$refs.chartRef
      if (!el) return
      if (!this.chart) {
        this.chart = echarts.init(el)
      }
      this.treeData = this.decorateNode(this.buildTreeData())
      this.chart.setOption({
        tooltip: {
          trigger: 'item',
          formatter: params => params.data && params.data.tooltip ? params.data.tooltip : params.name
        },
        series: [{
          type: 'tree',
          data: [this.treeData],
          orient: 'LR',
          top: '5%',
          bottom: '5%',
          left: '8%',
          right: '15%',
          roam: true,
          expandAndCollapse: true,
          // 默认展开到区域层，点击区域节点再展开其下资产，避免大量资产标签拥挤重叠
          initialTreeDepth: 2,
          symbol: 'circle',
          lineStyle: { color: '#cbd5e1', width: 1.5, curveness: 0.4 },
          emphasis: { focus: 'descendant' },
          animationDuration: 300,
          animationDurationUpdate: 300
        }]
      }, true)
      // 放行 roam 指针检查：使图表容器外部的代理滚轮/拖拽也能触发平移缩放
      this.allowExternalRoam()
      // 监听 finished 事件：每次渲染/动画完成后重新放行（节点展开/收起会重建视图重置 pointerChecker）
      this.chart.off('finished', this._onFinished)
      this._onFinished = () => this.allowExternalRoam()
      this.chart.on('finished', this._onFinished)
    },
    // 放行 RoamController 的指针命中检查（内部 API，失败不影响主功能）
    allowExternalRoam() {
      if (!this.chart) return
      try {
        const seriesModel = this.chart.getModel().getSeriesByIndex(0)
        const view = this.chart.getViewOfSeriesModel(seriesModel)
        if (view && view._controller) {
          view._controller.setPointerChecker(() => true)
        }
      } catch (e) {
        console.warn('allowExternalRoam failed', e)
      }
    },
    // 图表容器外（图例/头部/空白区域）的滚轮与拖拽代理：转投到图表，等效于在图内操作
    bindExternalRoam() {
      const root = this.$refs.pageRoot
      if (!root || this._roamBound) return
      this._roamBound = true
      this._onProxyWheel = e => this.proxyWheel(e)
      this._onProxyDown = e => this.proxyDown(e)
      this._onProxyMove = e => this.proxyMove(e)
      this._onProxyUp = e => this.proxyUp(e)
      root.addEventListener('wheel', this._onProxyWheel, { passive: false })
      root.addEventListener('mousedown', this._onProxyDown)
      document.addEventListener('mousemove', this._onProxyMove)
      document.addEventListener('mouseup', this._onProxyUp)
    },
    unbindExternalRoam() {
      const root = this.$refs.pageRoot
      if (root && this._onProxyWheel) {
        root.removeEventListener('wheel', this._onProxyWheel)
        root.removeEventListener('mousedown', this._onProxyDown)
        document.removeEventListener('mousemove', this._onProxyMove)
        document.removeEventListener('mouseup', this._onProxyUp)
      }
      this._roamBound = false
      this._proxyDragging = false
    },
    proxyWheel(e) {
      const chartEl = this.$refs.chartRef
      if (!this.chart || !chartEl || chartEl.contains(e.target) || this._dispatching) return
      e.preventDefault()
      this.zoomByWheel(e.deltaY < 0 ? 120 : -120)
    },
    proxyDown(e) {
      const chartEl = this.$refs.chartRef
      if (!this.chart || !chartEl || chartEl.contains(e.target) || e.button !== 0 || this._dispatching) return
      this._proxyDragging = true
      this.dispatchMouseToChart(chartEl, 'mousedown', e)
    },
    proxyMove(e) {
      if (!this._proxyDragging || !this.chart) return
      this.dispatchMouseToChart(this.$refs.chartRef, 'mousemove', e)
    },
    proxyUp(e) {
      if (!this._proxyDragging || !this.chart) return
      this._proxyDragging = false
      this.dispatchMouseToChart(this.$refs.chartRef, 'mouseup', e)
    },
    // 把原生鼠标事件按相同坐标转投到图表 DOM，走 zrender 完整事件管线
    dispatchMouseToChart(chartEl, type, e) {
      if (!chartEl || this._dispatching) return
      this._dispatching = true
      chartEl.dispatchEvent(new MouseEvent(type, {
        clientX: e.clientX,
        clientY: e.clientY,
        button: e.button,
        buttons: e.buttons,
        bubbles: true,
        cancelable: true
      }))
      this._dispatching = false
    },
    // 向画布中心派发合成 wheel 事件，走 ECharts 滚轮漫游的同一套处理逻辑，实现与滚轮完全一致的整体缩放
    zoomByWheel(deltaY) {
      if (!this.chart || this._dispatching) return
      this._dispatching = true
      const canvas = this.chart.getZr().painter.getViewportRoot()
      const rect = canvas.getBoundingClientRect()
      canvas.dispatchEvent(new WheelEvent('wheel', {
        deltaY,
        deltaMode: 0,
        clientX: rect.left + rect.width / 2,
        clientY: rect.top + rect.height / 2,
        bubbles: true,
        cancelable: true
      }))
      this._dispatching = false
    },
    // 放大（合成 wheel 事件经 Chrome 的 wheelDelta 兼容填充后符号与直觉相反，实测 deltaY>0 为放大）
    zoomIn() {
      this.zoomByWheel(120)
    },
    // 缩小
    zoomOut() {
      this.zoomByWheel(-120)
    },
    // 全部展开：所有节点 collapsed 置 false 后合并渲染（数据 diff 按稳定 id 精确匹配、瞬时更新，不会抖动）
    expandAll() {
      if (!this.chart || !this.treeData) return
      const expand = d => {
        d.collapsed = false
        ;(d.children || []).forEach(expand)
      }
      expand(this.treeData)
      this.chart.setOption({
        series: [{
          data: [this.treeData],
          // 瞬时更新，避免节点位置过渡动画引起抖动
          animationDurationUpdate: 0
        }]
      })
      // 恢复常规更新动画时长，保持后续节点折叠/展开动画
      this.chart.setOption({ series: [{ animationDurationUpdate: 300 }] })
    },
    // 重置视图：重新渲染以复位漫游/折叠状态
    resetView() {
      if (this.hasData) this.renderChart()
    },
    handleResize() {
      if (this.chart) this.chart.resize()
    }
  }
}
</script>

<style scoped>
.topology-graph {
  background: #fff;
  border-radius: 14px;
  border: 1px solid #e2e8f0;
  margin: 20px;
  padding: 24px;
  height: calc(100% - 85px);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.page-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
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

.legend-bar {
  display: flex;
  align-items: center;
  gap: 18px;
  font-size: 13px;
  color: #475569;
  margin-bottom: 10px;
}
.legend-item {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.legend-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
}
.legend-tip {
  margin-left: auto;
  color: #94a3b8;
  font-size: 12px;
}

.chart-wrap {
  position: relative;
  flex: 1;
  min-height: 400px;
}
.chart-container {
  position: absolute;
  inset: 0;
  border: 1px dashed #e2e8f0;
  border-radius: 10px;
  background: linear-gradient(180deg, #f8fafc 0%, #ffffff 100%);
}

/* 右侧缩放控制 */
.zoom-controls {
  position: absolute;
  right: 14px;
  top: 50%;
  transform: translateY(-50%);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  background: rgba(255, 255, 255, 0.92);
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  padding: 6px;
  box-shadow: 0 2px 8px rgba(15, 23, 42, 0.08);
  z-index: 10;
}
.zoom-btn {
  width: 30px;
  height: 30px;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  background: #ffffff;
  color: #334155;
  font-size: 18px;
  line-height: 1;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.zoom-btn:hover:not(:disabled) {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
}
.zoom-btn:disabled {
  color: #cbd5e1;
  cursor: not-allowed;
}
</style>
