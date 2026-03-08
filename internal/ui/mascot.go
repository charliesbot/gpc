package ui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color palette indices.
const (
	t = 0 // transparent (terminal background)
	k = 1 // black outline
	g = 2 // android green
	w = 3 // white (eye highlight)
	m = 4 // mint/light green (mouth area)
	d = 5 // dark green (shading)
)

var palette = map[int]string{
	k: "#222233",
	g: "#3DDC84",
	w: "#FFFFFF",
	m: "#7BEFB2",
	d: "#2DA96B",
}

// Bugdroid pixel art — 24 wide x 14 tall.
// Dome-shaped head with angled antennas, inspired by the Android mascot.
// Each cell is one pixel; pairs of rows become one terminal line via half-blocks.
var pixels = [][]int{
	{t, t, t, g, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, g, t, t, t}, // antennas tip
	{t, t, t, t, g, t, t, t, t, t, t, t, t, t, t, t, t, t, t, g, t, t, t, t}, // antennas mid
	{t, t, t, t, t, g, t, t, t, t, t, t, t, t, t, t, t, t, g, t, t, t, t, t}, // antennas base
	{t, t, t, t, t, k, g, g, g, g, g, g, g, g, g, g, g, g, k, t, t, t, t, t}, // dome top
	{t, t, t, t, k, g, g, g, g, g, g, g, g, g, g, g, g, g, g, k, t, t, t, t}, // dome wider
	{t, t, t, k, g, g, g, g, g, g, g, g, g, g, g, g, g, g, g, g, k, t, t, t}, // dome wider
	{t, t, k, g, g, g, g, g, g, g, g, g, g, g, g, g, g, g, g, g, g, k, t, t}, // dome wider
	{t, k, g, g, g, g, k, k, k, g, g, g, g, g, g, k, k, k, g, g, g, g, k, t}, // eyes top
	{t, k, g, g, g, g, k, w, k, g, g, g, g, g, g, k, w, k, g, g, g, g, k, t}, // eyes highlight
	{t, k, g, g, g, g, k, k, k, g, g, g, g, g, g, k, k, k, g, g, g, g, k, t}, // eyes bottom
	{t, k, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, k, t}, // mouth
	{t, k, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, m, k, t}, // mouth
	{t, t, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, k, t, t}, // bottom
	{t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t, t}, // padding
}

// renderCell uses the half-block technique to pack two vertical pixels into one
// terminal character cell. The top pixel uses foreground color and the bottom
// pixel uses background color on ▀ (upper half block).
func renderCell(top, bot int) string {
	if top == t && bot == t {
		return " "
	}
	if top == t {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette[bot])).
			Render("▄")
	}
	if bot == t {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette[top])).
			Render("▀")
	}
	if top == bot {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color(palette[top])).
			Render("█")
	}
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(palette[top])).
		Background(lipgloss.Color(palette[bot])).
		Render("▀")
}

// Mascot returns the Bugdroid pixel art as a styled string.
func Mascot() string {
	var lines []string
	for row := 0; row < len(pixels)-1; row += 2 {
		var line strings.Builder
		for col := 0; col < len(pixels[row]); col++ {
			line.WriteString(renderCell(pixels[row][col], pixels[row+1][col]))
		}
		lines = append(lines, line.String())
	}
	return strings.Join(lines, "\n")
}

// Banner returns the mascot with the app name and version styled below it.
func Banner(version string) string {
	mascot := Mascot()

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#3DDC84")).
		Render("gpc")

	ver := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#666666")).
		Render(version)

	subtitle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#888888")).
		Render("Google Play Console CLI")

	text := lipgloss.JoinVertical(lipgloss.Center,
		mascot,
		"",
		title+" "+ver,
		subtitle,
	)

	return lipgloss.NewStyle().
		Padding(1, 2).
		Render(text)
}
