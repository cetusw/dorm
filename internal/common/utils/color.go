package utils

import "google.golang.org/api/sheets/v4"

func RGBToInt(r, g, b uint8) int {
	return (int(r) << 16) | (int(g) << 8) | int(b)
}

func IntToSheetsColor(color int) *sheets.Color {
	// 1. Извлекаем 8-битные компоненты (0-255) с помощью битовых операций.
	r := (color >> 16) & 0xFF
	g := (color >> 8) & 0xFF
	b := color & 0xFF

	// 2. Создаем и возвращаем готовый объект для Google API.
	//    Каждый компонент преобразуется в float64 и нормализуется в диапазон 0.0-1.0.
	return &sheets.Color{
		Red:   float64(r) / 255.0,
		Green: float64(g) / 255.0,
		Blue:  float64(b) / 255.0,
		Alpha: 1.0, // По умолчанию делаем цвет полностью непрозрачным
	}
}

func IntToRGBFloats(color int) (r, g, b float64) {
	rInt := (color >> 16) & 0xFF
	gInt := (color >> 8) & 0xFF
	bInt := color & 0xFF

	r = float64(rInt) / 255.0
	g = float64(gInt) / 255.0
	b = float64(bInt) / 255.0

	return
}
