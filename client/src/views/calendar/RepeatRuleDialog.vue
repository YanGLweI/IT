<template>
  <el-dialog
    title="自定义重复"
    :visible.sync="dialogVisible"
    :width="dialogWidth"
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
          <el-select v-model="form.unit" size="small" style="width: 100px" @change="handleUnitChange">
            <el-option label="天" value="days" />
            <el-option label="周" value="weeks" />
            <el-option label="个月" value="months" />
            <el-option label="年" value="years" />
          </el-select>
        </div>
      </el-form-item>

      <!-- 周频率：多选周几 -->
      <el-form-item v-if="form.unit === 'weeks'" label="重复日期">
        <div class="weekday-buttons">
          <button
            type="button"
            v-for="(name, idx) in weekdayNames"
            :key="idx"
            class="weekday-btn"
            :class="{
              selected: selectedWeekdays.includes(idx),
              locked: idx === lockedWeekday
            }"
            :disabled="idx === lockedWeekday"
            @click="toggleWeekday(idx)"
          >{{ name }}</button>
        </div>
      </el-form-item>

      <!-- 月频率：按日期/按星期 -->
      <el-form-item v-if="form.unit === 'months'" label="重复规则">
        <el-select v-model="monthRuleType" size="small" style="width: 120px" @change="handleMonthRuleChange">
          <el-option label="按日期" value="day" />
          <el-option label="按星期" value="week" />
        </el-select>
      </el-form-item>

      <!-- 月频率-按日期：日期网格（纯展示） -->
      <el-form-item v-if="form.unit === 'months' && monthRuleType === 'day'" label="">
        <div class="month-day-grid">
          <span
            v-for="d in 31"
            :key="d"
            class="month-day-btn"
            :class="{ selected: d === lockedMonthDay }"
          >{{ d }}</span>
        </div>
      </el-form-item>

      <!-- 月频率-按星期：只读标签 -->
      <el-form-item v-if="form.unit === 'months' && monthRuleType === 'week'" label="">
        <div class="week-label">
          第{{ ordinalNames[lockedWeekOfMonth - 1] }}个 周{{ weekdayNames[lockedWeekday] }}
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
        {{ previewText }}
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
    rule: { type: Object, default: null },
    startDate: { type: Date, default: null }
  },
  data() {
    return {
      form: {
        interval: 1,
        unit: 'weeks',
        endDate: '',
        occurrences: null
      },
      endType: 'never',
      monthRuleType: 'day',
      selectedWeekdays: [],
      lockedWeekday: 1,
      lockedMonthDay: 1,
      lockedWeekOfMonth: 1,
      weekdayNames: ['日', '一', '二', '三', '四', '五', '六'],
      ordinalNames: ['一', '二', '三', '四', '五']
    }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(val) { this.$emit('close') }
    },
    dialogWidth() {
      return (this.form.unit === 'months' && this.monthRuleType === 'day') ? '460px' : '400px'
    },
    previewText() {
      const interval = this.form.interval
      const unitMap = { days: '天', weeks: '周', months: '个月', years: '年' }
      const unitName = unitMap[this.form.unit] || this.form.unit
      let base = `每 ${interval} ${unitName} 重复`

      if (this.form.unit === 'weeks' && this.selectedWeekdays.length > 0) {
        const names = this.selectedWeekdays.map(d => '周' + this.weekdayNames[d]).join('、')
        base += `（${names}）`
      } else if (this.form.unit === 'months') {
        if (this.monthRuleType === 'day') {
          base += `（${this.lockedMonthDay}日）`
        } else {
          base += `（第${this.ordinalNames[this.lockedWeekOfMonth - 1]}个 周${this.weekdayNames[this.lockedWeekday]}）`
        }
      }

      if (this.endType === 'date' && this.form.endDate) {
        base += `，直到 ${this.form.endDate}`
      } else if (this.endType === 'count' && this.form.occurrences) {
        base += `，共 ${this.form.occurrences} 次`
      }
      return '预览：' + base
    }
  },
  watch: {
    visible(val) {
      if (val) {
        this.initFromRule()
      }
    }
  },
  methods: {
    initFromRule() {
      // 根据开始日期计算锁定值
      const sd = this.startDate || new Date()
      this.lockedWeekday = sd.getDay()
      this.lockedMonthDay = sd.getDate()
      this.lockedWeekOfMonth = this.getWeekOfMonth(sd)

      if (this.rule) {
        this.form.interval = this.rule.interval || 1

        // 根据 rule.type 确定 form.unit（monthly_week/monthly_day 没有 unit 字段）
        if (this.rule.type === 'monthly_week' || this.rule.type === 'monthly_day') {
          this.form.unit = 'months'
        } else if (this.rule.type === 'yearly') {
          this.form.unit = 'years'
        } else if (this.rule.type === 'daily') {
          this.form.unit = 'days'
        } else {
          this.form.unit = this.rule.unit || 'weeks'
        }

        // 恢复周选择
        if (this.rule.type === 'workday') {
          this.selectedWeekdays = [1, 2, 3, 4, 5]
        } else if (this.rule.type === 'custom' && this.rule.unit === 'weeks' && this.rule.weekdays) {
          this.selectedWeekdays = [...this.rule.weekdays]
        } else {
          this.selectedWeekdays = [this.lockedWeekday]
        }

        // 恢复月规则类型
        if (this.rule.type === 'monthly_week') {
          this.monthRuleType = 'week'
        } else {
          this.monthRuleType = 'day'
        }

        // 恢复结束条件
        if (this.rule.endDate) {
          this.endType = 'date'
          this.form.endDate = this.rule.endDate
        } else if (this.rule.occurrences) {
          this.endType = 'count'
          this.form.occurrences = this.rule.occurrences
        } else {
          this.endType = 'never'
        }
      } else {
        this.form = { interval: 1, unit: 'weeks', endDate: '', occurrences: null }
        this.endType = 'never'
        this.monthRuleType = 'day'
        this.selectedWeekdays = [this.lockedWeekday]
      }
    },
    getWeekOfMonth(date) {
      const day = date.getDate()
      const weekday = date.getDay()
      const firstDay = new Date(date.getFullYear(), date.getMonth(), 1)
      const firstDayWeekday = firstDay.getDay()
      let firstOccurrence
      if (firstDayWeekday <= weekday) {
        firstOccurrence = 1 + (weekday - firstDayWeekday)
      } else {
        firstOccurrence = 1 + (7 - firstDayWeekday + weekday)
      }
      return Math.floor((day - firstOccurrence) / 7) + 1
    },
    handleUnitChange() {
      if (this.form.unit !== 'weeks') {
        this.selectedWeekdays = [this.lockedWeekday]
      } else if (this.selectedWeekdays.length === 0) {
        this.selectedWeekdays = [this.lockedWeekday]
      }
      if (this.form.unit !== 'months') {
        this.monthRuleType = 'day'
      }
    },
    handleMonthRuleChange() {
      // 切换时保持当前锁定值
    },
    toggleWeekday(idx) {
      if (idx === this.lockedWeekday) return
      const pos = this.selectedWeekdays.indexOf(idx)
      if (pos >= 0) {
        this.selectedWeekdays.splice(pos, 1)
      } else {
        this.selectedWeekdays.push(idx)
        this.selectedWeekdays.sort((a, b) => a - b)
      }
    },
    handleConfirm() {
      let rule

      if (this.form.unit === 'weeks') {
        // 周频率：custom + weekdays
        rule = {
          type: 'custom',
          interval: this.form.interval,
          unit: 'weeks',
          weekdays: [...this.selectedWeekdays].sort((a, b) => a - b)
        }
      } else if (this.form.unit === 'months') {
        if (this.monthRuleType === 'day') {
          // 月按日期
          rule = {
            type: 'monthly_day',
            interval: this.form.interval,
            monthDay: this.lockedMonthDay
          }
        } else {
          // 月按星期
          rule = {
            type: 'monthly_week',
            interval: this.form.interval,
            weekOfMonth: this.lockedWeekOfMonth,
            weekday: this.lockedWeekday
          }
        }
      } else {
        // 天/年：保持原有 custom 逻辑
        rule = {
          type: 'custom',
          interval: this.form.interval,
          unit: this.form.unit
        }
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

/* 周几按钮 */
.weekday-buttons {
  display: flex;
  gap: 6px;
}

.weekday-btn {
  width: 42px;
  height: 34px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #ffffff;
  color: #606266;
  font-size: 13px;
  cursor: pointer;
  transition: all 0.2s;
}

.weekday-btn:hover:not(:disabled) {
  border-color: #3b82f6;
  color: #3b82f6;
}

.weekday-btn.selected {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
}

.weekday-btn.locked {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
  cursor: default;
}

/* 月日期网格 */
.month-day-grid {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  max-width: 380px;
}

.month-day-btn {
  width: 36px;
  height: 32px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #ffffff;
  color: #909399;
  font-size: 13px;
  cursor: default;
}

.month-day-btn.selected {
  background: #3b82f6;
  border-color: #3b82f6;
  color: #ffffff;
  font-weight: 600;
}

/* 按星期标签 */
.week-label {
  display: inline-block;
  padding: 8px 20px;
  background: #f0f5ff;
  border: 1px solid #bfdbfe;
  border-radius: 8px;
  font-size: 14px;
  color: #3b82f6;
  font-weight: 500;
}
</style>
