<template>
  <div class="month-view">
    <div class="month-header">
      <div v-for="name in dayNames" :key="name" class="weekday-header">{{ name }}</div>
    </div>

    <div class="month-grid">
      <div
        v-for="(day, index) in calendarDays"
        :key="index"
        class="month-cell"
        :class="{
          'other-month': !isCurrentMonth(day),
          'is-today': isToday(day)
        }"
      >
        <div class="cell-date" :class="{ 'today-date': isToday(day) }">{{ day.getDate() }}</div>
        <div class="cell-events">
          <div
            v-for="event in getEventsForDay(day).slice(0, 3)"
            :key="event.id"
            class="month-event-item"
            :style="{ backgroundColor: getEventColor(event) + '18', color: getEventColor(event) }"
            @click.stop="$emit('event-click', event)"
            @dblclick.stop="$emit('event-dblclick', event)"
          >
            <span class="event-dot" :style="{ backgroundColor: getEventColor(event) }"></span>
            <span class="event-text">{{ event.title }}</span>
          </div>
          <div v-if="getEventsForDay(day).length > 3" class="more-events">
            +{{ getEventsForDay(day).length - 3 }} 更多
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MonthView',
  props: {
    currentDate: { type: Date, required: true },
    events: { type: Array, default: () => [] }
  },
  data() {
    return {
      dayNames: ['周一', '周二', '周三', '周四', '周五', '周六', '周日']
    }
  },
  computed: {
    calendarDays() {
      const year = this.currentDate.getFullYear()
      const month = this.currentDate.getMonth()
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
  methods: {
    isCurrentMonth(day) {
      return day.getMonth() === this.currentDate.getMonth()
    },
    isToday(day) {
      const today = new Date()
      return day.getFullYear() === today.getFullYear() &&
        day.getMonth() === today.getMonth() &&
        day.getDate() === today.getDate()
    },
    getEventsForDay(day) {
      return this.events.filter(event => {
        const start = new Date(event.start_time)
        const end = new Date(event.end_time)
        const dayStart = new Date(day)
        dayStart.setHours(0, 0, 0, 0)
        const dayEnd = new Date(day)
        dayEnd.setHours(23, 59, 59, 999)
        return start <= dayEnd && end >= dayStart
      })
    },
    getEventColor(event) {
      const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
      return colors[(event.id || 0) % colors.length]
    }
  }
}
</script>

<style scoped>
.month-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: #ffffff;
}

.month-header {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  border-bottom: 1px solid #f1f5f9;
  flex-shrink: 0;
}

.weekday-header {
  text-align: center;
  padding: 12px 0;
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  letter-spacing: 0.02em;
}

.month-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  grid-template-rows: repeat(6, 1fr);
  overflow: hidden;
}

.month-cell {
  border-right: 1px solid #f1f5f9;
  border-bottom: 1px solid #f1f5f9;
  padding: 6px 8px;
  overflow: hidden;
  cursor: default;
  min-height: 0;
  transition: background 0.15s ease;
}

.month-cell:nth-child(7n) {
  border-right: none;
}

.month-cell:hover {
  background: #fafbfc;
}

.month-cell.other-month {
  background: #fafbfc;
}

.month-cell.other-month .cell-date {
  color: #cbd5e1;
}

.month-cell.is-today {
  background: #eff6ff;
}

.cell-date {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
  margin-bottom: 4px;
  letter-spacing: -0.01em;
}

.today-date {
  color: #3b82f6;
  font-size: 14px;
}

.cell-events {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.month-event-item {
  display: flex;
  align-items: center;
  gap: 5px;
  padding: 2px 6px;
  border-radius: 6px;
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  line-height: 20px;
  transition: opacity 0.15s ease;
}

.month-event-item:hover {
  opacity: 0.8;
}

.event-dot {
  width: 5px;
  height: 5px;
  border-radius: 50%;
  flex-shrink: 0;
}

.event-text {
  overflow: hidden;
  text-overflow: ellipsis;
}

.more-events {
  font-size: 11px;
  color: #94a3b8;
  padding: 2px 6px;
  cursor: pointer;
  font-weight: 500;
}

.more-events:hover {
  color: #64748b;
}
</style>
