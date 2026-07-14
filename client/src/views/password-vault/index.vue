<template>
  <div class="password-vault-page">
    <!-- 左侧栏 -->
    <div class="sidebar">
      <div class="sidebar-search">
        <el-input v-model="categorySearch" placeholder="搜索分类..." size="small" prefix-icon="el-icon-search" clearable />
      </div>
      <div class="sidebar-nav">
        <div class="nav-item" :class="{ active: !selectedCategory && !showStarred }" @click="selectCategory(null, false)">
          <i class="el-icon-folder" />
          <span>全部</span>
          <span class="badge">{{ totalCount }}</span>
        </div>
        <div class="nav-item" :class="{ active: showStarred }" @click="selectCategory(null, true)">
          <i class="el-icon-star-off" />
          <span>收藏</span>
        </div>
        <div class="nav-divider" />
        <div class="nav-item" v-for="cat in filteredCategories" :key="cat.id"
          :class="{ active: selectedCategory === cat.id && !showStarred }"
          @click="selectCategory(cat.id, false)">
          <i class="el-icon-folder" />
          <span>{{ cat.name }}</span>
          <span class="badge">{{ cat.entry_count }}</span>
        </div>
      </div>
      <div class="sidebar-footer">
        <el-button type="text" icon="el-icon-plus" @click="openCategoryDialog(null)">添加分类</el-button>
      </div>
    </div>

    <!-- 右侧主区域 -->
    <div class="main-content">
      <!-- 工具栏 -->
      <div class="toolbar">
        <el-input v-model="keyword" placeholder="搜索条目名称/账号/URL..." size="small" prefix-icon="el-icon-search" clearable style="width: 280px" @clear="loadEntries" @keyup.enter.native="loadEntries" />
        <el-button size="small" type="primary" icon="el-icon-plus" @click="openEntryDialog(null)">新增条目</el-button>
      </div>

      <!-- 表格 -->
      <el-table :data="entries" v-loading="tableLoading" stripe style="width: 100%" @sort-change="handleSortChange">
        <el-table-column width="50" align="center">
          <template slot-scope="{ row }">
            <svg-icon :name="row.icon" :size="20" />
          </template>
        </el-table-column>
        <el-table-column prop="name" label="名称" min-width="180" show-overflow-tooltip />
        <el-table-column prop="username" label="账号" width="140" show-overflow-tooltip>
          <template slot-scope="{ row }">
            <span>{{ maskUsername(row.username) }}</span>
            <el-button type="text" size="mini" icon="el-icon-document-copy" @click="copyText(row.username)" style="margin-left: 4px" />
          </template>
        </el-table-column>
        <el-table-column prop="category_name" label="分类" width="110">
          <template slot-scope="{ row }">
            <el-tag size="mini" type="info">{{ row.category_name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="URL/端口" width="220" show-overflow-tooltip>
          <template slot-scope="{ row }">
            <span v-if="row.url">{{ row.url }}<span v-if="row.port">:{{ row.port }}</span></span>
            <span v-else class="text-muted">-</span>
          </template>
        </el-table-column>
        <el-table-column label="收藏" width="50" align="center">
          <template slot-scope="{ row }">
            <i :class="row.is_starred ? 'el-icon-star-on starred' : 'el-icon-star-off'" style="cursor: pointer; font-size: 16px;" @click="handleToggleStar(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="140" align="center">
          <template slot-scope="{ row }">
            <el-button type="text" size="mini" icon="el-icon-lock" @click="openUnlockDialog(row)" title="查看密码" />
            <el-button type="text" size="mini" icon="el-icon-edit" @click="openEntryDialog(row)" title="编辑" />
            <el-button type="text" size="mini" icon="el-icon-delete" style="color: #cf222e" @click="handleDelete(row)" title="删除" />
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination background layout="total, prev, pager, next, sizes" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage" @current-change="loadEntries" @size-change="loadEntries" :page-sizes="[10, 20, 50]" />
      </div>
    </div>

    <!-- 对话框 -->
    <UnlockDialog ref="unlockDialog" />
    <CategoryDialog ref="categoryDialog" :categories="categories" @saved="loadCategories" />
    <PasswordEntryDialog ref="entryDialog" :categories="categories" @saved="loadEntries" />
  </div>
</template>

<script>
import SvgIcon from '@/components/SvgIcon.vue'
import UnlockDialog from './UnlockDialog.vue'
import CategoryDialog from './CategoryDialog.vue'
import PasswordEntryDialog from './PasswordEntryDialog.vue'
import { getPasswordCategories, getPasswordEntries, deletePasswordEntry, togglePasswordEntryStar } from '@/api/password_vault'

export default {
  name: 'PasswordVault',
  components: { SvgIcon, UnlockDialog, CategoryDialog, PasswordEntryDialog },
  data() {
    return {
      categories: [],
      entries: [],
      selectedCategory: null,
      showStarred: false,
      categorySearch: '',
      keyword: '',
      total: 0,
      totalCount: 0,
      currentPage: 1,
      pageSize: 20,
      tableLoading: false
    }
  },
  computed: {
    filteredCategories() {
      if (!this.categorySearch) return this.categories
      const q = this.categorySearch.toLowerCase()
      return this.categories.filter(c => c.name.toLowerCase().includes(q))
    }
  },
  created() {
    this.loadCategories()
    this.loadEntries()
  },
  methods: {
    async loadCategories() {
      try {
        const res = await getPasswordCategories()
        if (res && res.code === 200) {
          this.categories = res.data || []
          this.totalCount = this.categories.reduce((sum, c) => sum + (c.entry_count || 0), 0)
        }
      } catch (e) { /* ignore */ }
    },
    async loadEntries() {
      this.tableLoading = true
      try {
        const params = {
          page: this.currentPage,
          page_size: this.pageSize
        }
        if (this.selectedCategory) params.category_id = this.selectedCategory
        if (this.showStarred) params.is_starred = 'true'
        if (this.keyword) params.keyword = this.keyword

        const res = await getPasswordEntries(params)
        if (res && res.code === 200) {
          this.entries = res.data || []
          this.total = res.total || 0
        }
      } catch (e) { /* ignore */ } finally {
        this.tableLoading = false
      }
    },
    selectCategory(categoryId, starred) {
      this.selectedCategory = categoryId
      this.showStarred = starred
      this.currentPage = 1
      this.loadEntries()
    },
    openCategoryDialog(category) {
      this.$refs.categoryDialog.open(category)
    },
    openEntryDialog(entry) {
      this.$refs.entryDialog.open(entry)
    },
    openUnlockDialog(entry) {
      this.$refs.unlockDialog.open(entry)
    },
    async handleToggleStar(row) {
      try {
        await togglePasswordEntryStar(row.id, !row.is_starred)
        row.is_starred = !row.is_starred
      } catch (e) {
        this.$message.error('操作失败')
      }
    },
    handleDelete(row) {
      this.$confirm(`确定删除「${row.name}」吗？`, '确认删除', { type: 'warning' }).then(async () => {
        try {
          await deletePasswordEntry(row.id)
          this.$message.success('删除成功')
          this.loadEntries()
          this.loadCategories()
        } catch (e) {
          this.$message.error(e.response?.data?.message || '删除失败')
        }
      }).catch(() => {})
    },
    handleSortChange() {
      this.loadEntries()
    },
    maskUsername(username) {
      if (!username || username.length <= 2) return username
      return username[0] + '***' + username[username.length - 1]
    },
    copyText(text) {
      navigator.clipboard.writeText(text)
      this.$message.success('已复制')
    }
  }
}
</script>

<style scoped>
.password-vault-page {
  display: flex;
  height: calc(100vh - 100px);
  background: #fafbfc;
  border-radius: 16px;
  overflow: hidden;
}

/* 左侧栏 */
.sidebar {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  border-right: 1px solid #f0f2f5;
}
.sidebar-search {
  padding: 20px 16px 12px;
}
.sidebar-nav {
  flex: 1;
  overflow-y: auto;
  padding: 0 12px 12px;
}
.nav-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 10px 14px;
  border-radius: 10px;
  cursor: pointer;
  font-size: 14px;
  color: #4a5568;
  transition: all 0.2s ease;
  margin-bottom: 2px;
}
.nav-item:hover {
  background: #f7f8fa;
  color: #2d3748;
}
.nav-item.active {
  background: linear-gradient(135deg, #e8f4fd 0%, #f0f7ff 100%);
  color: #3b82f6;
  font-weight: 500;
}
.nav-item i {
  font-size: 16px;
  opacity: 0.7;
}
.nav-item.active i {
  opacity: 1;
}
.nav-item .badge {
  margin-left: auto;
  font-size: 12px;
  color: #94a3b8;
  background: #f1f5f9;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
}
.nav-item.active .badge {
  background: rgba(59, 130, 246, 0.12);
  color: #3b82f6;
}
.nav-divider {
  height: 1px;
  background: #f0f2f5;
  margin: 8px 12px;
}
.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #f0f2f5;
}
.sidebar-footer .el-button {
  color: #64748b;
  font-size: 13px;
}
.sidebar-footer .el-button:hover {
  color: #3b82f6;
}

/* 右侧主区域 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 24px;
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
.toolbar .el-input {
  border-radius: 10px;
}
.toolbar .el-button {
  border-radius: 10px;
  padding: 8px 16px;
}
.pagination-wrapper {
  margin-top: 20px;
  display: flex;
  justify-content: flex-end;
}
.starred {
  color: #f59e0b;
}
.text-muted {
  color: #94a3b8;
}

/* 表格样式优化 */
.password-vault-page ::v-deep .el-table {
  border-radius: 12px;
  overflow: hidden;
}
.password-vault-page ::v-deep .el-table th {
  background: #f8fafc;
  color: #64748b;
  font-weight: 500;
  font-size: 13px;
}
.password-vault-page ::v-deep .el-table td {
  color: #334155;
}
.password-vault-page ::v-deep .el-table--striped .el-table__body tr.el-table__row--striped td {
  background: #fafbfc;
}
.password-vault-page ::v-deep .el-table--striped .el-table__body tr:hover > td {
  background: #f7f8fa !important;
}
.password-vault-page ::v-deep .el-tag {
  border-radius: 6px;
  border: none;
  background: #f1f5f9;
  color: #64748b;
  font-size: 12px;
}
</style>
