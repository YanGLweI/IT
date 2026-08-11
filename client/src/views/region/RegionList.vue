<template>
  <div class="region-list">
    <!-- 页面头部 -->
    <div class="page-header">
      <div class="header-left">
        <h2 class="page-title">区域管理</h2>
        <p class="page-subtitle">管理公司资产区域划分，支持按区域分类资产</p>
      </div>
      <div class="header-actions">
        <el-button type="primary" size="small" icon="el-icon-plus" @click="handleAdd">新增区域</el-button>
      </div>
    </div>

    <div class="table-card" ref="tableCard">
      <div class="table-wrapper">
      <el-table :data="regions" stripe :max-height="tableMaxHeight">
        <el-table-column label="#" width="130" align="center">
          <template slot-scope="{ $index }">
            <div class="sort-btns">
              <el-button size="mini" type="text" icon="el-icon-arrow-up"
                :disabled="$index === 0" @click="moveRegion(regions[$index], 'up')" />
              <el-button size="mini" type="text" icon="el-icon-arrow-down"
                :disabled="$index === regions.length - 1" @click="moveRegion(regions[$index], 'down')" />
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="name" label="区域名称" />
        <el-table-column prop="description" label="描述" />
        <el-table-column label="子网" width="200" align="center">
          <template slot-scope="scope">
            <span v-if="scope.row.subnet" style="font-family: monospace">{{ scope.row.subnet }}</span>
            <span v-else style="color: #999">—</span>
          </template>
        </el-table-column>
        <el-table-column label="网络等级" width="160" align="center">
          <template slot-scope="scope">
            <el-tag v-if="scope.row.network_level === 1" size="small">一级（互联网）</el-tag>
            <el-tag v-else-if="scope.row.network_level === 2" type="success" size="small">二级</el-tag>
            <el-tag v-else-if="scope.row.network_level === 3" type="warning" size="small">三级</el-tag>
            <el-tag v-else-if="scope.row.network_level === 4" size="small" style="background:#E6A23C;border-color:#E6A23C;color:#fff">四级</el-tag>
            <el-tag v-else-if="scope.row.network_level === 5" type="danger" size="small">五级（高安全区）</el-tag>
            <span v-else style="color:#999">未设置</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" align="center" fixed="right">
          <template slot-scope="scope">
            <el-button size="mini" @click="handleEdit(scope.row)">编辑</el-button>
            <el-button size="mini" type="danger" @click="handleDelete(scope.row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>
      </div>
    </div>

    <el-dialog class="vault-dialog" :title="dialogTitle" :visible.sync="dialogVisible" width="500px" :close-on-click-modal="false">
      <el-form :model="form" :rules="rules" ref="formRef" label-width="80px">
        <el-form-item label="区域名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入区域名称" />
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="form.description" type="textarea" :rows="3" placeholder="请输入描述" />
        </el-form-item>
        <el-form-item label="网络等级">
          <el-select v-model="form.network_level" placeholder="请选择网络等级" style="width: 100%">
            <el-option label="一级（互联网）" :value="1" />
            <el-option label="二级" :value="2" />
            <el-option label="三级" :value="3" />
            <el-option label="四级" :value="4" />
            <el-option label="五级（高安全区）" :value="5" />
          </el-select>
        </el-form-item>
        <el-form-item label="子网地址" prop="subnet_ip">
          <el-input v-model="form.subnet_ip" placeholder="请输入子网地址，如 192.168.1.0" />
        </el-form-item>
        <el-form-item label="掩码" prop="netmask">
          <el-input v-model="form.netmask" placeholder="简化掩码如 24，或标准掩码如 255.255.255.0" />
          <div class="mask-hint" :class="maskHintClass">{{ maskHint }}</div>
        </el-form-item>
      </el-form>
      <span slot="footer">
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </span>
    </el-dialog>

    <!-- 双控验证弹窗 -->
    <DualControlDialog ref="dualControl" />
  </div>
</template>

<script>
import { getRegions, createRegion, updateRegion, deleteRegion, reorderRegion } from '@/api/region'
import DualControlDialog from '@/components/DualControlDialog.vue'
import tableHeightMixin from '@/mixins/table-height'

