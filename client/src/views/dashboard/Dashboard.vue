<template>
  <div class="dashboard">
    <el-row :gutter="20" style="margin-bottom: 20px">
      <el-col :span="6">
        <el-card shadow="hover">
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
        <el-card shadow="hover">
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
        <el-card shadow="hover">
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
        <el-card shadow="hover">
          <div class="stat-card">
            <div class="stat-icon" style="background: #909399"><i class="el-icon-warning-outline"></i></div>
            <div class="stat-info">
              <div class="stat-value">{{ idleCount }}</div>
              <div class="stat-label">闲置资产</div>
            </div>
          </div>
        </el-card>
      </el-col>
    </el-row>
    <el-row :gutter="20">
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
        region_stats: [],
        os_stats: [],
        status_stats: []
      },
      regionChartInstance: null,
      osChartInstance: null
    }
  },
  computed: {
    inUseCount() {
      const item = this.summary.status_stats.find(s => s.status === '在用')
      return item ? item.count : 0
    },
    idleCount() {
      const item = this.summary.status_stats.find(s => s.status === '闲置')
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
  },
  methods: {
    async fetchData() {
      try {
        const res = await getDashboardSummary()
        this.summary = res.data
        this.$nextTick(() => {
          this.renderRegionChart()
          this.renderOSChart()
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
    }
  }
}
</script>

<style scoped>
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
