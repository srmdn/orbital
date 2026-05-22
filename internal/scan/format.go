package scan

import "fmt"

// FormatSize formats a size in MB to a human-readable string.
func FormatSize(mb int64) string {
	neg := ""
	if mb < 0 {
		neg = "-"
		mb = -mb
	}
	if mb >= 1024 {
		return neg + fmt.Sprintf("%.1f GB", float64(mb)/1024.0)
	}
	if mb == 0 {
		return "0 MB"
	}
	return neg + fmt.Sprintf("%d MB", mb)
}

// FormatCleanHint returns the hint string as-is. Empty hints return empty string.
// Real formatting happens in the display layer.
func FormatCleanHint(hint string) string {
	return hint
}
