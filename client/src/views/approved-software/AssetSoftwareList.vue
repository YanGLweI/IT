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
      <el-checkbox-group v-model="selectedSoftwareIds">
        <el-checkbox
          v-for="sw in allSoftware"
          :key="sw.id"
          :label="sw.id"
          style="display: block; margin-bottom: 6px"
        >
          {{ sw.name }}
          <span v-if="sw.version" style="color: #999; font-size: 12px">({{ sw.version }})</span>
        </el-checkbox>
      </el-checkbox-group>
      <div v-if="allSoftware.length === 0" style="text-align: center; color: #999; padding: 20px">
        暂无核准软件，请先在"核准软件目录"中添加
      </div>
      <span slot="footer">
        <el-button @click="editDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="handleSaveLinks">保存</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
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
      exporting: false
    }
  },
  mounted() {
    this.fetchData()
    this.fetchAllSoftware()
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
        const res = await getApprovedSoftware()
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
        this.$message.success('关联更新成功')
        this.editDialogVisible = false
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      } finally {
        this.submitting = false
      }
    },
    async handleExport() {
      this.exporting = true
      try {
        const res = await exportPatchUpdateRecord()
        // 检查是否返回了错误JSON（blob情况下需要转换）
        if (res instanceof Blob) {
          const link = document.createElement('a')
          link.href = URL.createObjectURL(res)
          const now = new Date()
          const yearMonth = `${now.getFullYear()}年${now.getMonth() + 1}月`
          link.download = `第三方应用补丁更新记录表(${yearMonth}).xlsx`
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
    }
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
</style>
