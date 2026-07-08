<template>
  <div class="dashboard">
    <canvas ref="bgCanvas" class="tech-bg-canvas"></canvas>

    <!-- 核心指标 -->
    <div class="section-header">
      <span class="section-title">CORE METRICS</span>
      <span class="section-subtitle">核心指标</span>
      <div class="section-line"></div>
    </div>
    <el-row :gutter="24" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/assets')">
          <div class="stat-card">
            <div class="stat-indicator" style="background: #409EFF"></div>
            <div class="stat-icon-wrap"><i class="el-icon-monitor" style="color: #409EFF"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.total_assets }}</div>
              <div class="stat-label">总资产数</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/regions')">
          <div class="stat-card">
            <div class="stat-indicator" style="background: #67C23A"></div>
            <div class="stat-icon-wrap"><i class="el-icon-place" style="color: #67C23A"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.total_regions }}</div>
              <div class="stat-label">区域数量</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/backup-management')">
          <div class="stat-card">
            <div class="stat-indicator" style="background: #E6A23C"></div>
            <div class="stat-icon-wrap"><i class="el-icon-folder-opened" style="color: #E6A23C"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.backup_assets }}</div>
              <div class="stat-label">备份资产</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" :class="{ 'stat-alert': summary.total_unfixed_vulns > 0 }" @click.native="navigateTo('/vulnerability-scan')">
          <div class="stat-card">
            <div class="stat-indicator" :style="{ background: summary.total_unfixed_vulns > 0 ? '#EF4444' : '#22C55E' }"></div>
            <div class="stat-icon-wrap"><i class="el-icon-warning-outline" :style="{ color: summary.total_unfixed_vulns > 0 ? '#EF4444' : '#22C55E' }"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.total_unfixed_vulns }}</div>
              <div class="stat-label">漏洞总数</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 运维指标 -->
    <div class="section-header">
      <span class="section-title">OPERATIONS</span>
      <span class="section-subtitle">运维指标</span>
      <div class="section-line"></div>
    </div>
    <el-row :gutter="24" class="stat-row">
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/sftp-accounts')">
          <div class="stat-card">
            <div class="stat-indicator" style="background: #00BCD4"></div>
            <div class="stat-icon-wrap"><i class="el-icon-connection" style="color: #00BCD4"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.total_sftp_accounts }}</div>
              <div class="stat-label">SFTP账号</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/user-permissions')">
          <div class="stat-card">
            <div class="stat-indicator" style="background: #a855f7"></div>
            <div class="stat-icon-wrap"><i class="el-icon-user" style="color: #a855f7"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.total_user_permissions }}</div>
              <div class="stat-label">用户岗位</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" :class="{ 'stat-alert': summary.need_update_software > 0 }" @click.native="navigateTo('/approved-software')">
          <div class="stat-card">
            <div class="stat-indicator" :style="{ background: summary.need_update_software > 0 ? '#EF4444' : '#22C55E' }"></div>
            <div class="stat-icon-wrap"><i class="el-icon-refresh" :style="{ color: summary.need_update_software > 0 ? '#EF4444' : '#22C55E' }"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.need_update_software }}</div>
              <div class="stat-label">需更新软件</div>
            </div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card shadow="hover" class="stat-card-clickable" @click.native="navigateTo('/operation-logs')">
          <div class="stat-card">
            <div class="stat-indicator" style="background: #FF9800"></div>
            <div class="stat-icon-wrap"><i class="el-icon-document" style="color: #FF9800"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ animatedValues.monthly_op_count }}</div>
              <div class="stat-label">本月操作</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <!-- 图表区域 -->
    <el-row :gutter="24" class="chart-row">
      <el-col :span="12">
        <el-card shadow="hover" class="chart-card chart-stagger">
          <div slot="header" class="chart-header">
            <div class="chart-header-bar"></div>
            <span class="chart-header-title">区域资产分布</span>
          </div>
          <div ref="regionChart" style="height: 380px"></div>
        </el-card>
      </el-col>
      <el-col :span="12">
        <el-card shadow="hover" class="chart-card chart-stagger">
          <div slot="header" class="chart-header">
            <div class="chart-header-bar"></div>
            <span class="chart-header-title">操作系统分布</span>
          </div>
          <div ref="osChart" style="height: 380px"></div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="24" class="chart-row">
      <el-col :span="14">
        <el-card shadow="hover" class="chart-card chart-stagger">
          <div slot="header" class="chart-header">
            <div class="chart-header-bar"></div>
            <span class="chart-header-title">漏洞趋势</span>
          </div>
          <div ref="vulnTrendChart" style="height: 380px"></div>
        </el-card>
      </el-col>
      <el-col :span="10">
        <el-card shadow="hover" class="chart-card chart-stagger">
          <div slot="header" class="chart-header">
            <div class="chart-header-bar"></div>
            <span class="chart-header-title">近30天操作日志</span>
          </div>
          <div ref="trendChart" style="height: 380px"></div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="24" class="chart-row">
      <el-col :span="12">
        <el-card shadow="hover" class="chart-card chart-stagger">
          <div slot="header" class="chart-header">
            <div class="chart-header-bar"></div>
            <span class="chart-header-title">软件更新状态</span>
          </div>
          <div ref="softwareChart" style="height: 380px"></div>
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
        vuln_trend: [],
        trend_stats: [],
        software_update_stats: []
      },
      animatedValues: {
        total_assets: 0,
        total_regions: 0,
        total_sftp_accounts: 0,
        total_user_permissions: 0,
        need_update_software: 0,
        monthly_op_count: 0,
        total_unfixed_vulns: 0,
        backup_assets: 0
      },
      regionChartInstance: null,
      osChartInstance: null,
      vulnTrendChartInstance: null,
      trendChartInstance: null,
      softwareChartInstance: null,
      bgAnimationId: null
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
    if (this.vulnTrendChartInstance) this.vulnTrendChartInstance.dispose()
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
        // 连线（距离小于180px）
        for (let i = 0; i < particles.length; i++) {
          for (let j = i + 1; j < particles.length; j++) {
            const dx = particles[i].x - particles[j].x
            const dy = particles[i].y - particles[j].y
            const dist = Math.sqrt(dx * dx + dy * dy)
            if (dist < 180) {
              ctx.beginPath()
              ctx.moveTo(particles[i].x, particles[i].y)
              ctx.lineTo(particles[j].x, particles[j].y)
              ctx.strokeStyle = `rgba(64, 158, 255, ${0.25 * (1 - dist / 180)})`
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
          this.animateCountUp()
        }
        this.$nextTick(() => {
          this.renderRegionChart()
          this.renderOSChart()
          this.renderVulnTrendChart()
          this.renderTrendChart()
          this.renderSoftwareChart()
          this.initStaggerAnimation()
        })
      } catch (e) {
        console.error('获取看板数据失败:', e)
        this.$message.error('获取看板数据失败')
      }
    },
    animateCountUp() {
      const duration = 800
      const keys = ['total_assets', 'total_regions', 'total_sftp_accounts', 'total_user_permissions', 'need_update_software', 'monthly_op_count', 'total_unfixed_vulns', 'backup_assets']
      keys.forEach(key => {
        const target = this.summary[key] || 0
        const start = this.animatedValues[key] || 0
        const startTime = performance.now()
        const step = (now) => {
          const elapsed = now - startTime
          const progress = Math.min(elapsed / duration, 1)
          const eased = 1 - Math.pow(1 - progress, 3)
          this.animatedValues[key] = Math.round(start + (target - start) * eased)
          if (progress < 1) requestAnimationFrame(step)
        }
        requestAnimationFrame(step)
      })
    },
    initStaggerAnimation() {
      const cards = this.$el.querySelectorAll('.chart-stagger')
      cards.forEach((card, i) => {
        card.style.opacity = '0'
        card.style.transform = 'translateY(30px)'
        setTimeout(() => {
          card.style.transition = 'opacity 0.5s cubic-bezier(0.4,0,0.2,1), transform 0.5s cubic-bezier(0.4,0,0.2,1)'
          card.style.opacity = '1'
          card.style.transform = 'translateY(0)'
        }, i * 80)
      })
    },
    renderRegionChart() {
      if (!this.$refs.regionChart) return
      this.regionChartInstance = echarts.init(this.$refs.regionChart)
      const stats = this.summary.region_stats || []
      this.regionChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'axis',
          backgroundColor: 'rgba(10, 22, 40, 0.95)',
          borderColor: 'rgba(64, 158, 255, 0.4)',
          borderWidth: 1,
          textStyle: { color: '#F8FAFC', fontSize: 12 },
          padding: [10, 14],
          extraCssText: 'border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.4);'
        },
        grid: { left: 50, right: 20, top: 30, bottom: 80 },
        xAxis: {
          type: 'category',
          data: stats.map(s => s.region_name || '未分配'),
          axisLabel: { interval: 0, rotate: 35, fontSize: 11, color: 'rgba(255,255,255,0.7)' },
          axisLine: { lineStyle: { color: 'rgba(255,255,255,0.15)' } }
        },
        yAxis: { type: 'value', name: '资产数量', nameTextStyle: { color: 'rgba(255,255,255,0.6)' }, axisLabel: { color: 'rgba(255,255,255,0.6)' }, splitLine: { lineStyle: { color: 'rgba(255,255,255,0.06)' } } },
        series: [{
          type: 'bar',
          data: stats.map(s => s.count),
          barMaxWidth: 40,
          itemStyle: {
            color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: '#409EFF' }, { offset: 1, color: 'rgba(64,158,255,0.15)' }] },
            borderRadius: [4, 4, 0, 0]
          }
        }]
      })
    },
    renderOSChart() {
      if (!this.$refs.osChart) return
      this.osChartInstance = echarts.init(this.$refs.osChart)
      const stats = this.summary.os_stats || []
      const techColors = ['#409EFF', '#22C55E', '#E6A23C', '#a855f7', '#00BCD4', '#FF9800', '#EF4444', '#9C27B0']
      this.osChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(10, 22, 40, 0.95)',
          borderColor: 'rgba(64, 158, 255, 0.4)',
          borderWidth: 1,
          textStyle: { color: '#F8FAFC', fontSize: 12 },
          padding: [10, 14],
          extraCssText: 'border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.4);'
        },
        legend: { bottom: 10, type: 'scroll', itemWidth: 12, itemHeight: 12, textStyle: { fontSize: 12, color: 'rgba(255,255,255,0.7)' } },
        color: techColors,
        series: [{
          type: 'pie',
          radius: ['35%', '60%'],
          center: ['50%', '45%'],
          data: stats.map(s => ({ name: s.os_type, value: s.count })).sort((a, b) => a.value - b.value),
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 11, color: 'rgba(255,255,255,0.75)' },
          emphasis: {
            itemStyle: { shadowBlur: 20, shadowColor: 'rgba(64, 158, 255, 0.3)' }
          }
        }]
      })
    },
    handleResize() {
      if (this.regionChartInstance) this.regionChartInstance.resize()
      if (this.osChartInstance) this.osChartInstance.resize()
      if (this.vulnTrendChartInstance) this.vulnTrendChartInstance.resize()
      if (this.trendChartInstance) this.trendChartInstance.resize()
      if (this.softwareChartInstance) this.softwareChartInstance.resize()
    },
    renderVulnTrendChart() {
      if (!this.$refs.vulnTrendChart) return
      this.vulnTrendChartInstance = echarts.init(this.$refs.vulnTrendChart)
      const data = this.summary.vuln_trend || []
      const quarters = data.map(d => d.year + ' Q' + d.quarter)
      const critical = data.map(d => d.critical_count)
      const high = data.map(d => d.high_count)
      const medium = data.map(d => d.medium_count)

      this.vulnTrendChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'axis',
          backgroundColor: 'rgba(10, 22, 40, 0.95)',
          borderColor: 'rgba(64, 158, 255, 0.4)',
          borderWidth: 1,
          textStyle: { color: '#F8FAFC', fontSize: 12 },
          padding: [10, 14],
          extraCssText: 'border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.4);'
        },
        legend: {
          data: ['严重', '高危', '中危'],
          bottom: 10,
          textStyle: { fontSize: 12, color: 'rgba(255,255,255,0.7)' }
        },
        grid: { left: 50, right: 20, top: 30, bottom: 50 },
        xAxis: {
          type: 'category',
          data: quarters,
          boundaryGap: false,
          axisLabel: { fontSize: 11, color: 'rgba(255,255,255,0.65)' },
          axisLine: { lineStyle: { color: 'rgba(255,255,255,0.15)' } }
        },
        yAxis: {
          type: 'value',
          name: '漏洞数量',
          minInterval: 1,
          nameTextStyle: { color: 'rgba(255,255,255,0.6)' },
          axisLabel: { color: 'rgba(255,255,255,0.6)' },
          splitLine: { lineStyle: { color: 'rgba(255,255,255,0.06)' } }
        },
        series: [
          {
            name: '严重',
            type: 'line',
            stack: 'vuln',
            smooth: true,
            symbol: 'circle',
            symbolSize: 5,
            data: critical,
            lineStyle: { color: '#EF4444', width: 2, shadowColor: 'rgba(239,68,68,0.4)', shadowBlur: 8 },
            itemStyle: { color: '#EF4444', borderWidth: 2, borderColor: 'rgba(239,68,68,0.3)' },
            areaStyle: {
              color: {
                type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                colorStops: [
                  { offset: 0, color: 'rgba(239, 68, 68, 0.4)' },
                  { offset: 1, color: 'rgba(239, 68, 68, 0.02)' }
                ]
              }
            }
          },
          {
            name: '高危',
            type: 'line',
            stack: 'vuln',
            smooth: true,
            symbol: 'circle',
            symbolSize: 5,
            data: high,
            lineStyle: { color: '#E6A23C', width: 2, shadowColor: 'rgba(230,162,60,0.4)', shadowBlur: 8 },
            itemStyle: { color: '#E6A23C', borderWidth: 2, borderColor: 'rgba(230,162,60,0.3)' },
            areaStyle: {
              color: {
                type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                colorStops: [
                  { offset: 0, color: 'rgba(230, 162, 60, 0.35)' },
                  { offset: 1, color: 'rgba(230, 162, 60, 0.02)' }
                ]
              }
            }
          },
          {
            name: '中危',
            type: 'line',
            stack: 'vuln',
            smooth: true,
            symbol: 'circle',
            symbolSize: 5,
            data: medium,
            lineStyle: { color: '#C6B75E', width: 2, shadowColor: 'rgba(198,183,94,0.3)', shadowBlur: 6 },
            itemStyle: { color: '#C6B75E', borderWidth: 2, borderColor: 'rgba(198,183,94,0.3)' },
            areaStyle: {
              color: {
                type: 'linear', x: 0, y: 0, x2: 0, y2: 1,
                colorStops: [
                  { offset: 0, color: 'rgba(198, 183, 94, 0.3)' },
                  { offset: 1, color: 'rgba(198, 183, 94, 0.02)' }
                ]
              }
            }
          }
        ]
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
          backgroundColor: 'rgba(10, 22, 40, 0.95)',
          borderColor: 'rgba(64, 158, 255, 0.4)',
          borderWidth: 1,
          textStyle: { color: '#F8FAFC', fontSize: 12 },
          padding: [10, 14],
          extraCssText: 'border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.4);'
        },
        grid: { left: 50, right: 20, top: 30, bottom: 40 },
        xAxis: {
          type: 'category',
          data: dates,
          axisLabel: { interval: 4, fontSize: 11, color: 'rgba(255,255,255,0.65)' },
          axisLine: { lineStyle: { color: 'rgba(255,255,255,0.15)' } }
        },
        yAxis: { type: 'value', name: '操作次数', minInterval: 1, nameTextStyle: { color: 'rgba(255,255,255,0.6)' }, axisLabel: { color: 'rgba(255,255,255,0.6)' }, splitLine: { lineStyle: { color: 'rgba(255,255,255,0.06)' } } },
        series: [{
          type: 'line',
          data: counts,
          smooth: true,
          areaStyle: { color: { type: 'linear', x: 0, y: 0, x2: 0, y2: 1, colorStops: [{ offset: 0, color: 'rgba(64, 158, 255, 0.25)' }, { offset: 1, color: 'rgba(64, 158, 255, 0.02)' }] } },
          lineStyle: { color: '#409EFF', width: 2, shadowColor: 'rgba(64,158,255,0.4)', shadowBlur: 8 },
          itemStyle: { color: '#409EFF', borderWidth: 2, borderColor: 'rgba(64,158,255,0.3)' },
          symbol: 'circle',
          symbolSize: 5
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
        { name: '已是最新', value: uptodateCount ? uptodateCount.count : 0, itemStyle: { color: '#22C55E' } },
        { name: '需要更新', value: updateCount ? updateCount.count : 0, itemStyle: { color: '#EF4444' } }
      ].sort((a, b) => a.value - b.value)
      this.softwareChartInstance.setOption({
        backgroundColor: 'transparent',
        tooltip: {
          trigger: 'item',
          backgroundColor: 'rgba(10, 22, 40, 0.95)',
          borderColor: 'rgba(64, 158, 255, 0.4)',
          borderWidth: 1,
          textStyle: { color: '#F8FAFC', fontSize: 12 },
          padding: [10, 14],
          extraCssText: 'border-radius: 8px; box-shadow: 0 4px 20px rgba(0,0,0,0.4);'
        },
        legend: { bottom: 10, textStyle: { fontSize: 12, color: 'rgba(255,255,255,0.7)' } },
        series: [{
          type: 'pie',
          radius: ['40%', '65%'],
          center: ['50%', '45%'],
          data: data,
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 12, color: 'rgba(255,255,255,0.75)' },
          emphasis: {
            itemStyle: { shadowBlur: 20, shadowColor: 'rgba(0, 0, 0, 0.5)' }
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
  padding: 24px;
  overflow: hidden;
}
.tech-bg-canvas {
  position: absolute;
  top: 0; left: 0;
  width: 100%; height: 100%;
  pointer-events: none;
  z-index: 0;
}
.dashboard > .el-row,
.dashboard > .section-header {
  position: relative;
  z-index: 1;
}

/* 分组标题 */
.section-header {
  margin-bottom: 16px;
  position: relative;
  display: flex;
  align-items: baseline;
  gap: 8px;
}
.section-title {
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 2px;
  color: rgba(64, 158, 255, 0.7);
  text-transform: uppercase;
  font-family: 'Maple Mono NF', 'Fira Code', monospace;
}
.section-subtitle {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.35);
}
.section-line {
  flex: 1;
  height: 1px;
  background: linear-gradient(90deg, rgba(64, 158, 255, 0.3) 0%, rgba(168, 85, 247, 0.15) 50%, transparent 100%);
  margin-left: 8px;
}

.stat-row {
  margin-bottom: 24px;
}
.chart-row {
  margin-bottom: 24px;
}

/* 玻璃卡片效果 */
.dashboard ::v-deep .el-card {
  background: rgba(13, 33, 55, 0.2) !important;
  border-width: 1px !important;
  border-style: solid !important;
  border-color: rgba(64, 158, 255, 0.15);
  border-radius: 12px !important;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.06) !important;
  position: relative;
  overflow: hidden;
}
.dashboard ::v-deep .el-card .el-card__header {
  color: rgba(255, 255, 255, 0.95);
  border-bottom: none !important;
  font-weight: 500;
  background: transparent !important;
  padding-bottom: 12px;
}
.dashboard ::v-deep .el-card .el-card__body {
  background: transparent !important;
}

/* 可点击统计卡片 */
.stat-card-clickable {
  cursor: pointer;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1),
              box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
.stat-card-clickable:hover {
  transform: translateY(-8px);
  box-shadow: 0 16px 48px rgba(64, 158, 255, 0.25), inset 0 1px 0 rgba(255, 255, 255, 0.1) !important;
  border-color: rgba(64, 158, 255, 0.4) !important;
  background-color: rgba(13, 33, 55, 0.35) !important;
}
/* 异常卡片悬停时保持红色光效，不与蓝色混合 */
.stat-card-clickable.stat-alert:hover {
  transform: translateY(-8px);
  box-shadow: 0 16px 48px rgba(239, 68, 68, 0.3), inset 0 1px 0 rgba(255, 255, 255, 0.1) !important;
  border-color: rgba(239, 68, 68, 0.6) !important;
  background-color: rgba(60, 20, 20, 0.35) !important;
  animation-play-state: paused !important;
}
/* hover 顶部光线 - 正常卡片为蓝色 */
.stat-card-clickable::after {
  content: '';
  position: absolute;
  top: 0; left: 10%; right: 10%;
  height: 2px;
  background: linear-gradient(90deg, transparent, #409EFF, transparent);
  opacity: 0;
  transition: opacity 0.3s cubic-bezier(0.4, 0, 0.2, 1);
  z-index: 10;
}
.stat-card-clickable:hover::after {
  opacity: 1;
}
/* 异常卡片悬停时顶部改为红色光线 */
.stat-card-clickable.stat-alert::after {
  background: linear-gradient(90deg, transparent, #EF4444, transparent);
}

/* 异常状态脉冲动画 - 移到全局样式块 */

/* 统计卡片内容 */
.stat-card {
  display: flex;
  align-items: center;
}
.stat-indicator {
  width: 4px;
  height: 48px;
  border-radius: 2px;
  margin-right: 16px;
  flex-shrink: 0;
  opacity: 0.8;
}
.stat-icon-wrap {
  width: 44px;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 14px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.04);
  flex-shrink: 0;
}
.stat-icon-wrap i {
  font-size: 22px;
}
.stat-value {
  font-size: 30px;
  font-weight: 700;
  color: #F8FAFC;
  text-shadow: 0 0 10px rgba(64, 158, 255, 0.3);
  font-variant-numeric: tabular-nums;
  line-height: 1.2;
}
.stat-label {
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 4px;
  letter-spacing: 0.5px;
}

/* 图表卡片 header */
.chart-header {
  display: flex;
  align-items: center;
  gap: 10px;
}
.chart-header-bar {
  width: 3px;
  height: 16px;
  border-radius: 2px;
  background: linear-gradient(180deg, #409EFF, #a855f7);
  flex-shrink: 0;
}
.chart-header-title {
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
}
/* header 底部光线分割线 */
.dashboard ::v-deep .chart-card .el-card__header {
  position: relative;
}
.dashboard ::v-deep .chart-card .el-card__header::after {
  content: '';
  position: absolute;
  bottom: 0; left: 0; right: 0;
  height: 1px;
  background: linear-gradient(90deg, rgba(64, 158, 255, 0.3), rgba(168, 85, 247, 0.15), transparent);
}
</style>

<style>
/* 全局样式 - 确保玻璃效果 */
.dashboard .el-card.el-card {
  background-color: rgba(13, 33, 55, 0.2) !important;
  border-color: rgba(64, 158, 255, 0.15);
}
.dashboard .el-card::before {
  content: '';
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  border-radius: inherit;
  background: rgba(13, 33, 55, 0.15);
  backdrop-filter: blur(5px) saturate(1.4);
  -webkit-backdrop-filter: blur(5px) saturate(1.4);
  z-index: -1;
}
/* 统计卡片不需要 backdrop-filter 玻璃层，避免遮盖边框动画 */
.dashboard .stat-card-clickable::before {
  display: none !important;
}
.dashboard .stat-card-clickable:hover::before {
  display: none !important;
}
.dashboard .el-card .el-card__header,
.dashboard .el-card .el-card__body {
  background: transparent !important;
}

/* 异常状态脉冲动画（全局，覆盖 Element UI 优先级） */
/* 注意：不可对 border-color 使用 !important，否则 WebKit/Blink 会阻止 keyframes 动画 */
.dashboard .el-card.stat-alert {
  animation: pulse-border 2.5s ease-in-out infinite !important;
}
.dashboard .el-card.stat-alert:hover {
  animation: none !important;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1), box-shadow 0.3s cubic-bezier(0.4, 0, 0.2, 1), border-color 0.2s ease !important;
}
@keyframes pulse-border {
  0%, 100% {
    border-color: rgba(239, 68, 68, 0.2);
    box-shadow: 0 8px 32px rgba(0,0,0,0.3), 0 0 0 0 rgba(239,68,68,0);
  }
  50% {
    border-color: rgba(239, 68, 68, 0.85);
    box-shadow: 0 8px 32px rgba(0,0,0,0.3), 0 0 24px 3px rgba(239,68,68,0.3);
  }
}
</style>
