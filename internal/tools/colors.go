package tools

import "fmt"

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorGray   = "\033[37m"
	ColorWhite  = "\033[97m"

	// Bold colors
	ColorBoldRed    = "\033[1;31m"
	ColorBoldGreen  = "\033[1;32m"
	ColorBoldYellow = "\033[1;33m"
	ColorBoldBlue   = "\033[1;34m"
)

// Colorize wraps text with ANSI color codes
func Colorize(color, text string) string {
	return color + text + ColorReset
}

// Red returns text in red color
func Red(text string) string {
	return Colorize(ColorRed, text)
}

// Green returns text in green color
func Green(text string) string {
	return Colorize(ColorGreen, text)
}

// Yellow returns text in yellow color
func Yellow(text string) string {
	return Colorize(ColorYellow, text)
}

// Blue returns text in blue color
func Blue(text string) string {
	return Colorize(ColorBlue, text)
}

// Gray returns text in gray color
func Gray(text string) string {
	return Colorize(ColorGray, text)
}

// BoldRed returns text in bold red color
func BoldRed(text string) string {
	return Colorize(ColorBoldRed, text)
}

// BoldGreen returns text in bold green color
func BoldGreen(text string) string {
	return Colorize(ColorBoldGreen, text)
}

// BoldYellow returns text in bold yellow color
func BoldYellow(text string) string {
	return Colorize(ColorBoldYellow, text)
}

// BoldBlue returns text in bold blue color
func BoldBlue(text string) string {
	return Colorize(ColorBoldBlue, text)
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
