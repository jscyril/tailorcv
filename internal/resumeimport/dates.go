package resumeimport

import (
	"fmt"
	"regexp"
	"strings"
)

var months = map[string]int{
	"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6, "jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12,
	// Seasons map to their usual first month.
	"spring": 3, "summer": 6, "fall": 9, "autumn": 9, "winter": 12,
}

var (
	monthYear   = regexp.MustCompile(`(?i)^([a-z]+)\.?,?\s+(\d{4})$`)
	numericDate = regexp.MustCompile(`^(\d{1,2})[/.-](\d{4})$|^(\d{4})[/.-](\d{1,2})$`)
	yearOnly    = regexp.MustCompile(`^(\d{4})$`)
	rangeSplit  = regexp.MustCompile(`\s*(?:–|—|-|\bto\b)\s*`)
	currentWord = regexp.MustCompile(`(?i)^(present|current|now|ongoing|today)$`)
	datePrefix  = regexp.MustCompile(`(?i)^(expected|anticipated|graduat(?:ed|ing|ion))[:\s]+`)
)

// parseMonth turns "June 2020", "Sept. 2019", "06/2020", "2020-06", or
// "2020" into YYYY-MM. A bare year becomes January for a start date and
// December for an end date.
func parseMonth(value string, end bool) (string, bool) {
	value = strings.TrimSpace(value)
	if match := monthYear.FindStringSubmatch(value); match != nil {
		name := strings.ToLower(match[1])
		if len(name) >= 3 {
			if month, ok := months[name[:3]]; ok {
				return fmt.Sprintf("%s-%02d", match[2], month), true
			}
		}
		if month, ok := months[name]; ok {
			return fmt.Sprintf("%s-%02d", match[2], month), true
		}
		return "", false
	}
	if match := numericDate.FindStringSubmatch(value); match != nil {
		month, year := match[1], match[2]
		if match[3] != "" {
			year, month = match[3], match[4]
		}
		var number int
		fmt.Sscanf(month, "%d", &number)
		if number < 1 || number > 12 {
			return "", false
		}
		return fmt.Sprintf("%s-%02d", year, number), true
	}
	if match := yearOnly.FindStringSubmatch(value); match != nil {
		if end {
			return match[1] + "-12", true
		}
		return match[1] + "-01", true
	}
	return "", false
}

// parseRange reads "June 2020 -- Present" style ranges. A single date is
// treated as both start and end (a one-off project or graduation).
func parseRange(value string) (start, end string, current, ok bool) {
	parts := rangeSplit.Split(datePrefix.ReplaceAllString(strings.TrimSpace(value), ""), 2)
	start, ok = parseMonth(parts[0], false)
	if !ok {
		return "", "", false, false
	}
	if len(parts) == 1 {
		return start, start, false, true
	}
	if currentWord.MatchString(strings.TrimSpace(parts[1])) {
		return start, "", true, true
	}
	end, ok = parseMonth(parts[1], true)
	return start, end, false, ok
}

// looksLikeDate reports whether a heading part is a date range, used to tell
// which of Jake's four heading arguments holds the dates.
func looksLikeDate(value string) bool {
	_, _, _, ok := parseRange(value)
	return ok
}
