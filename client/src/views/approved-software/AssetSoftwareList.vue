<template>
  <div class="asset-software-list">
    <el-card>
      <div slot="header">
        <span>资产对应表 — 第三方软件与资产关联</span>
      </div>
      <el-table :data="list" border stripe v-loading="loading">
        <el-table-column type="index" label="序号" width="60" align="center" :index="indexMethod" />
        <el-table-column prop="computer_name" label="计算机名" min-width="150" />
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
        <el-table-column label="操作" width="100" fixed="right">
          <template slot-scope="scope">
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
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

    <!-- 编辑关联软件弹窗 -->
    <el-dialog title="关联核准软件" :visible.sync="editDialogVisible" width="560px" :close-on-click-modal="false">
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
  getApprovedSoftware
} from '@/api/approved_software'
import DualControlDialog from '@/components/DualControlDialog.vue'

export default {
  name: 'AssetSoftwareList',
  components: { DualControlDialog },
  data() {
    return {
      list: [],
      loading: false,
      total: 0,
      currentPage: 1,
      pageSize: 10,
      allSoftware: [],
      editDialogVisible: false,
      editRow: null,
      selectedSoftwareIds: [],
      submitting: false
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
          page_size: this.pageSize
        })
        this.list = res.data || []
        this.total = res.total || 0
      } catch (e) {
        console.error(e)
      } finally {
        this.loading = false
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
    }
  }
}
</script>
