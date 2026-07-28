package nextdate

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const DateLayout = "20060102"

type monthRule struct {
	days        [32]bool
	months      [13]bool
	lastDay     bool
	previousDay bool
}

func NextDate(now time.Time, dstart string, repeat string) (string, error) {
	start, err := time.Parse(DateLayout, dstart)
	if err != nil {
		return "", fmt.Errorf("incorrect date: %w", err)
	}
	now = time.Date(
		now.Year(),
		now.Month(),
		now.Day(),
		0,
		0,
		0,
		0,
		time.UTC,
	)
	parts := strings.Fields(repeat)
	if len(parts) == 0 {
		return "", fmt.Errorf("incorrect repeat rule")
	}
	switch parts[0] {
	case "d":
		return nextByDays(now, start, parts)
	case "y":
		return nextByYear(now, start, parts)
	case "w":
		return nextByWeekdays(now, start, parts)
	case "m":
		return nextByMonthDays(now, start, parts)
	default:
		return "", fmt.Errorf("incorrect rule %q", parts[0])
	}
}

func nextByDays(now time.Time, start time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", fmt.Errorf("incorrect interval format")
	}
	days, err := strconv.Atoi(parts[1])
	if err != nil {
		return "", fmt.Errorf("incorrect format fo days")
	}
	if days < 1 || days > 400 {
		return "", fmt.Errorf("incorrect number fo days")
	}
	next := start.AddDate(0, 0, days)

	for !next.After(now) {
		next = next.AddDate(0, 0, days)
	}
	return next.Format(DateLayout), nil
}

func nextByYear(now time.Time, start time.Time, parts []string) (string, error) {
	if len(parts) != 1 {
		return "", fmt.Errorf("incorrect interval format")
	}

	next := start.AddDate(1, 0, 0)

	for !next.After(now) {
		next = next.AddDate(1, 0, 0)
	}
	return next.Format(DateLayout), nil
}

func nextByWeekdays(now time.Time, start time.Time, parts []string) (string, error) {
	if len(parts) != 2 {
		return "", fmt.Errorf("incorrect weekdays format")
	}
	var allowed [8]bool
	weekdays := strings.Split(parts[1], ",")
	for _, value := range weekdays {
		if value == "" {
			return "", fmt.Errorf("weekdays is empty")
		}
		weekday, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("weekdays %q have to be number", value)
		}

		if weekday < 1 || weekday > 7 {
			return "", fmt.Errorf("weekday have to be 1 - 7, weekday = %d", weekday)
		}
		allowed[weekday] = true
	}
	next := start.AddDate(0, 0, 1)

	for {
		if next.After(now) && allowed[isoWeekday(next)] {
			return next.Format(DateLayout), nil
		}

		next = next.AddDate(0, 0, 1)
	}
}

func isoWeekday(date time.Time) int {
	weekday := int(date.Weekday())

	if weekday == 0 {
		return 7
	}

	return weekday
}

func nextByMonthDays(now time.Time, start time.Time, parts []string) (string, error) {
	if len(parts) < 2 || len(parts) > 3 {
		return "", fmt.Errorf("rule m have to next format: m <day> [month]")
	}
	rule, err := parseMonthRule(parts)
	if err != nil {
		return "", err
	}
	next := start.AddDate(0, 0, 1)

	limitBase := now
	if start.After(limitBase) {
		limitBase = start
	}

	limit := limitBase.AddDate(400, 0, 0)

	for !next.After(limit) {
		if next.After(now) && matchesMonthRule(next, rule) {
			return next.Format(DateLayout), nil
		}

		next = next.AddDate(0, 0, 1)
	}

	return "", fmt.Errorf("правило не образует допустимой даты")
}

func parseMonthRule(parts []string) (monthRule, error) {
	var rule monthRule

	dayValues := strings.Split(parts[1], ",")
	for _, value := range dayValues {
		if value == "" {
			return monthRule{}, fmt.Errorf("empty month day")
		}

		day, err := strconv.Atoi(value)
		if err != nil {
			return monthRule{}, fmt.Errorf("month day %q not a number", value)
		}

		switch {
		case day >= 1 && day <= 31:
			rule.days[day] = true

		case day == -1:
			rule.lastDay = true

		case day == -2:
			rule.previousDay = true

		default:
			return monthRule{}, fmt.Errorf("month day have to be 1 to 31, -1 or -2")
		}
	}
	if len(parts) == 2 {
		for month := 1; month <= 12; month++ {
			rule.months[month] = true
		}
		return rule, nil
	}
	monthValues := strings.Split(parts[2], ",")

	for _, value := range monthValues {
		if value == "" {
			return monthRule{}, fmt.Errorf("empty month")
		}

		month, err := strconv.Atoi(value)
		if err != nil {
			return monthRule{}, fmt.Errorf("month %q is not a number", value)
		}

		if month < 1 || month > 12 {
			return monthRule{}, fmt.Errorf("month have to be 1 to 12")
		}

		rule.months[month] = true
	}

	return rule, nil
}

func matchesMonthRule(date time.Time, rule monthRule) bool {
	month := int(date.Month())

	if !rule.months[month] {
		return false
	}

	day := date.Day()

	if rule.days[day] {
		return true
	}

	lastDay := daysInMonth(date.Year(), date.Month())

	if rule.lastDay && day == lastDay {
		return true
	}

	if rule.previousDay && day == lastDay-1 {
		return true
	}

	return false
}

func daysInMonth(year int, month time.Month) int {
	return time.Date(
		year,
		month+1,
		0,
		0,
		0,
		0,
		0,
		time.UTC,
	).Day()
}
