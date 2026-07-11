package services

import (
	"encoding/json"
	"time"
)

// CalendarRepeatRule 重复规则JSON结构
type CalendarRepeatRule struct {
	Type        string  `json:"type"`                  // daily | weekly | monthly_week | monthly_day | yearly | workday | custom | none
	Interval    int     `json:"interval,omitempty"`     // 间隔数量
	Unit        string  `json:"unit,omitempty"`         // days | weeks | months | years (仅custom用)
	Weekday     int     `json:"weekday,omitempty"`      // 0-6 (周日-周六)，仅weekly/monthly_week用
	MonthDay    int     `json:"monthDay,omitempty"`     // 1-31，仅monthly_day/yearly用
	WeekOfMonth int     `json:"weekOfMonth,omitempty"`  // 1-5，仅monthly_week用（第几个星期X）
	MonthOfYear int     `json:"monthOfYear,omitempty"`  // 1-12，仅yearly用
	EndDate     string  `json:"endDate,omitempty"`      // 可选，结束日期 YYYY-MM-DD
	Occurrences *int    `json:"occurrences,omitempty"`  // 可选，重复次数
}

// ParseRepeatRule 解析重复规则JSON
func ParseRepeatRule(ruleJSON string) (*CalendarRepeatRule, error) {
	if ruleJSON == "" || ruleJSON == "null" {
		return nil, nil
	}
	var rule CalendarRepeatRule
	err := json.Unmarshal([]byte(ruleJSON), &rule)
	if err != nil {
		return nil, err
	}
	if rule.Type == "" || rule.Type == "none" {
		return nil, nil
	}
	return &rule, nil
}

