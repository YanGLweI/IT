<template>
  <div class="dashboard">
    <canvas ref="bgCanvas" class="tech-bg-canvas"></canvas>
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/assets')">
          <div class="stat-card">
            <div class="stat-icon" style="background: #409EFF"><i class="el-icon-monitor"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ summary.total_assets || 0 }}</div>
              <div class="stat-label">总资产数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/regions')">
          <div class="stat-card">
            <div class="stat-icon" style="background: #67C23A"><i class="el-icon-place"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ summary.total_regions || 0 }}</div>
              <div class="stat-label">区域数量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/assets', { status: '在用' })">
          <div class="stat-card">
            <div class="stat-icon" style="background: #E6A23C"><i class="el-icon-check"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ inUseCount }}</div>
              <div class="stat-label">在用资产</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/vulnerability-scan')">
          <div class="stat-card">
            <div class="stat-icon" :style="{ background: summary.total_unfixed_vulns > 0 ? '#F56C6C' : '#67C23A' }"><i class="el-icon-warning-outline"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ summary.total_unfixed_vulns || 0 }}</div>
              <div class="stat-label">漏洞总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/sftp-accounts')">
          <div class="stat-card">
            <div class="stat-icon" style="background: #00BCD4"><i class="el-icon-connection"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ summary.total_sftp_accounts || 0 }}</div>
              <div class="stat-label">SFTP账号</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/user-permissions')">
          <div class="stat-card">
            <div class="stat-icon" style="background: #9C27B0"><i class="el-icon-user"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ summary.total_user_permissions || 0 }}</div>
              <div class="stat-label">用户岗位</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/approved-software')">
          <div class="stat-card">
            <div class="stat-icon" :style="{ background: summary.need_update_software > 0 ? '#F56C6C' : '#67C23A' }"><i class="el-icon-refresh"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ summary.need_update_software || 0 }}</div>
              <div class="stat-label">需更新软件</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/operation-logs')">
          <div class="stat-card">
            <div class="stat-icon" style="background: #FF9800"><i class="el-icon-document"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ summary.monthly_op_count || 0 }}</div>
              <div class="stat-label">本月操作</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="12">
        <el-card shadow="hover">
          <div slot="header">区域资产分布</div>
          <div ref="regionChart" style="height: 420px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <div slot="header">操作系统分布</div>
          <div ref="osChart" style="height: 420px"></div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="12">
        <el-card shadow="hover">
          <div slot="header">资产等级分布</div>
          <div ref="levelChart" style="height: 420px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover">
          <div slot="header">近30天操作日志趋势</div>
          <div ref="trendChart" style="height: 420px"></div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20">
      <el-col :span="12">
        <el-card shadow="hover">
          <div slot="header">软件更新状态</div>
          <div ref="softwareChart" style="height: 420px"></div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import * as echarts from 'echarts'
import { getDashboardSummary } from '@/api/dashboard'

