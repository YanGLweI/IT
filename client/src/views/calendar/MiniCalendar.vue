<template>
  <div class="mini-calendar">
    <!-- 月份头部 -->
    <div class="mini-cal-header">
      <button class="mini-cal-nav" @click="prevMonth">
        <i class="el-icon-arrow-left"></i>
      </button>
      <span class="mini-cal-title" @click="toggleYearMonthPicker">
        {{ viewMonth.getFullYear() }}年{{ viewMonth.getMonth() + 1 }}月
        <i class="el-icon-arrow-down title-arrow" :class="{ 'is-open': showYearMonthPicker }"></i>
      </span>
      <button class="mini-cal-nav" @click="nextMonth">
        <i class="el-icon-arrow-right"></i>
      </button>
    </div>

    <!-- 年月快速选择下拉框 -->
    <div v-if="showYearMonthPicker" class="year-month-picker" @click.stop>
      <div class="ym-col-years">
        <div
          v-for="y in yearList"
          :key="y"
          class="ym-item"
          :class="{ 'is-active': y === pickerYear }"
          @click="selectPickerYear(y)"
        >
          {{ y }}
          <i class="el-icon-arrow-right ym-arrow"></i>
        </div>
      </div>
      <div class="ym-col-months">
        <div
          v-for="m in monthList"
          :key="m"
          class="ym-item"
          :class="{ 'is-selected': m === pickerMonth && pickerYear === viewMonth.getFullYear() }"
          @click="selectPickerMonth(m)"
        >
          {{ m }}月
          <i v-if="m === pickerMonth && pickerYear === viewMonth.getFullYear()" class="el-icon-check ym-check"></i>
        </div>
      </div>
    </div>

    <!-- 星期表头 -->
    <div class="mini-cal-weekdays">
      <span v-for="name in dayNames" :key="name" class="mini-cal-weekday">{{ name }}</span>
    </div>

    <!-- 日期网格 -->
    <div class="mini-cal-grid">
      <div
        v-for="(day, index) in calendarDays"
        :key="index"
        class="mini-cal-day"
        :class="{
          'other-month': !isCurrentMonth(day),
          'is-today': isToday(day),
          'is-selected': isSelected(day)
        }"
        @click="selectDate(day)"
      >
        {{ day.getDate() }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MiniCalendar',
  props: {
    currentDate: { type: Date, required: true }
  },
  data() {
    return {
      viewMonth: new Date(this.currentDate),
      dayNames: ['一', '二', '三', '四', '五', '六', '日'],
      showYearMonthPicker: false,
      pickerYear: this.currentDate.getFullYear(),
      pickerMonth: this.currentDate.getMonth() + 1
    }
  },
  computed: {
    yearList() {
      const current = new Date().getFullYear()
      const years = []
      for (let y = current - 2; y <= current + 3; y++) {
        years.push(y)
      }
      return years
    },
    monthList() {
      return Array.from({ length: 12 }, (_, i) => i + 1)
    },
    calendarDays() {
      const year = this.viewMonth.getFullYear()
      const month = this.viewMonth.getMonth()
      const firstDay = new Date(year, month, 1)
      const dayOfWeek = firstDay.getDay()
      const startOffset = dayOfWeek === 0 ? 6 : dayOfWeek - 1

      const days = []
      const startDate = new Date(firstDay)
      startDate.setDate(startDate.getDate() - startOffset)

      for (let i = 0; i < 42; i++) {
        const d = new Date(startDate)
        d.setDate(d.getDate() + i)
        days.push(d)
      }
      return days
    }
  },
  watch: {
    currentDate(newVal) {
      this.viewMonth = new Date(newVal)
      this.pickerYear = newVal.getFullYear()
      this.pickerMonth = newVal.getMonth() + 1
    }
  },
  mounted() {
    document.addEventListener('click', this.closeYearMonthPicker)
  },
  beforeDestroy() {
    document.removeEventListener('click', this.closeYearMonthPicker)
  },
  methods: {
    toggleYearMonthPicker(e) {
      e.stopPropagation()
      this.showYearMonthPicker = !this.showYearMonthPicker
      if (this.showYearMonthPicker) {
        this.pickerYear = this.viewMonth.getFullYear()
        this.pickerMonth = this.viewMonth.getMonth() + 1
      }
    },
    closeYearMonthPicker() {
      this.showYearMonthPicker = false
    },
    selectPickerYear(y) {
      this.pickerYear = y
    },
    selectPickerMonth(m) {
      this.pickerMonth = m
      const d = new Date(this.pickerYear, m - 1, 1)
      this.viewMonth = new Date(d)
      this.showYearMonthPicker = false
      this.$emit('month-change', d)
    },
    prevMonth() {
      const d = new Date(this.viewMonth)
      d.setMonth(d.getMonth() - 1)
      this.viewMonth = d
    },
    nextMonth() {
      const d = new Date(this.viewMonth)
      d.setMonth(d.getMonth() + 1)
      this.viewMonth = d
    },
    selectDate(day) {
      this.$emit('date-change', new Date(day))
    },
    isCurrentMonth(day) {
      return day.getMonth() === this.viewMonth.getMonth()
    },
    isToday(day) {
      const today = new Date()
      return day.getFullYear() === today.getFullYear() &&
        day.getMonth() === today.getMonth() &&
        day.getDate() === today.getDate()
    },
    isSelected(day) {
      return day.getFullYear() === this.currentDate.getFullYear() &&
        day.getMonth() === this.currentDate.getMonth() &&
        day.getDate() === this.currentDate.getDate()
    }
  }
}
</script>

