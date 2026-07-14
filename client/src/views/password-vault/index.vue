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
          @click="selectCategory(cat.id, false)"
          @contextmenu.prevent="showContextMenu($event, cat)">
          <svg-icon :name="cat.icon || 'server'" :size="16" />
          <span class="nav-cat-name">{{ cat.name }}</span>
          <span class="badge">{{ cat.entry_count }}</span>
        </div>
      </div>

      <!-- 右键上下文菜单 -->
      <div v-show="contextMenuVisible" class="context-menu" :style="{ left: contextMenuX + 'px', top: contextMenuY + 'px' }">
        <div class="context-menu-item" @click="handleSortCategory(contextMenuCat, 'up'); hideContextMenu()">
          <i class="el-icon-arrow-up" />
          <span>上移</span>
        </div>
        <div class="context-menu-item" @click="handleSortCategory(contextMenuCat, 'down'); hideContextMenu()">
          <i class="el-icon-arrow-down" />
          <span>下移</span>
        </div>
        <div class="context-menu-divider" v-if="!contextMenuCat || !contextMenuCat.is_preset" />
        <div v-if="contextMenuCat && !contextMenuCat.is_preset" class="context-menu-item" @click="openCategoryDialog(contextMenuCat); hideContextMenu()">
          <i class="el-icon-edit" />
          <span>编辑</span>
        </div>
        <div v-if="contextMenuCat && !contextMenuCat.is_preset" class="context-menu-item danger" @click="handleDeleteCategory(contextMenuCat); hideContextMenu()">
          <i class="el-icon-delete" />
          <span>删除</span>
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
        <el-table-column prop="username" label="账号" min-width="140" show-overflow-tooltip>
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
        <el-table-column prop="created_by" label="创建人" width="100">
          <template slot-scope="{ row }">
            <el-tag size="mini">{{ row.created_by }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="updated_by" label="更新人" width="100">
          <template slot-scope="{ row }">
            <el-tag size="mini">{{ row.updated_by }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="更新时间" width="160">
          <template slot-scope="{ row }">
            <span>{{ formatDate(row.updated_at) }}</span>
          </template>
        </el-table-column>
        <el-table-column label="收藏" width="50" align="center">
          <template slot-scope="{ row }">
            <i :class="row.is_starred ? 'el-icon-star-on starred' : 'el-icon-star-off'" style="cursor: pointer; font-size: 16px;" @click="handleToggleStar(row)" />
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template slot-scope="{ row }">
            <el-button type="text" size="medium" icon="el-icon-lock" @click="openUnlockDialog(row)" title="查看密码" />
            <el-button type="text" size="medium" icon="el-icon-edit" @click="openEntryDialog(row)" title="编辑" />
            <el-button type="text" size="medium" icon="el-icon-delete" style="color: #cf222e" @click="handleDelete(row)" title="删除" />
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <div class="pagination-wrapper">
        <el-pagination background layout="total, prev, pager, next, sizes" :total="total" :page-size.sync="pageSize" :current-page.sync="currentPage" @current-change="loadEntries" @size-change="loadEntries" :page-sizes="[10, 20, 50]" />
      </div>
    </div>

    <!-- 对话框 -->
    <DualControlDialog ref="dualControl" />
    <UnlockDialog ref="unlockDialog" />
    <CategoryDialog ref="categoryDialog" :categories="categories" @saved="loadCategories" />
    <PasswordEntryDialog ref="entryDialog" :categories="categories" @saved="loadEntries" />
  </div>
</template>

<script>
import SvgIcon from '@/components/SvgIcon.vue'
import DualControlDialog from '@/components/DualControlDialog.vue'
import UnlockDialog from './UnlockDialog.vue'
import CategoryDialog from './CategoryDialog.vue'
import PasswordEntryDialog from './PasswordEntryDialog.vue'
import { getPasswordCategories, getPasswordEntries, deletePasswordEntry, togglePasswordEntryStar, deletePasswordCategory, sortPasswordCategory } from '@/api/password_vault'

export default {
  name: 'PasswordVault',
  components: { SvgIcon, DualControlDialog, UnlockDialog, CategoryDialog, PasswordEntryDialog },
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
      tableLoading: false,
      contextMenuVisible: false,
      contextMenuX: 0,
      contextMenuY: 0,
      contextMenuCat: null
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
    document.addEventListener('click', this.hideContextMenu)
    document.addEventListener('keydown', this.handleKeydown)
  },
  beforeDestroy() {
    document.removeEventListener('click', this.hideContextMenu)
    document.removeEventListener('keydown', this.handleKeydown)
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
      const original = row.is_starred
      row.is_starred = !row.is_starred
      try {
        await togglePasswordEntryStar(row.id, row.is_starred)
      } catch (e) {
        row.is_starred = original
        this.$message.error('操作失败')
      }
    },
    async handleDelete(row) {
      try {
        await this.$confirm(`确定删除「${row.name}」吗？`, '确认删除', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deletePasswordEntry(row.id, dualToken)
        this.$message.success('删除成功')
        this.loadEntries()
        this.loadCategories()
      } catch (e) {
        if (e.message !== 'canceled' && e !== 'cancel' && e.message !== 'cancel') {
          this.$message.error(e.response?.data?.message || '删除失败')
        }
      }
    },
    async handleDeleteCategory(cat) {
      try {
        await this.$confirm(`确定删除分类「${cat.name}」吗？`, '确认删除', { type: 'warning' })
        const dualToken = await this.$refs.dualControl.open()
        await deletePasswordCategory(cat.id, dualToken)
        this.$message.success('删除成功')
        if (this.selectedCategory === cat.id) {
          this.selectedCategory = null
          this.loadEntries()
        }
        this.loadCategories()
      } catch (e) {
        if (e.message !== 'canceled' && e !== 'cancel' && e.message !== 'cancel') {
          this.$message.error(e.response?.data?.message || '删除失败')
        }
      }
    },
    async handleSortCategory(cat, direction) {
      try {
        const res = await sortPasswordCategory(cat.id, direction)
        if (res && res.code === 200 && res.message) {
          this.$message.info(res.message)
        }
        this.loadCategories()
      } catch (e) {
        this.$message.error(e.response?.data?.message || '操作失败')
      }
    },
    handleSortChange() {
      this.loadEntries()
    },
    maskUsername(username) {
      if (!username || username.length <= 2) return username
      return username[0] + '***' + username[username.length - 1]
    },
    copyText(text) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
          this.$message.success('已复制')
        }).catch(() => {
          this.fallbackCopy(text)
        })
      } else {
        this.fallbackCopy(text)
      }
    },
    fallbackCopy(text) {
      const ta = document.createElement('textarea')
      ta.value = text
      ta.style.position = 'fixed'
      ta.style.left = '-9999px'
      document.body.appendChild(ta)
      ta.select()
      try {
        document.execCommand('copy')
        this.$message.success('已复制')
      } catch (e) {
        this.$message.error('复制失败，请手动复制')
      }
      document.body.removeChild(ta)
    },
    formatDate(dateStr) {
      if (!dateStr) return '-'
      const d = new Date(dateStr)
      const pad = n => String(n).padStart(2, '0')
      return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())} ${pad(d.getHours())}:${pad(d.getMinutes())}`
    },
    showContextMenu(e, cat) {
      this.contextMenuCat = cat
      this.contextMenuX = e.clientX
      this.contextMenuY = e.clientY
      this.contextMenuVisible = true
      this.$nextTick(() => {
        const menu = this.$el.querySelector('.context-menu')
        if (menu) {
          const rect = menu.getBoundingClientRect()
          if (rect.right > window.innerWidth) {
            this.contextMenuX = window.innerWidth - rect.width - 8
          }
          if (rect.bottom > window.innerHeight) {
            this.contextMenuY = window.innerHeight - rect.height - 8
          }
        }
      })
    },
    hideContextMenu() {
      this.contextMenuVisible = false
    },
    handleKeydown(e) {
      if (e.key === 'Escape') this.hideContextMenu()
    }
  }
}
</script>

<style scoped>
.password-vault-page {
  display: flex;
  height: calc(100vh - 100px);
  background: #f0f2f5;
  border-radius: 12px;
  overflow: hidden;
}

/* 左侧栏 */
.sidebar {
  width: 240px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: #ffffff;
  margin-right: 5px;
  border-radius: 12px;
  border-right: 1px solid #e2e8f0;
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
  background: #e8ecf1;
  color: #1e293b;
}
.nav-item.active {
  background: linear-gradient(135deg, #dbeafe 0%, #e8f0fe 100%);
  color: #2563eb;
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
  color: #64748b;
  background: #e2e8f0;
  padding: 2px 8px;
  border-radius: 12px;
  font-weight: 500;
}
.nav-item .nav-cat-name {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.nav-divider {
  height: 1px;
  background: #e2e8f0;
  margin: 8px 12px;
}
.sidebar-footer {
  padding: 16px;
  border-top: 1px solid #e2e8f0;
}
.sidebar-footer .el-button {
  color: #475569;
  font-size: 13px;
}
.sidebar-footer .el-button:hover {
  color: #2563eb;
}

/* 右侧主区域 */
.main-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding: 20px;
  background: #ffffff;
  border-radius: 12px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
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
  background: #f1f5f9;
  color: #475569;
  font-weight: 600;
  font-size: 12px;
}
.password-vault-page ::v-deep .el-table td {
  color: #1e293b;
}
.password-vault-page ::v-deep .el-table--striped .el-table__body tr.el-table__row--striped td {
  background: #f8fafc;
}
.password-vault-page ::v-deep .el-table--striped .el-table__body tr:hover > td {
  background: #f1f5f9 !important;
}
.password-vault-page ::v-deep .el-tag {
  border-radius: 6px;
  border: none;
  font-size: 12px;
}
.password-vault-page ::v-deep .el-tag--info {
  background: #e2e8f0;
  color: #475569;
}

/* 右键上下文菜单 */
.context-menu {
  position: fixed;
  z-index: 9999;
  min-width: 140px;
  background: #ffffff;
  border-radius: 10px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12), 0 1px 4px rgba(0, 0, 0, 0.08);
  padding: 6px 0;
  border: 1px solid #e2e8f0;
}
.context-menu-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 9px 18px;
  font-size: 13px;
  color: #475569;
  cursor: pointer;
  transition: background 0.15s ease;
}
.context-menu-item:hover {
  background: #f1f5f9;
  color: #1e293b;
}
.context-menu-item.danger {
  color: #cf222e;
}
.context-menu-item.danger:hover {
  background: #fef2f2;
}
.context-menu-item i {
  font-size: 14px;
  width: 16px;
  text-align: center;
}
.context-menu-divider {
  height: 1px;
  background: #e2e8f0;
  margin: 4px 0;
}
</style>
