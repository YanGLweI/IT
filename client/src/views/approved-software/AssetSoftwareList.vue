<template>
  <div class="asset-software-list">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">资产对应表</h2>
        <p class="page-subtitle">第三方软件与资产关联管理</p>
      </div>
      <div class="header-actions">
        <el-button type="success" size="small" icon="el-icon-download" :loading="exporting" @click="handleExport">导出补丁更新记录表</el-button>
      </div>
    </div>

    <!-- 筛选栏 -->
    <div class="filter-bar">
      <el-input
        v-model="search"
        placeholder="计算机名或IP地址"
        size="small"
        clearable
        style="width: 240px"
        @clear="handleSearch"
        @keyup.enter.native="handleSearch"
      />
      <el-select
        v-model="selectedSoftwareFilter"
        multiple
        filterable
        collapse-tags
        size="small"
        placeholder="按软件筛选"
        style="flex: 1; min-width: 200px; max-width: 400px"
        @change="handleSearch"
      >
        <el-option
          v-for="sw in allSoftware"
          :key="sw.id"
          :label="sw.name + (sw.version ? ' (' + sw.version + ')' : '')"
          :value="sw.id"
        />
      </el-select>
    </div>

    <!-- 数据表格 -->
    <div class="table-card" ref="tableCard">
      <div class="table-wrapper">
        <el-table :data="list" stripe v-loading="loading" :max-height="tableMaxHeight">
          <el-table-column type="index" label="序号" width="75" align="center" :index="indexMethod" />
          <el-table-column prop="computer_name" label="计算机名" width="200" />
          <el-table-column prop="ip_address" label="IP地址" width="150" />
          <el-table-column label="第三方软件" min-width="250">
            <template slot-scope="scope">
              <template v-if="scope.row.software_list && scope.row.software_list.length > 0">
                <el-tag
                  v-for="sw in scope.row.software_list"
                  :key="sw.id"
                  size="small"
                  style="margin: 2px 4px 2px 0"
                >{{ sw.name }} {{ sw.version ? '(' + sw.version + ')' : '' }}</el-tag>
              </template>
              <span v-else style="color: #999">未关联</span>
            </template>
          </el-table-column>
          <el-table-column label="操作" width="150" fixed="right" align="center">
            <template slot-scope="scope">
              <el-button size="mini" @click="handleEdit(scope.row)">关联</el-button>
            </template>
          </el-table-column>
        </el-table>
      </div>
    </div>

    <!-- 分页 -->
    <div class="pagination-wrap">
      <el-pagination
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size.sync="pageSize"
        :current-page.sync="currentPage"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </div>

    <!-- 编辑关联软件弹窗 -->
    <el-dialog class="vault-dialog" title="关联核准软件" :visible.sync="editDialogVisible" width="560px" :close-on-click-modal="false">
      <div style="margin-bottom: 12px; color: #606266; font-size: 14px">
        资产：<strong>{{ editRow ? editRow.computer_name : '' }}</strong>（{{ editRow ? editRow.ip_address : '' }}）
      </div>
      <el-divider />
      <div style="margin-bottom: 10px; color: #909399; font-size: 13px">请勾选该资产上已安装的核准软件：</div>
      <div class="software-list-wrapper">
        <el-checkbox-group v-model="selectedSoftwareIds">
          <div class="software-list">
            <el-checkbox
              v-for="sw in allSoftware"
              :key="sw.id"
              :label="sw.id"
            >
              {{ sw.name }}
              <span v-if="sw.version" style="color: #999; font-size: 12px">({{ sw.version }})</span>
            </el-checkbox>
          </div>
        </el-checkbox-group>
        <div v-if="allSoftware.length === 0" class="empty-state">
          暂无核准软件，请先在"核准软件目录"中添加
        </div>
      </div>
      <span slot="footer">
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSaveLinks">保存</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />

    <!-- 关联第三方软件提示弹窗 -->
    <el-dialog
      class="link-prompt-dialog"
      title="关联第三方软件"
      :visible.sync="showLinkPrompt"
      width="480px"
      :close-on-click-modal="false"
    >
      <div class="prompt-content">
        <p class="prompt-text">是否立即为新资产关联第三方软件？</p>
        <el-tag v-if="newAssetComputerName" size="small" type="success">{{ newAssetComputerName }}</el-tag>
      </div>
      <span slot="footer">
        <el-button @click="handleSkipAsset">稍后处理</el-button>
        <el-button type="primary" @click="handleGoToAsset">立即前往</el-button>
      </span>
    </el-dialog>
  </div>