<style scoped>
.mini-calendar {
  user-select: none;
  position: relative;
}

.mini-cal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 0 12px;
}

.mini-cal-title {
  font-size: 14px;
  font-weight: 600;
  color: #1e293b;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: all 0.15s ease;
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

.mini-cal-title:hover {
  background: #f1f5f9;
}

.title-arrow {
  font-size: 10px;
  transition: transform 0.2s ease;
}

.title-arrow.is-open {
  transform: rotate(180deg);
}

.year-month-picker {
  position: absolute;
  top: 40px;
  left: 0;
  right: 0;
  background: #ffffff;
  border-radius: 10px;
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
  z-index: 100;
  display: flex;
  padding: 8px 0;
  max-height: 260px;
  overflow-y: auto;
}

.ym-col-years,
.ym-col-months {
  flex: 1;
  padding: 0 4px;
}

.ym-col-years {
  border-right: 1px solid #f1f5f9;
}

.ym-item {
  padding: 8px 12px;
  font-size: 13px;
  color: #334155;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s ease;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.ym-item:hover {
  background: #f1f5f9;
}

.ym-item.is-active {
  background: #f1f5f9;
  font-weight: 600;
  color: #1e293b;
}

.ym-item.is-selected {
  color: #3b82f6;
  font-weight: 600;
}

.ym-arrow {
  font-size: 10px;
  color: #94a3b8;
}

.ym-check {
  font-size: 12px;
  color: #3b82f6;
  font-weight: 700;
}

.mini-cal-nav {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 6px;
  background: transparent;
  color: #64748b;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  transition: all 0.15s ease;
}

.mini-cal-nav:hover {
  background: #f1f5f9;
  color: #334155;
}

.mini-cal-weekdays {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  margin-bottom: 4px;
}

.mini-cal-weekday {
  text-align: center;
  font-size: 11px;
  font-weight: 600;
  color: #94a3b8;
  padding: 4px 0;
}

.mini-cal-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 2px;
}

.mini-cal-day {
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  color: #334155;
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.15s ease;
  margin: 0 auto;
}

.mini-cal-day:hover {
  background: #f1f5f9;
}

.mini-cal-day.other-month {
  color: #cbd5e1;
}

.mini-cal-day.is-today {
  background: #f1f5f9;
  color: #475569;
  font-weight: 600;
}

.mini-cal-day.is-selected {
  background: #3b82f6;
  color: #ffffff;
  font-weight: 600;
}

.mini-cal-day.is-selected.is-today {
  background: #3b82f6;
  color: #ffffff;
}
</style>