export default {
  name: 'Dashboard',
  data() {
    return {
      summary: {
        total_assets: 0,
        total_regions: 0,
        total_sftp_accounts: 0,
        total_user_permissions: 0,
        need_update_software: 0,
        monthly_op_count: 0,
        total_unfixed_vulns: 0,
        region_stats: [],
        os_stats: [],
        status_stats: [],
        level_stats: [],
        trend_stats: [],
        software_update_stats: []
      },
      regionChartInstance: null,
      osChartInstance: null,
      levelChartInstance: null,
      trendChartInstance: null,
      softwareChartInstance: null,
      bgAnimationId: null
    }
  },
  computed: {
    inUseCount() {
      const item = this.summary.status_stats.find(s => s.status === '在用')
      return item ? item.count : 0
    }
  },
  mounted() {
    this.initBgCanvas()
    this.fetchData()
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.handleResize)
    if (this.bgAnimationId) cancelAnimationFrame(this.bgAnimationId)
    if (this.regionChartInstance) this.regionChartInstance.dispose()
    if (this.osChartInstance) this.osChartInstance.dispose()
    if (this.levelChartInstance) this.levelChartInstance.dispose()
    if (this.trendChartInstance) this.trendChartInstance.dispose()
    if (this.softwareChartInstance) this.softwareChartInstance.dispose()
  },
  methods: {
    initBgCanvas() {
      const canvas = this.$refs.bgCanvas
      if (!canvas) return
      const ctx = canvas.getContext('2d')
      const resize = () => {
        canvas.width = canvas.parentElement.offsetWidth
        canvas.height = canvas.parentElement.offsetHeight
      }
      resize()
      window.addEventListener('resize', resize)

      // 生成粒子
      const particles = []
      for (let i = 0; i < 150; i++) {
        particles.push({
          x: Math.random() * canvas.width,
          y: Math.random() * canvas.height,
          r: Math.random() * 3 + 1.5,
          dx: (Math.random() - 0.5) * 0.5,
          dy: (Math.random() - 0.5) * 0.5,
          opacity: Math.random() * 0.3 + 0.6
        })
      }

      const animate = () => {
        ctx.clearRect(0, 0, canvas.width, canvas.height)
        // 绘制粒子
        particles.forEach(p => {
          ctx.beginPath()
          ctx.arc(p.x, p.y, p.r, 0, Math.PI * 2)
          ctx.fillStyle = `rgba(64, 158, 255, ${p.opacity})`
          ctx.fill()
          // 光晕
          ctx.beginPath()
          ctx.arc(p.x, p.y, p.r * 3, 0, Math.PI * 2)
          ctx.fillStyle = `rgba(64, 158, 255, ${p.opacity * 0.1})`
          ctx.fill()
          p.x += p.dx
          p.y += p.dy
          if (p.x < 0 || p.x > canvas.width) p.dx *= -1
          if (p.y < 0 || p.y > canvas.height) p.dy *= -1
        })
        // 连线（距离小于150px）
        for (let i = 0; i < particles.length; i++) {
          for (let j = i + 1; j < particles.length; j++) {
            const dx = particles[i].x - particles[j].x
            const dy = particles[i].y - particles[j].y
            const dist = Math.sqrt(dx * dx + dy * dy)
            if (dist < 150) {
              ctx.beginPath()
              ctx.moveTo(particles[i].x, particles[i].y)
              ctx.lineTo(particles[j].x, particles[j].y)
              ctx.strokeStyle = `rgba(64, 158, 255, ${0.25 * (1 - dist / 150)})`
              ctx.lineWidth = 0.6
              ctx.stroke()
            }
          }
        }
        this.bgAnimationId = requestAnimationFrame(animate)
      }
      animate()
    },
    navigateTo(path, query) {
      this.$router.push({ path, query })
    },
    async fetchData() {
      try {
        const res = await getDashboardSummary()
        if (res && res.data) {
          Object.assign(this.summary, res.data)
        }
        this.$nextTick(() => {
          this.renderRegionChart()
          this.renderOSChart()
          this.renderLevelChart()
          this.renderTrendChart()
          this.renderSoftwareChart()
        })
      } catch (e) {
        console.error('获取看板数据失败:', e)
        this.$message.error('获取看板数据失败')
      }
    },
    renderRegionChart() {
      if (!this.$refs.regionChart) return
      this.regionChartInstance = echarts.init(this.$refs.regionChart)
      const stats = this.summary.region_stats || []
      this.regionChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'axis',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          borderColor: 'rgba(64, 158, 255, 0.3)',
          textStyle: { color: '#fff', fontSize: 12 },
          padding: [8, 12]
        },
        grid: { left: 50, right: 20, top: 30, bottom: 80 },
        xAxis: {
          type: 'category',
          data: stats.map(s => s.region_name || '未分配'),
          axisLabel: { interval: 0, rotate: 35, fontSize: 11, color: 'rgba(255,255,255,0.7)' },
          axisLine: { lineStyle: { color: 'rgba(255,255,255,0.2)' } }
        },
        yAxis: { type: 'value', name: '资产数量', nameTextStyle: { color: 'rgba(255,255,255,0.7)' }, axisLabel: { color: 'rgba(255,255,255,0.7)' }, splitLine: { lineStyle: { color: 'rgba(255,255,255,0.1)' } } },
        series: [{
          type: 'bar',
          data: stats.map(s => s.count),
          barMaxWidth: 40,
          itemStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: '#409EFF' }, { offset: 1, color: 'rgba(64,158,255,0.3)' }] } }
        }]
      })
    },
    renderOSChart() {
      if (!this.$refs.osChart) return
      this.osChartInstance = echarts.init(this.$refs.osChart)
      const stats = this.summary.os_stats || []
      this.osChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          borderColor: 'rgba(64, 158, 255, 0.3)',
          textStyle: { color: '#fff', fontSize: 12 },
          padding: [8, 12]
        },
        legend: { bottom: 10, type: 'scroll', itemWidth: 12, itemHeight: 12, textStyle: { fontSize: 12, color: 'rgba(255,255,255,0.8)' } },
        series: [{
          type: 'pie',
          radius: ['35%', '60%'],
          center: ['50%', '45%'],
          data: stats.map(s => ({ name: s.os_type, value: s.count })).sort((a, b) => a.value - b.value),
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 11, color: 'rgba(255,255,255,0.8)' }
        }]
      })
    },
    handleResize() {
      if (this.regionChartInstance) this.regionChartInstance.resize()
      if (this.osChartInstance) this.osChartInstance.resize()
      if (this.levelChartInstance) this.levelChartInstance.resize()
      if (this.trendChartInstance) this.trendChartInstance.resize()
      if (this.softwareChartInstance) this.softwareChartInstance.resize()
    },
    renderLevelChart() {
      if (!this.$refs.levelChart) return
      this.levelChartInstance = echarts.init(this.$refs.levelChart)
      const stats = this.summary.level_stats || []
      const colorMap = { '高': '#F56C6C', '中': '#E6A23C', '低': '#67C23A', '未分级': '#909399' }
      this.levelChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          borderColor: 'rgba(64, 158, 255, 0.3)',
          textStyle: { color: '#fff', fontSize: 12 },
          padding: [8, 12]
        },
        legend: { bottom: 10, textStyle: { fontSize: 12, color: 'rgba(255,255,255,0.8)' } },
        series: [{
          type: 'pie',
          radius: ['35%', '60%'],
          center: ['50%', '45%'],
          data: stats.map(s => ({
            name: s.level,
            value: s.count,
            itemStyle: { color: colorMap[s.level] || '#409EFF' }
          })).sort((a, b) => a.value - b.value),
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 11, color: 'rgba(255,255,255,0.8)' }
        }]
      })
    },
    renderTrendChart() {
      if (!this.$refs.trendChart) return
      this.trendChartInstance = echarts.init(this.$refs.trendChart)
      const stats = this.summary.trend_stats || []
      // 补全30天日期，无数据的天填0
      const dateMap = {}
      stats.forEach(s => { dateMap[s.date] = s.count })
      const dates = []
      const counts = []
      const now = new Date()
      for (let i = 29; i >= 0; i--) {
        const d = new Date(now)
        d.setDate(d.getDate() - i)
        const dateStr = d.toISOString().slice(0, 10)
        dates.push(dateStr.slice(5)) // MM-DD
        counts.push(dateMap[dateStr] || 0)
      }
      this.trendChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'axis',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          borderColor: 'rgba(64, 158, 255, 0.3)',
          textStyle: { color: '#fff', fontSize: 12 },
          padding: [8, 12]
        },
        grid: { left: 50, right: 20, top: 30, bottom: 40 },
        xAxis: {
          type: 'category',
          data: dates,
          axisLabel: { interval: 4, fontSize: 11, color: 'rgba(255,255,255,0.7)' },
          axisLine: { lineStyle: { color: 'rgba(255,255,255,0.2)' } }
        },
        yAxis: { type: 'value', name: '操作次数', minInterval: 1, nameTextStyle: { color: 'rgba(255,255,255,0.7)' }, axisLabel: { color: 'rgba(255,255,255,0.7)' }, splitLine: { lineStyle: { color: 'rgba(255,255,255,0.1)' } } },
        series: [{
          type: 'line',
          data: counts,
          smooth: true,
          areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(64, 158, 255, 0.3)' }, { offset: 1, color: 'rgba(64, 158, 255, 0.02)' }] } },
          lineStyle: { color: '#409EFF', width: 2 },
          itemStyle: { color: '#409EFF' },
          symbol: 'circle',
          symbolSize: 4
        }]
      })
    },
    renderSoftwareChart() {
      if (!this.$refs.softwareChart) return
      this.softwareChartInstance = echarts.init(this.$refs.softwareChart)
      const stats = this.summary.software_update_stats || []
      const updateCount = stats.find(s => s.need_update === true || s.need_update === 1)
      const uptodateCount = stats.find(s => s.need_update === false || s.need_update === 0)
      const data = [
        { name: '已是最新', value: uptodateCount ? uptodateCount.count : 0, itemStyle: { color: '#67C23A' } },
        { name: '需要更新', value: updateCount ? updateCount.count : 0, itemStyle: { color: '#F56C6C' } }
      ].sort((a, b) => a.value - b.value)
      this.softwareChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(0, 0, 0, 0.8)',
          borderColor: 'rgba(64, 158, 255, 0.3)',
          textStyle: { color: '#fff', fontSize: 12 },
          padding: [8, 12]
        },
        legend: { bottom: 10, textStyle: { fontSize: 12, color: 'rgba(255,255,255,0.8)' } },
        series: [{
          type: 'pie',
          radius: ['40%', '65%'],
          center: ['50%', '45%'],
          data: data,
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 12, color: 'rgba(255,255,255,0.8)' },
          emphasis: {
            itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
          }
        }]
      })
    }
  }
}
</script>