</template>

<script>
import {
  getAssetSoftwareList,
  getAssetSoftwareLinks,
  updateAssetSoftwareLinks,
  getApprovedSoftware,
  exportPatchUpdateRecord
} from '@/api/approved_software'
import DualControlDialog from '@/components/DualControlDialog.vue'
import tableHeightMixin from '@/mixins/table-height'

export default {
  name: 'AssetSoftwareList',
  components: { DualControlDialog },
  mixins: [tableHeightMixin],
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      currentPage: 1,
      pageSize: 10,
      search: '',
      selectedSoftwareFilter: [],
      allSoftware: [],
      editDialogVisible: false,
      editRow: null,
      selectedSoftwareIds: [],
      submitting: false,
      exporting: false,
      autoLocateHandled: false,  // 标记是否已处理自动定位
      newAssetId: null,        // 记录新创建的资产 ID
      newAssetComputerName: '', // 新创建资产的计算机名
      showLinkPrompt: false    // 显示关联提示弹窗
    }
  },
  async mounted() {
    this.fetchData()
    await this.fetchAllSoftware()
    
    // 检查是否有来自资产创建的跳转参数
    const assetId = this.$route.query.asset_id
    if (assetId && !this.autoLocateHandled) {
      // 立即清理查询参数，避免后续重复触发
      this.autoLocateHandled = true
      this.$router.replace({ query: {} })
      
      // 数据应该已经加载完成，直接定位
      console.log('[AutoLocate] 开始定位资产 ID:', assetId)
      this.locateAndOpen(assetId)
    }
    
    // 监听子组件的 success 事件
    this.$on('child-success', this.handleChildSuccess)
  },
  methods: {
    indexMethod(index) {
      return (this.currentPage - 1) * this.pageSize + index + 1
    },
    async fetchData() {
      this.loading = true
      try {
        const res = await getAssetSoftwareList({
          page: this.currentPage,
          page_size: this.pageSize,
          search: this.search || undefined,
          software_ids: this.selectedSoftwareFilter.length > 0 ? this.selectedSoftwareFilter.join(',') : undefined
        })
        this.list = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
        this.$nextTick(() => this.calcTableHeight())
      }
    },
    async fetchAllSoftware() {
      try {
        const res = await getApprovedSoftware({ all: true })
        this.allSoftware = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    handleSizeChange(size) {
      this.pageSize = size
      this.currentPage = 1
      this.fetchData()
    },
    handlePageChange(page) {
      this.currentPage = page
      this.fetchData()
    },
    handleSearch() {
      this.currentPage = 1
      this.fetchData()
    },
    handleResetFilter() {
      this.search = ''
      this.selectedSoftwareFilter = []
      this.currentPage = 1
      this.fetchData()
    },
    async handleEdit(row) {
      this.editRow = row
      try {
        const res = await getAssetSoftwareLinks(row.id)
        this.selectedSoftwareIds = res.data || []
      } catch (e) {
        this.selectedSoftwareIds = []
        console.error(e)
      }
      this.editDialogVisible = true
    },
    async handleSaveLinks() {
      this.submitting = true
      try {
        const dualToken = await this.$refs.dualControl.open()
        await updateAssetSoftwareLinks(this.editRow.id, this.selectedSoftwareIds, dualToken)
        
        // 关闭弹窗并刷新数据
        this.editDialogVisible = false
        this.fetchData()
        
        this.$message.success('关联更新成功')
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      } finally {
        this.submitting = false
      }
    },
    locateAndOpen(assetId) {
      console.log('[AutoLocate] 开始查找资产:', assetId)
      
      // 清除筛选条件
      this.search = ''
      this.selectedSoftwareFilter = []
      this.currentPage = 1
      
      // 首次查找
      let row = this.list.find(item => String(item.id) === String(assetId))
      
      if (row) {
        // 找到则直接打开
        console.log('[AutoLocate] 找到资产，准备打开对话框')
        this.$nextTick(() => {
          this.handleEdit(row)
        })
      } else {
        // 未找到则刷新数据再尝试
        console.log('[AutoLocate] 首次未找到，刷新数据...')
        this.fetchData().then(() => {
          const updatedRow = this.list.find(item => String(item.id) === String(assetId))
          if (updatedRow) {
            console.log('[AutoLocate] 刷新后找到资产，打开对话框')
            this.$nextTick(() => {
              this.handleEdit(updatedRow)
            })
          } else {
            console.warn('[AutoLocate] 刷新后仍未找到资产 ID:', assetId)
            this.$message.warning(`未找到资产 ID${assetId}，请确认资产已创建成功`)
          }
        }).catch(err => {
          console.error('定位资产失败:', err)
          this.$message.error('定位资产失败，请手动查找')
        })
      }
    },
    async handleExport() {
      this.exporting = true
      try {
        const res = await exportPatchUpdateRecord()
        // 检查是否返回了错误 JSON（blob 情况下需要转换）
        if (res instanceof Blob) {
          const link = document.createElement('a')
          link.href = URL.createObjectURL(res)
          const now = new Date()
          const yearMonth = `${now.getFullYear()}年${now.getMonth() + 1}月`
          link.download = `第三方应用补丁更新记录表 (${yearMonth}).xlsx`
          document.body.appendChild(link)
          link.click()
          document.body.removeChild(link)
          URL.revokeObjectURL(link.href)
          this.$message.success('导出成功')
        } else {
          this.$message.warning('当前没有需要更新的软件')
        }
      } catch (e) {
        console.error('导出失败:', e)
        this.$message.error('导出失败，请重试')
      } finally {
        this.exporting = false
      }
    },
    // 处理稍后点击“稍后处理”按钮
    handleSkipAsset() {
      this.showLinkPrompt = false
      this.newAssetId = null
      this.newAssetComputerName = ''
    },
    // 处理点击“立即前往”按钮
    handleGoToAsset() {
      this.showLinkPrompt = false
      const assetId = this.newAssetId
      const assetName = this.newAssetComputerName
      this.newAssetId = null
      this.newAssetComputerName = ''
        
      // 延迟跳转确保弹窗关闭完成（不再清理参数，避免 NavigationDuplicate）
      this.$nextTick(() => {
        setTimeout(() => {
          // 跳转到资产对应表页面并自动定位
          this.$router.push({
            name: 'AssetSoftware',
            query: { asset_id: assetId, auto_open: 'true' }
          })
        }, 300)
      })
    },
    // 接收子组件的 success 事件
    handleChildSuccess(newAssetData) {
      if (newAssetData && newAssetData.id) {
        this.newAssetId = newAssetData.id
        this.newAssetComputerName = newAssetData.computer_name || '新资产'
        this.showLinkPrompt = true
      }
    },
    // 处理自动打开（通过 URL 参数）
    handleAutoOpen(assetId) {
      // 立即清理查询参数，避免后续重复触发
      this.$router.replace({ query: {} })
      
      // 数据应该已经加载完成，直接定位（无需等待）
      console.log('[AutoLocate] 开始定位资产 ID:', assetId)
      this.locateAndOpen(assetId)
    }
  },
  beforeDestroy() {
    // 清除子组件事件监听
    this.$off('child-success', this.handleChildSuccess)
  }
}
</script>

<style scoped>
.asset-software-list {
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
  margin-bottom: 20px;
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

.table-card {
}

.table-wrapper {
}

.filter-bar .el-button {
  border-radius: 10px;
}

.header-actions .el-button--success {
  border-radius: 10px;
}

/* 软件列表弹窗样式 */
/* 动态控制列表高度：大屏上限 400px，小屏随视窗高度收缩，保证弹窗不超出视窗 */
.software-list-wrapper {
  max-height: min(400px, calc(100vh - 400px));
  overflow-y: auto;
  overflow-x: hidden;
  padding-right: 4px;
}

.software-list :deep(.el-checkbox) {
  display: flex;
  align-items: flex-start;
  margin-bottom: 8px;
  margin-right: 0;
}

.software-list :deep(.el-checkbox__label) {
  flex: 1;
  white-space: normal;
  word-break: break-word;
  line-height: 1.5;
}

.empty-state {
  text-align: center;
  color: #999;
  padding: 20px;
}

/* 关联提示弹窗样式 */
.link-prompt-dialog .prompt-content {
  padding: 16px 0;
  text-align: center;
}

.link-prompt-dialog .prompt-text {
  margin: 0 0 12px 0;
  color: #606266;
  font-size: 14px;
  line-height: 1.6;
}
</style>
