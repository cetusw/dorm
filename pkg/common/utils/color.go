package utils

import (
	"encoding/hex"
	"fmt"
	"strings"

	"google.golang.org/api/sheets/v4"
)

func HexToSheetsColor(hexString string) (*sheets.Color, error) {
	cleanHex := strings.TrimPrefix(hexString, "#")

	if len(cleanHex) != 6 {
		return nil, fmt.Errorf("invalid hex. must contain 6 chars, got %d", len(cleanHex))
	}

	rgb, err := hex.DecodeString(cleanHex)
	if err != nil {
		return nil, fmt.Errorf("invalid hex: %w", err)
	}

	color := &sheets.Color{
		Red:   float64(rgb[0]) / 255.0,
		Green: float64(rgb[1]) / 255.0,
		Blue:  float64(rgb[2]) / 255.0,
		Alpha: 1.0,
	}

	return color, nil
}

func DarkenColorRGB(color *sheets.Color, factor float64) *sheets.Color {
	if color == nil {
		return nil
	}
	if factor < 0 {
		factor = 0
	}
	if factor > 1 {
		factor = 1
	}
	return &sheets.Color{
		Red:   color.Red * factor,
		Green: color.Green * factor,
		Blue:  color.Blue * factor,
		Alpha: color.Alpha,
	}
}
