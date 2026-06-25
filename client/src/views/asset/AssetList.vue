<template>
  <div class="asset-list">
    <el-card>
      <div slot="header" style="display: flex; justify-content: space-between; align-items: center">
        <span>IT资产管理</span>
        <el-button type="primary" size="small" icon="el-icon-plus" @click="handleAdd">新增资产</el-button>
      </div>

      <el-tabs v-model="activeTab" @tab-click="handleTabClick">
        <el-tab-pane label="全部" name="all" />
        <el-tab-pane
          v-for="region in regions"
          :key="region.id"
          :label="region.name"
          :name="String(region.id)"
        />
      </el-tabs>

      <el-table
        :data="assets"
        border
        stripe
        @sort-change="handleSortChange"
        v-loading="loading"
      >
        <el-table-column type="index" label="序号" width="60" align="center" :index="indexMethod" />
        <el-table-column prop="computer_name" label="计算机名" sortable="custom" />
        <el-table-column prop="ip_address" label="IP地址" sortable="custom" />
        <el-table-column prop="os_type" label="操作系统" sortable="custom" />
        <el-table-column prop="purpose" label="用途" />
        <el-table-column prop="asset_level" label="资产等级" width="100" sortable="custom">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.asset_level" size="small">{{ scope.row.asset_level }}</el-tag>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80" sortable="custom">
          <template slot-scope="scope">
            <el-tag :type="statusType(scope.row.status)" size="small">{{ scope.row.status }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200">
          <template slot-scope="scope">
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        style="margin-top: 15px; text-align: right"
        background
        layout="total, sizes, prev, pager, next, jumper"
        :total="total"
        :page-size.sync="pageSize"
        :current-page.sync="currentPage"
        :page-sizes="[10, 20, 50, 100]"
        @size-change="handleSizeChange"
        @current-change="handlePageChange"
      />
    </el-card>

    <asset-form
      :visible.sync="formVisible"
      :edit-data="editData"
      :regions="regions"
      @success="fetchData"
      ref="assetForm"
    />
    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getAssets, deleteAsset } from '@/api/asset'
import DualControlDialog from '@/components/DualControlDialog.vue'
import { getRegions } from '@/api/region'
import AssetForm from './AssetForm.vue'

export default {
  name: 'AssetList',
  components: { AssetForm, DualControlDialog },
  data() {
    return {
      assets: [],
      regions: [],
      activeTab: 'all',
      formVisible: false,
      editData: null,
      loading: false,
      // 分页
      total: 0,
      currentPage: 1,
      pageSize: 10,
      // 排序
      sortBy: 'id',
      sortOrder: 'desc'
    }
  },
  mounted() {
    this.fetchData()
    this.fetchRegions()
  },
  methods: {
    async fetchData() {
      this.loading = true
      try {
        const params = {
          page: this.currentPage,
          page_size: this.pageSize,
          sort_by: this.sortBy,
          sort_order: this.sortOrder
        }
        if (this.activeTab !== 'all') {
          params.region_id = this.activeTab
        }
        const res = await getAssets(params)
        this.assets = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
      }
    },
    async fetchRegions() {
      try {
        const res = await getRegions()
        this.regions = res.data || []
      } catch (e) {
        console.error(e)
      }
    },
    handleTabClick() {
      this.currentPage = 1
      this.fetchData()
    },
    handleSortChange({ prop, order }) {
      if (prop && order) {
        this.sortBy = prop
        this.sortOrder = order === 'ascending' ? 'asc' : 'desc'
      } else {
        this.sortBy = 'id'
        this.sortOrder = 'desc'
      }
      this.fetchData()
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
    indexMethod(index) {
      return (this.currentPage - 1) * this.pageSize + index + 1
    },
    handleAdd() {
      this.editData = null
      this.formVisible = true
    },
    handleEdit(row) {
      this.editData = { ...row }
      this.formVisible = true
    },
    async handleDelete(row) {
      try {
        await this.$confirm('确定要删除该资产吗？', '提示', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deleteAsset(row.id, dualToken)
        this.$message.success('删除成功')
        this.fetchData()
      } catch (e) {
        if (e.message !== 'canceled') console.error(e)
      }
    },
    statusType(status) {
      switch (status) {
        case '在用': return 'success'
        case '闲置': return 'warning'
        case '报废': return 'danger'
        default: return 'info'
      }
    }
  }
}
</script>