export default {
  name: 'RegionList',
  components: { DualControlDialog },
  mixins: [tableHeightMixin],
  data() {
    return {
      regions: [],
      dialogVisible: false,
      dialogTitle: '新增区域',
      form: { name: '', description: '', network_level: 0, subnet_ip: '', netmask: '' },
      rules: {
        name: [{ required: true, message: '请输入区域名称', trigger: 'blur' }],
        subnet_ip: [{ validator: this.validateSubnetIp, trigger: 'blur' }],
        netmask: [{ validator: this.validateNetmask, trigger: 'blur' }]
      }
    }
  },
  computed: {
    maskHint() {
      const mask = (this.form.netmask || '').trim()
      if (!mask) return ''
      if (/^\d{1,2}$/.test(mask)) {
        const n = Number(mask)
        if (n >= 0 && n <= 32) return `标准掩码：${this.prefixToMask(n)}`
        return '掩码格式不正确（简化掩码为 0-32）'
      }
      const prefix = this.parseNetmask(mask)
      if (prefix === null) return '掩码格式不正确，如 24 或 255.255.255.0'
      return `简化掩码：/${prefix}`
    },
    maskHintClass() {
      return this.maskHint.startsWith('掩码格式不正确') ? 'mask-hint-error' : ''
    }
  },
  mounted() {
    this.fetchData()
  },
  methods: {
    async fetchData() {
      try {
        const res = await getRegions()
        this.regions = res.data || []
      } catch (e) {
        console.error(e)
      } finally {
        this.$nextTick(() => this.calcTableHeight())
      }
    },
    handleAdd() {
      this.dialogTitle = '新增区域'
      this.form = { name: '', description: '', network_level: 0, subnet_ip: '', netmask: '' }
      this.dialogVisible = true
    },
    handleEdit(row) {
      this.dialogTitle = '编辑区域'
      this.form = { name: row.name, description: row.description, network_level: row.network_level || 0, subnet_ip: '', netmask: '' }
      if (row.subnet) {
        const idx = row.subnet.lastIndexOf('/')
        if (idx > -1) {
          this.form.subnet_ip = row.subnet.slice(0, idx)
          this.form.netmask = row.subnet.slice(idx + 1)
        }
      }
      this.form.id = row.id
      this.dialogVisible = true
    },
    // 校验 IP 地址（四段 0-255）
    isValidIp(ip) {
      const parts = ip.split('.')
      if (parts.length !== 4) return false
      return parts.every(p => /^\d{1,3}$/.test(p) && Number(p) >= 0 && Number(p) <= 255)
    },
    // 掩码转 CIDR 前缀：纯数字 0-32 → 前缀；合法点分十进制掩码 → 前缀；非法返回 null
    parseNetmask(mask) {
      const m = (mask || '').trim()
      if (m === '') return null
      if (/^\d{1,2}$/.test(m)) {
        const n = Number(m)
        return n >= 0 && n <= 32 ? n : null
      }
      const parts = m.split('.')
      if (parts.length !== 4) return null
      const bytes = []
      for (const p of parts) {
        if (!/^\d{1,3}$/.test(p)) return null
        const b = Number(p)
        if (b < 0 || b > 255) return null
        bytes.push(b)
      }
      // 校验二进制位为连续 1 后跟连续 0（如 255.0.255.0 非法）
      let seenZero = false
      let prefix = 0
      for (const b of bytes) {
        for (let i = 7; i >= 0; i--) {
          const bit = (b >> i) & 1
          if (bit === 1) {
            if (seenZero) return null
            prefix++
          } else {
            seenZero = true
          }
        }
      }
      return prefix
    },
    // CIDR 前缀转标准掩码（如 24 → 255.255.255.0）
    prefixToMask(prefix) {
      const mask = []
      for (let i = 0; i < 4; i++) {
        const bits = Math.max(0, Math.min(8, prefix - i * 8))
        mask.push(bits === 0 ? 0 : (0xff << (8 - bits)) & 0xff)
      }
      return mask.join('.')
    },
    // 组装提交用的 CIDR 字符串（两框同空返回空串）
    buildSubnet() {
      const ip = (this.form.subnet_ip || '').trim()
      const mask = (this.form.netmask || '').trim()
      if (!ip && !mask) return ''
      const prefix = this.parseNetmask(mask)
      return `${ip}/${prefix}`
    },
    validateSubnetIp(rule, value, callback) {
      const ip = (value || '').trim()
      const mask = (this.form.netmask || '').trim()
      if (!ip && !mask) { callback(); return }
      if (!ip || !mask) {
        callback(new Error('子网地址与掩码需同时填写'))
        return
      }
      if (!this.isValidIp(ip)) {
        callback(new Error('子网地址格式不正确，如 192.168.1.0'))
        return
      }
      callback()
    },
    validateNetmask(rule, value, callback) {
      const ip = (this.form.subnet_ip || '').trim()
      const mask = (value || '').trim()
      if (!ip && !mask) { callback(); return }
      if (!ip || !mask) {
        callback(new Error('子网地址与掩码需同时填写'))
        return
      }
      if (this.parseNetmask(mask) === null) {
        callback(new Error('掩码格式不正确，如 24 或 255.255.255.0'))
        return
      }
      callback()
    },
    handleSubmit() {
      this.$refs.formRef.validate(async valid => {
        if (!valid) return
        try {
          const dualToken = await this.$refs.dualControl.open()
          const payload = {
            name: this.form.name,
            description: this.form.description,
            network_level: this.form.network_level,
            subnet: this.buildSubnet()
          }
          if (this.form.id) {
            await updateRegion(this.form.id, payload, dualToken)
            this.$message.success('更新成功')
          } else {
            await createRegion(payload, dualToken)
            this.$message.success('创建成功')
          }
          this.dialogVisible = false
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        }
      })
    },
    handleDelete(row) {
      this.$confirm('确定要删除该区域吗？', '提示', { type: 'warning' }).then(async () => {
        try {
          const dualToken = await this.$refs.dualControl.open()
          await deleteRegion(row.id, dualToken)
          this.$message.success('删除成功')
          this.fetchData()
        } catch (e) {
          if (e.message !== 'canceled') console.error(e)
        }
      }).catch(() => {})
    },
    async moveRegion(row, direction) {
      try {
        await reorderRegion({ id: row.id, direction })
        await this.fetchData()
      } catch (e) {
        this.$message.error(e.response?.data?.message || '移动失败')
      }
    }
  }
}
</script>

<style scoped>
.region-list {
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

/* --- 页面头部 --- */
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

/* --- 掩码格式提示 --- */
.mask-hint {
  font-size: 12px;
  line-height: 18px;
  color: #67c23a;
  margin-top: 2px;
}
.mask-hint-error {
  color: #f56c6c;
}

/* --- 排序按钮 --- */
.sort-btns {
  display: flex;
  gap: 2px;
  justify-content: center;
}

/* --- 主按钮 --- */
.header-actions .el-button--primary {
  background: #3b82f6;
  border: none;
  border-radius: 10px;
  padding: 9px 18px;
  font-size: 13px;
  font-weight: 500;
}
.header-actions .el-button--primary:hover {
  background: #2563eb;
}
</style>
