package display

import (
	"fmt"
	"strings"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
)

func PrintWeather(data map[string]interface{}) {
	temp := data["temp_c"].(float64)
	condition := data["condition"].(string)
	city := data["city"].(string)
	humidity := data["humidity"].(int)
	wind := data["wind_kph"].(float64)

	// Выбор цвета в зависимости от температуры
	tempColor := ColorBlue
	if temp > 25 {
		tempColor = ColorRed
	} else if temp > 15 {
		tempColor = ColorYellow
	} else if temp > 0 {
		tempColor = ColorGreen
	}

	fmt.Println(strings.Repeat("─", 40))
	fmt.Printf("%s📍 %s%s\n", ColorCyan, city, ColorReset)
	fmt.Printf("%s🌡️  Температура: %s%.1f°C%s\n",
		ColorWhite, tempColor, temp, ColorReset)
	fmt.Printf("%s☁️  Состояние: %s%s%s\n",
		ColorWhite, ColorPurple, condition, ColorReset)
	fmt.Printf("%s💧 Влажность: %s%d%%%s\n",
		ColorWhite, ColorBlue, humidity, ColorReset)
	fmt.Printf("%s💨 Ветер: %s%.1f км/ч%s\n",
		ColorWhite, ColorCyan, wind, ColorReset)
	fmt.Println(strings.Repeat("─", 40))
}