// ExpandRecurringEvents 将重复规则展开为日期列表
// startDate: 日程原始开始时间
// rangeStart, rangeEnd: 查询范围
// 返回该范围内所有匹配的实例开始时间
func ExpandRecurringEvents(rule *CalendarRepeatRule, startDate, rangeStart, rangeEnd time.Time) []time.Time {
	if rule == nil {
		// 非重复日程，检查是否在范围内
		if !startDate.Before(rangeStart) && !startDate.After(rangeEnd) {
			return []time.Time{startDate}
		}
		return nil
	}

	var instances []time.Time
	var endDate time.Time
	if rule.EndDate != "" {
		endDate, _ = time.Parse("2006-01-02", rule.EndDate)
	}

	interval := rule.Interval
	if interval < 1 {
		interval = 1
	}

	occurrenceCount := 0
	maxOccurrences := 1000 // 安全上限
	if rule.Occurrences != nil && *rule.Occurrences > 0 {
		maxOccurrences = *rule.Occurrences
	}

	current := startDate

	switch rule.Type {
	case "daily":
		for !current.After(rangeEnd) {
			if !current.Before(rangeStart) {
				instances = append(instances, current)
			}
			occurrenceCount++
			if occurrenceCount >= maxOccurrences {
				break
			}
			if !endDate.IsZero() && current.After(endDate) {
				break
			}
			current = current.AddDate(0, 0, interval)
		}

	case "weekly":
		for !current.After(rangeEnd) {
			if !current.Before(rangeStart) && int(current.Weekday()) == rule.Weekday {
				instances = append(instances, current)
			}
			occurrenceCount++
			if occurrenceCount >= maxOccurrences {
				break
			}
			if !endDate.IsZero() && current.After(endDate) {
				break
			}
			current = current.AddDate(0, 0, interval*7)
		}

	case "monthly_week":
		// 每月第N个星期X
		// 从startDate月份开始逐月检查
		year, monthVal, _ := startDate.Date()
		monthNum := int(monthVal)
		for {
			candidate := getNthWeekdayOfMonth(year, time.Month(monthNum), rule.WeekOfMonth, rule.Weekday, startDate.Hour(), startDate.Minute(), startDate.Second(), startDate.Location())
			if candidate.After(rangeEnd) {
				break
			}
			if !candidate.Before(rangeStart) && !candidate.Before(startDate) {
				instances = append(instances, candidate)
			}
			occurrenceCount++
			if occurrenceCount >= maxOccurrences {
				break
			}
			if !endDate.IsZero() && candidate.After(endDate) {
				break
			}
			// 增加interval个月
			monthNum += interval
			for monthNum > 12 {
				monthNum -= 12
				year++
			}
		}

	case "monthly_day":
		// 每月X日
		year, monthVal, _ := startDate.Date()
		monthNum := int(monthVal)
		for {
			candidate := time.Date(year, time.Month(monthNum), rule.MonthDay, startDate.Hour(), startDate.Minute(), startDate.Second(), 0, startDate.Location())
			if candidate.After(rangeEnd) {
				break
			}
			if !candidate.Before(rangeStart) && !candidate.Before(startDate) {
				// 该月是否有这一天（如2月30日不存在）
				if candidate.Day() == rule.MonthDay {
					instances = append(instances, candidate)
				}
			}
			occurrenceCount++
			if occurrenceCount >= maxOccurrences {
				break
			}
			if !endDate.IsZero() && candidate.After(endDate) {
				break
			}
			monthNum += interval
			for monthNum > 12 {
				monthNum -= 12
				year++
			}
		}

	case "yearly":
		// 每年X月X日
		year := startDate.Year()
		for {
			candidate := time.Date(year, time.Month(rule.MonthOfYear), rule.MonthDay, startDate.Hour(), startDate.Minute(), startDate.Second(), 0, startDate.Location())
			if candidate.After(rangeEnd) {
				break
			}
			if !candidate.Before(rangeStart) && !candidate.Before(startDate) {
				if candidate.Month() == time.Month(rule.MonthOfYear) && candidate.Day() == rule.MonthDay {
					instances = append(instances, candidate)
				}
			}
			occurrenceCount++
			if occurrenceCount >= maxOccurrences {
				break
			}
			if !endDate.IsZero() && candidate.After(endDate) {
				break
			}
			year += interval
		}

	case "workday":
		// 每个工作日（周一至周五）
		for !current.After(rangeEnd) {
			weekday := current.Weekday()
			if !current.Before(rangeStart) && weekday != time.Saturday && weekday != time.Sunday {
				instances = append(instances, current)
			}
			occurrenceCount++
			if occurrenceCount >= maxOccurrences {
				break
			}
			if !endDate.IsZero() && current.After(endDate) {
				break
			}
			current = current.AddDate(0, 0, 1)
		}

	case "custom":
		// 自定义：每N天/周/月/年
		for !current.After(rangeEnd) {
			if !current.Before(rangeStart) {
				instances = append(instances, current)
			}
			occurrenceCount++
			if occurrenceCount >= maxOccurrences {
				break
			}
			if !endDate.IsZero() && current.After(endDate) {
				break
			}
			switch rule.Unit {
			case "days":
				current = current.AddDate(0, 0, interval)
			case "weeks":
				current = current.AddDate(0, 0, interval*7)
			case "months":
				current = current.AddDate(0, interval, 0)
			case "years":
				current = current.AddDate(interval, 0, 0)
			default:
				current = current.AddDate(0, 0, interval)
			}
		}
	}

	return instances
}

// getNthWeekdayOfMonth 获取某月第N个星期X的日期
func getNthWeekdayOfMonth(year int, month time.Month, weekOfMonth, weekday, hour, minute, second int, loc *time.Location) time.Time {
	// 获取该月第一天
	firstDay := time.Date(year, month, 1, hour, minute, second, 0, loc)
	firstWeekday := int(firstDay.Weekday())

	// 计算第一个目标星期X的日期
	diff := weekday - firstWeekday
	if diff < 0 {
		diff += 7
	}
	firstTarget := firstDay.AddDate(0, 0, diff)

	// 计算第N个
	target := firstTarget.AddDate(0, 0, (weekOfMonth-1)*7)

	// 检查是否还在同一个月
	if target.Month() != month {
		// 返回月末最后一天作为标记
		return time.Date(year, month+1, 0, hour, minute, second, 0, loc)
	}

	return target
}

// CheckTimeOverlap 检查两个时间段是否重叠
func CheckTimeOverlap(start1, end1, start2, end2 time.Time) bool {
	return start1.Before(end2) && start2.Before(end1)
}
