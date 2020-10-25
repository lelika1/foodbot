package foodbot

import (
	"fmt"
	"strings"
)

type weeklyReport struct {
	Today        uint32
	TodayInLimit bool
	History      []shortDayReport
}

type shortDayReport struct {
	Date    string
	Kcal    uint32
	InLimit bool
}

func formatWeeklyReport(report weeklyReport) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "`%s Today:         ` *%v kcal*\n", color(report.TodayInLimit), report.Today)
	for _, r := range report.History {
		fmt.Fprintf(&sb, "`%s %v:` *%v kcal*\n", color(r.InLimit), r.Date, r.Kcal)
	}
	return sb.String()
}

func formatDayReport(reports []Report) string {
	if len(reports) == 0 {
		return "*You ate nothing so far\\.*"
	}

	type Line struct {
		Begin, End string
	}

	var lines []Line

	var total uint32
	var maxLen int
	for i, r := range reports {
		kcal := r.Kcal * r.Grams / 100
		total += kcal

		lines = append(lines, Line{
			Begin: fmt.Sprintf("%s: %v", r.Time.Format("15:04:05"), r.Product),
			End:   fmt.Sprintf("%v kcal", kcal),
		})

		if l := len(lines[i].Begin) + len(lines[i].End); maxLen < l {
			maxLen = l
		}
	}

	var sb strings.Builder
	sb.WriteString("*You ate today:*\n")
	for _, line := range lines {
		spaces := maxLen - (len(line.Begin) + len(line.End)) + 1
		fmt.Fprintf(&sb, "`%s%s`*%s*\n", line.Begin, strings.Repeat(" ", spaces), line.End)
	}
	fmt.Fprintf(&sb, "\n`Total:` *%v kcal*\n", total)
	return sb.String()
}

func color(inLimit bool) string {
	if inLimit {
		return "✅"
	}
	return "❌"
}