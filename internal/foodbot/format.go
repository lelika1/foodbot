package foodbot

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

type dayResult struct {
	Date    string
	Kcal    uint32
	InLimit bool
}

// 0 element - Today, 1 - yestrday, ...
func formatStat7(week []dayResult) string {
	var sb strings.Builder
	for d, r := range week {
		if d == 0 {
			fmt.Fprintf(&sb, "`%s Today:         ` *%v kcal*\n", color(r.InLimit), r.Kcal)
			continue
		}
		fmt.Fprintf(&sb, "`%s %v:` *%v kcal*\n", color(r.InLimit), r.Date, r.Kcal)

	}
	return sb.String()
}

func formatStat(reports []Report, limit uint32) string {
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
			Begin: fmt.Sprintf("%s: %v", r.When.Format("15:04:05"), r.Product),
			End:   fmt.Sprintf("%v kcal", kcal),
		})

		if l := utf8.RuneCountInString(lines[i].Begin) + utf8.RuneCountInString(lines[i].End); maxLen < l {
			maxLen = l
		}
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "You ate today *%v kcal*:\n", total)
	for _, line := range lines {
		spaces := maxLen - (utf8.RuneCountInString(line.Begin) + utf8.RuneCountInString(line.End)) + 1
		fmt.Fprintf(&sb, "`%s%s`*%s*\n", line.Begin, strings.Repeat(" ", spaces), line.End)
	}

	if total < limit {
		fmt.Fprintf(&sb, "\n✅ Still *%v kcal* left\\.\n", limit-total)
	} else {
		fmt.Fprintf(&sb, "\n❌ You ate *%v kcal* over the limit\\.\n", total-limit)
	}

	return sb.String()
}

func color(inLimit bool) string {
	if inLimit {
		return "✅"
	}
	return "❌"
}
