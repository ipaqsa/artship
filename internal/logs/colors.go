package logs

import "fmt"

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorGray   = "\033[37m"

	// Bold colors
	colorBoldRed    = "\033[1;31m"
	colorBoldGreen  = "\033[1;32m"
	colorBoldYellow = "\033[1;33m"
	colorBoldBlue   = "\033[1;34m"
)

// colorize wraps text with ANSI color codes
func colorize(color, text string) string {
	return color + text + colorReset
}

// Red returns text in red color
func Red(text string) string {
	return colorize(colorRed, text)
}

// Green returns text in green color
func Green(text string) string {
	return colorize(colorGreen, text)
}

// Yellow returns text in yellow color
func Yellow(text string) string {
	return colorize(colorYellow, text)
}

// Blue returns text in blue color
func Blue(text string) string {
	return colorize(colorBlue, text)
}

// Gray returns text in gray color
func Gray(text string) string {
	return colorize(colorGray, text)
}

// BoldRed returns text in bold red color
func BoldRed(text string) string {
	return colorize(colorBoldRed, text)
}

// BoldGreen returns text in bold green color
func BoldGreen(text string) string {
	return colorize(colorBoldGreen, text)
}

// BoldYellow returns text in bold yellow color
func BoldYellow(text string) string {
	return colorize(colorBoldYellow, text)
}

// BoldBlue returns text in bold blue color
func BoldBlue(text string) string {
	return colorize(colorBoldBlue, text)
}

// DiffSymbol returns colored symbol for diff status
func DiffSymbol(status string) string {
	switch status {
	case "added":
		return Green("+ ")
	case "removed":
		return Red("- ")
	case "modified":
		return Yellow("~ ")
	case "unchanged":
		return Gray("  ")
	default:
		return "  "
	}
}

// FormatDiffLine formats a diff line with color
func FormatDiffLine(status, path, details string) string {
	symbol := DiffSymbol(status)
	var coloredPath string

	switch status {
	case "added":
		coloredPath = Green(path)
	case "removed":
		coloredPath = Red(path)
	case "modified":
		coloredPath = Yellow(path)
	case "unchanged":
		coloredPath = Gray(path)
	default:
		coloredPath = path
	}

	if details != "" {
		return fmt.Sprintf("%s%s %s", symbol, coloredPath, Gray(details))
	}

	return fmt.Sprintf("%s%s", symbol, coloredPath)
}
