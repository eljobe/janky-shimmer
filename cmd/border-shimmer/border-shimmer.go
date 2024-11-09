// Copyright (c) Pepper Lebeck-Jobe, 2024
// This file is licensed under the MIT License.
// See the LICENSE.md file in the project root.

// border-shimmer pulses border colors in configurable ways.
//
// Users must have the `borders` command installed.
// See: https://github.com/FelixKratz/JankyBorders
package main

import (
	"flag"
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Colors         []string `toml:"colors"`
	InactiveColors []string `toml:"inactive_colors"`
	Secs           float64  `toml:"secs"`
	FPS            float64  `toml:"fps"`
	Width          float64  `toml:"width"`
}

func defaultConfig() Config {
	return Config{
		Colors: []string{
			"#FF0000FF", // Red
			"#FFA500FF", // Orange
			"#FFFF00FF", // Yellow
			"#008000FF", // Green
			"#0000FFFF", // Blue
			"#4B0082FF", // Indigo
		},
		InactiveColors: []string{},
		Secs:           3.0,
		FPS:            3.0,
		Width:          5,
	}
}

func getConfigFilePath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Printf("Error getting user home directory: %v\n", err)
			os.Exit(1)
		}
		configDir = filepath.Join(homeDir, ".config")
	}
	return filepath.Join(configDir, "border-shimmer", "config.toml")
}

func loadConfigFile() (Config, error) {
	config := Config{}
	configFilePath := getConfigFilePath()

	data, err := os.ReadFile(configFilePath)
	if err == nil {
		err = toml.Unmarshal(data, &config)
		if err != nil {
			return config, fmt.Errorf("error parsing config file %s: %v", configFilePath, err)
		}
		return config, nil
	}

	// No config file found, return empty config
	return config, nil
}

func parseColors(colors []string) ([]color.RGBA, error) {
	var colorList []color.RGBA
	for _, c := range colors {
		// Remove "#" prefix if present
		c = strings.TrimPrefix(c, "#")

		if len(c) != 8 {
			return nil, fmt.Errorf("invalid color format: %s", c)
		}

		// Parse as uint32
		var val uint32
		_, err := fmt.Sscanf(c, "%08X", &val)
		if err != nil {
			return nil, fmt.Errorf("invalid color value: %s", c)
		}

		// Extract ARGB components
		r := uint8(val >> 24)
		g := uint8((val >> 16) & 0xFF)
		b := uint8((val >> 8) & 0xFF)
		a := uint8(val & 0xFF)

		colorList = append(colorList, color.RGBA{R: r, G: g, B: b, A: a})
	}
	return colorList, nil
}

func interpolateColor(c1, c2 color.RGBA, t float64) color.RGBA {
	r := uint8(float64(c1.R)*(1-t) + float64(c2.R)*t)
	g := uint8(float64(c1.G)*(1-t) + float64(c2.G)*t)
	b := uint8(float64(c1.B)*(1-t) + float64(c2.B)*t)
	a := uint8(float64(c1.A)*(1-t) + float64(c2.A)*t)
	return color.RGBA{r, g, b, a}
}

func colorToHex(c color.RGBA) string {
	// Format: 0xAARRGGBB
	return fmt.Sprintf("0x%02X%02X%02X%02X", c.A, c.R, c.G, c.B)
}

func main() {
	// Default configuration
	cfg := defaultConfig()

	// Parse command-line flags
	colorsFlag := flag.String("colors", "", "Comma-separated list of colors in #RRGGBBAA format")
	inactiveColorsFlag := flag.String("inactive_colors", "", "Comma-separated list of inactive colors in #RRGGBBAA format")
	secsFlag := flag.Float64("secs", 0, "Number of seconds between each color")
	fpsFlag := flag.Float64("fps", 0, "Frames per second (number of intervening colors per second)")
	widthFlag := flag.Float64("width", 0, "Width of the border")
	flag.Parse()

	// Load configuration file
	fileConfig, err := loadConfigFile()
	if err != nil {
		fmt.Printf("Error loading config file: %v\n", err)
		os.Exit(1)
	}

	// Override default config with config file settings
	if len(fileConfig.Colors) > 0 {
		cfg.Colors = fileConfig.Colors
	}
	if len(fileConfig.InactiveColors) > 0 {
		cfg.InactiveColors = fileConfig.InactiveColors
	}
	if fileConfig.Secs > 0 {
		cfg.Secs = fileConfig.Secs
	}
	if fileConfig.FPS > 0 {
		cfg.FPS = fileConfig.FPS
	}
	if fileConfig.Width > 0 {
		cfg.Width = fileConfig.Width
	}

	// Override config with command-line flags
	if *colorsFlag != "" {
		cfg.Colors = strings.Split(*colorsFlag, ",")
	}
	if *inactiveColorsFlag != "" {
		cfg.InactiveColors = strings.Split(*inactiveColorsFlag, ",")
	}
	if *secsFlag > 0 {
		cfg.Secs = *secsFlag
	}
	if *fpsFlag > 0 {
		cfg.FPS = *fpsFlag
	}
	if *widthFlag > 0 {
		cfg.Width = *widthFlag
	}

	// Parse colors
	colors, err := parseColors(cfg.Colors)
	if err != nil {
		fmt.Printf("Error parsing colors: %v\n", err)
		os.Exit(1)
	}

	var inactiveColors []color.RGBA
	if len(cfg.InactiveColors) > 0 {
		inactiveColors, err = parseColors(cfg.InactiveColors)
		if err != nil {
			fmt.Printf("Error parsing inactive colors: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Use active colors with offset
		offset := len(colors) / 2
		inactiveColors = append(colors[offset:], colors[:offset]...)
	}

	if len(inactiveColors) != len(colors) {
		fmt.Printf("Error: The number of inactive colors must match the number of active colors.\n")
		os.Exit(1)
	}

	stepsPerTransition := int(cfg.Secs * cfg.FPS)
	delay := time.Duration(1.0 / cfg.FPS * float64(time.Second))

	for {
		for i := 0; i < len(colors); i++ {
			currentActiveColor := colors[i]
			nextActiveColor := colors[(i+1)%len(colors)] // Loop back to the first color

			currentInactiveColor := inactiveColors[i]
			nextInactiveColor := inactiveColors[(i+1)%len(inactiveColors)]

			for step := 0; step <= stepsPerTransition; step++ {
				t := float64(step) / float64(stepsPerTransition)
				interpolatedActiveColor := interpolateColor(currentActiveColor, nextActiveColor, t)
				activeColorHex := colorToHex(interpolatedActiveColor)

				interpolatedInactiveColor := interpolateColor(currentInactiveColor, nextInactiveColor, t)
				inactiveColorHex := colorToHex(interpolatedInactiveColor)

				binary := "borders"
				argMap := map[string]string{
					"active_color":   "glow(" + activeColorHex + ")",
					"inactive_color": inactiveColorHex,
					"width":          fmt.Sprintf("%f", cfg.Width),
				}
				var args []string
				for k, v := range argMap {
					args = append(args, fmt.Sprintf("%s=%s", k, v))
				}
				cmd := exec.Command(binary, args...)
				err := cmd.Run()
				if err != nil {
					fmt.Printf("Error running borders command: %v\n", err)
				}

				time.Sleep(delay)
			}
		}
	}
}
