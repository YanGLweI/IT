<template>
  <el-dialog
    title="自定义重复规则"
    :visible.sync="dialogVisible"
    width="400px"
    :close-on-click-modal="false"
    append-to-body
    custom-class="repeat-rule-dialog"
    @close="handleClose"
  >
    <el-form label-width="80px">
      <el-form-item label="重复频率">
        <div style="display: flex; align-items: center; gap: 8px">
          <span>每</span>
          <el-input-number v-model="form.interval" :min="1" :max="999" size="small" style="width: 100px" />
          <el-select v-model="form.unit" size="small" style="width: 100px">
            <el-option label="天" value="days" />
            <el-option label="周" value="weeks" />
            <el-option label="个月" value="months" />
            <el-option label="年" value="years" />
          </el-select>
        </div>
      </el-form-item>

      <el-form-item label="结束条件">
        <el-radio-group v-model="endType" size="small">
          <el-radio label="never">永不结束</el-radio>
          <el-radio label="date">结束于日期</el-radio>
          <el-radio label="count">重复次数</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item v-if="endType === 'date'" label="结束日期">
        <el-date-picker
          v-model="form.endDate"
          type="date"
          placeholder="选择结束日期"
          value-format="yyyy-MM-dd"
          size="small"
          style="width: 100%"
        />
      </el-form-item>

      <el-form-item v-if="endType === 'count'" label="重复次数">
        <el-input-number v-model="form.occurrences" :min="1" :max="999" size="small" style="width: 200px" />
        <span style="margin-left: 8px; color: #909399">次</span>
      </el-form-item>

      <div class="rule-preview">
        <i class="el-icon-info"></i>
        预览：每 {{ form.interval }} {{ unitName }} 重复一次{{ endPreview }}
      </div>
    </el-form>

    <div slot="footer">
      <el-button size="small" @click="handleClose">取消</el-button>
      <el-button size="small" type="primary" @click="handleConfirm">确定</el-button>
    </div>
  </el-dialog>
</template>

<script>
export default {
  name: 'RepeatRuleDialog',
  props: {
    visible: { type: Boolean, default: false },
    rule: { type: Object, default: null }
  },
  data() {
    return {
      form: {
        interval: 1,
        unit: 'weeks',
        endDate: '',
        occurrences: null
      },
      endType: 'never'
    }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(val) { this.$emit('close') }
    },
    unitName() {
      const map = { days: '天', weeks: '周', months: '个月', years: '年' }
      return map[this.form.unit] || this.form.unit
    },
    endPreview() {
      if (this.endType === 'date' && this.form.endDate) {
        return `，直到 ${this.form.endDate}`
      } else if (this.endType === 'count' && this.form.occurrences) {
        return `，共 ${this.form.occurrences} 次`
      }
      return ''
    }
  },
  watch: {
    visible(val) {
      if (val && this.rule) {
        this.form.interval = this.rule.interval || 1
        this.form.unit = this.rule.unit || 'weeks'
        if (this.rule.endDate) {
          this.endType = 'date'
          this.form.endDate = this.rule.endDate
        } else if (this.rule.occurrences) {
          this.endType = 'count'
          this.form.occurrences = this.rule.occurrences
        } else {
          this.endType = 'never'
        }
      } else if (val) {
        this.form = { interval: 1, unit: 'weeks', endDate: '', occurrences: null }
        this.endType = 'never'
      }
    }
  },
  methods: {
    handleConfirm() {
      const rule = {
        type: 'custom',
        interval: this.form.interval,
        unit: this.form.unit
      }
      if (this.endType === 'date' && this.form.endDate) {
        rule.endDate = this.form.endDate
      } else if (this.endType === 'count' && this.form.occurrences) {
        rule.occurrences = this.form.occurrences
      }
      this.$emit('confirm', rule)
    },
    handleClose() {
      this.$emit('close')
    }
  }
}
</script>

<style>
.repeat-rule-dialog.el-dialog {
  border-radius: 16px;
  overflow: hidden;
}

.repeat-rule-dialog .el-dialog__header {
  padding: 16px 20px;
  border-bottom: 1px solid #f1f5f9;
  margin: 0;
}

.repeat-rule-dialog .el-dialog__title {
  font-size: 15px;
  font-weight: 600;
  color: #1e293b;
}

.repeat-rule-dialog .el-dialog__body {
  padding: 20px;
}

.repeat-rule-dialog .el-dialog__footer {
  padding: 12px 20px;
  border-top: 1px solid #f1f5f9;
}
</style>

<style scoped>
.rule-preview {
  margin-top: 16px;
  padding: 10px 14px;
  background: #f8fafc;
  border-radius: 8px;
  font-size: 13px;
  color: #475569;
  border: 1px solid #f1f5f9;
}

.rule-preview i {
  margin-right: 6px;
  color: #3b82f6;
}
</style>