<style scoped>
.dashboard {
  position: relative;
  min-height: 100vh;
  background: linear-gradient(135deg, #0a1628 0%, #0d2137 30%, #0f1b2d 60%, #131a2e 100%);
  background-attachment: fixed;
  padding: 20px;
  overflow: hidden;
}
.tech-bg-canvas {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  pointer-events: none;
  z-index: 0;
}
.dashboard > .el-row {
  position: relative;
  z-index: 1;
}

/* 玻璃卡片效果 */
.dashboard ::v-deep .el-card {
  background: rgba(13, 33, 55, 0.2) !important;
  border: 1px solid rgba(64, 158, 255, 0.2) !important;
  border-radius: 12px !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.08) !important;
  position: relative;
  overflow: hidden;
}
.dashboard ::v-deep .el-card .el-card__header {
  color: rgba(255, 255, 255, 0.95);
  border-bottom: 1px solid rgba(64, 158, 255, 0.15) !important;
  font-weight: 500;
  background: transparent !important;
}
.dashboard ::v-deep .el-card .el-card__body {
  background: transparent !important;
}

/* 可点击卡片 */
.stat-card-clickable {
  cursor: pointer;
  transition: transform 0.3s ease, box-shadow 0.3s ease, border-color 0.3s ease;
}
.stat-card-clickable:hover {
  transform: translateY(-6px);
  box-shadow: 0 12px 40px rgba(64, 158, 255, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.15) !important;
  border-color: rgba(64, 158, 255, 0.5) !important;
  background-color: rgba(13, 33, 55, 0.35) !important;
}

.stat-card {
  display: flex;
  align-items: center;
}
.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.2);
}
.stat-icon i {
  font-size: 28px;
  color: #fff;
}
.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #fff;
  text-shadow: 0 0 10px rgba(64, 158, 255, 0.3);
}
.stat-label {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 5px;
}
</style>

<style>
/* 全局样式 - 确保玻璃效果 */
.dashboard .el-card.el-card {
  background-color: rgba(13, 33, 55, 0.2) !important;
  border-color: rgba(64, 158, 255, 0.2) !important;
}
.dashboard .el-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  border-radius: inherit;
  background: rgba(13, 33, 55, 0.15);
  backdrop-filter: blur(5px) saturate(1.4);
  -webkit-backdrop-filter: blur(5px) saturate(1.4);
  z-index: -1;
}
.dashboard .stat-card-clickable:hover::before {
  background: rgba(13, 33, 55, 0.3);
}
.dashboard .el-card .el-card__header,
.dashboard .el-card .el-card__body {
  background: transparent !important;
}
</style>
