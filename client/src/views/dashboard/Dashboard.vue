<template>
  <div class="dashboard">
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
      softwareChartInstance: null
    }
  },
  computed: {
    inUseCount() {
      const item = this.summary.status_stats.find(s => s.status === '在用')
      return item ? item.count : 0
    }
  },
  mounted() {
    this.fetchData()
    window.addEventListener('resize', this.handleResize)
  },
  beforeDestroy() {
    window.removeEventListener('resize', this.handleResize)
    if (this.regionChartInstance) this.regionChartInstance.dispose()
    if (this.osChartInstance) this.osChartInstance.dispose()
    if (this.levelChartInstance) this.levelChartInstance.dispose()
    if (this.trendChartInstance) this.trendChartInstance.dispose()
    if (this.softwareChartInstance) this.softwareChartInstance.dispose()
  },
  methods: {
    navigateTo(path, query) {
      this.$router.push({ path, query })
    },
    async fetchData() {
      try {
        const res = await getDashboardSummary()
        this.summary = res.data
        this.$nextTick(() => {
          this.renderRegionChart()
          this.renderOSChart()
          this.renderLevelChart()
          this.renderTrendChart()
          this.renderSoftwareChart()
        })
      } catch (e) {
        console.error(e)
      }
    },
    renderRegionChart() {
      if (!this.$refs.regionChart) return
      this.regionChartInstance = echarts.init(this.$refs.regionChart)
      const stats = this.summary.region_stats || []
      this.regionChartInstance.setOption({
        tooltip: { trigger: 'axis' },
        grid: { left: 50, right: 20, top: 30, bottom: 80 },
        xAxis: {
          type: 'category',
          data: stats.map(s => s.region_name || '未分配'),
          axisLabel: { interval: 0, rotate: 35, fontSize: 11 }
        },
        yAxis: { type: 'value', name: '资产数量' },
        series: [{
          type: 'bar',
          data: stats.map(s => s.count),
          barMaxWidth: 40,
          itemStyle: { color: '#409EFF' }
        }]
      })
    },
    renderOSChart() {
      if (!this.$refs.osChart) return
      this.osChartInstance = echarts.init(this.$refs.osChart)
      const stats = this.summary.os_stats || []
      this.osChartInstance.setOption({
        tooltip: { trigger: 'item' },
        legend: { bottom: 10, type: 'scroll', itemWidth: 12, itemHeight: 12, textStyle: { fontSize: 12 } },
        series: [{
          type: 'pie',
          radius: ['35%', '60%'],
          center: ['50%', '45%'],
          data: stats.map(s => ({ name: s.os_type, value: s.count })),
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 11 },
          emphasis: {
            itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
          }
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
        tooltip: { trigger: 'item' },
        legend: { bottom: 10, textStyle: { fontSize: 12 } },
        series: [{
          type: 'pie',
          radius: ['35%', '60%'],
          center: ['50%', '45%'],
          data: stats.map(s => ({
            name: s.level,
            value: s.count,
            itemStyle: { color: colorMap[s.level] || '#409EFF' }
          })),
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 11 },
          emphasis: {
            itemStyle: { shadowBlur: 10, shadowOffsetX: 0, shadowColor: 'rgba(0, 0, 0, 0.5)' }
          }
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
        tooltip: { trigger: 'axis' },
        grid: { left: 50, right: 20, top: 30, bottom: 40 },
        xAxis: {
          type: 'category',
          data: dates,
          axisLabel: { interval: 4, fontSize: 11 }
        },
        yAxis: { type: 'value', name: '操作次数', minInterval: 1 },
        series: [{
          type: 'line',
          data: counts,
          smooth: true,
          areaStyle: { color: 'rgba(64, 158, 255, 0.15)' },
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
      ]
      this.softwareChartInstance.setOption({
        tooltip: { trigger: 'item' },
        legend: { bottom: 10, textStyle: { fontSize: 12 } },
        series: [{
          type: 'pie',
          radius: ['40%', '65%'],
          center: ['50%', '45%'],
          data: data,
          label: { formatter: '{b}: {c} ({d}%)', fontSize: 12 },
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
.stat-card-clickable {
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
}
.stat-card-clickable:hover {
  transform: translateY(-4px);
}
.stat-card {
  display: flex;
  align-items: center;
}
.stat-icon {
  width: 60px;
  height: 60px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-right: 15px;
}
.stat-icon i {
  font-size: 28px;
  color: #fff;
}
.stat-value {
  font-size: 28px;
  font-weight: bold;
  color: #333;
}
.stat-label {
  font-size: 14px;
  color: #999;
  margin-top: 5px;
}
</style>
