package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"time"
)

const (
	version = "1.0.4"
)

type Config struct {
	Speed          int
	Sensitivity    float64
	CheckInterval  int
	DeviationLimit float64
	ShowVersion    bool
	Verbose        bool
	RandomStart    bool
	RunMinutes     int
}

func main() {
	config := parseFlags()

	if config.ShowVersion {
		fmt.Printf("DVD Screen Saver v%s\n", version)
		return
	}

	runDVDEffect(config)
}

func parseFlags() *Config {
	config := &Config{}

	flag.IntVar(&config.Speed, "speed", 10, "Скорость движения курсора (пикселей за шаг)")
	flag.IntVar(&config.Speed, "s", 10, "Скорость движения курсора (короткая версия)")

	flag.Float64Var(&config.Sensitivity, "sensitivity", 15.0, "Чувствительность обнаружения движения мыши")
	flag.Float64Var(&config.Sensitivity, "sens", 15.0, "Чувствительность (короткая версия)")

	flag.IntVar(&config.CheckInterval, "interval", 100, "Интервал проверки движения (миллисекунды)")
	flag.IntVar(&config.CheckInterval, "i", 100, "Интервал проверки (короткая версия)")

	flag.Float64Var(&config.DeviationLimit, "deviation", 150.0, "Лимит отклонения для обнаружения вмешательства")
	flag.Float64Var(&config.DeviationLimit, "d", 150.0, "Лимит отклонения (короткая версия)")

	flag.BoolVar(&config.ShowVersion, "version", false, "Показать версию программы")
	flag.BoolVar(&config.ShowVersion, "v", false, "Показать версию (короткая версия)")

	flag.BoolVar(&config.Verbose, "verbose", false, "Подробный вывод")

	flag.BoolVar(&config.RandomStart, "random", true, "Случайная начальная позиция")
	flag.BoolVar(&config.RandomStart, "r", true, "Случайная позиция (короткая версия)")

	flag.IntVar(&config.RunMinutes, "time", 0, "Авто-выход через указанное количество минут (0 — без таймера)")
	flag.IntVar(&config.RunMinutes, "t", 0, "Авто-выход через указанное количество минут")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `DVD Screen Saver v%s

Эффект DVD-логотипа для вашего курсора мыши.
Программа прекращает работу при обнаружении движения мыши пользователем.

ИСПОЛЬЗОВАНИЕ:
    dvd [flags]

ФЛАГИ:
`, version)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, `
ПРИМЕРЫ:
    dvd                              # Запуск с настройками по умолчанию
    dvd -s 20                        # Быстрая скорость (20 пикселей)
    dvd --speed 5 --sensitivity 10   # Медленно и более чувствительно
    dvd -v                           # Показать версию
    dvd --verbose                    # Подробный режим

ГОРЯЧИЕ КЛАВИШИ:
    Ctrl+C        Выход из программы
    Движение мыши Автоматический выход

АВТОР:
    Создано merdernoty с ❤️ для автоматизации
`)
	}

	flag.Parse()

	return config
}

func runDVDEffect(config *Config) {
	rand.Seed(time.Now().UnixNano())

	screenWidth, screenHeight := getScreenSize()

	var x, y int
	if config.RandomStart {
		x, y = rand.Intn(screenWidth), rand.Intn(screenHeight)
	} else {
		x, y = screenWidth/2, screenHeight/2
	}
	dx, dy := config.Speed, config.Speed

	stop := setupSignalHandler()

	printBanner(config)

	startTime := time.Now()
	hasTimeLimit := config.RunMinutes > 0
	timeLimit := time.Duration(config.RunMinutes) * time.Minute

	measurementCounter := 0
	userMovementDetected := false
	checkInterval := time.Duration(config.CheckInterval) * time.Millisecond
	lastCheckTime := time.Now()
	iterations := 0

	consecutiveDetections := 0
	requiredDetections := 3

	for {
		select {
		case <-stop:
			fmt.Println("\n🛑 Завершено вручную (Ctrl+C).")
			printStats(iterations, time.Since(startTime))
			return

		default:
			currentTime := time.Now()
			iterations++

			if hasTimeLimit && currentTime.Sub(startTime) >= timeLimit {
				fmt.Printf("\n⏱ Установленное время (%d мин) истекло — выход.\n", config.RunMinutes)
				printStats(iterations, currentTime.Sub(startTime))
				return
			}

			if currentTime.Sub(lastCheckTime) >= checkInterval {
				moveMouse(x, y)
				time.Sleep(20 * time.Millisecond)

				beforeX, beforeY := getMousePos()

				time.Sleep(50 * time.Millisecond)

				afterX, afterY := getMousePos()

				expectedDist := distance(beforeX, beforeY, x, y)
				actualDist := distance(afterX, afterY, x, y)
				movement := distance(beforeX, beforeY, afterX, afterY)

				if config.Verbose {
					fmt.Printf("\r🔍 Проверка: ожид=%.1f, факт=%.1f, движ=%.1f | Позиция: (%4d, %4d) | Итераций: %d",
						expectedDist, actualDist, movement, x, y, iterations)
				}

				if movement > config.Sensitivity {
					consecutiveDetections++

					if config.Verbose {
						fmt.Printf("\n⚠️  Обнаружено движение: %.1f px (попытка %d/%d)\n",
							movement, consecutiveDetections, requiredDetections)
					}

					if consecutiveDetections >= requiredDetections {
						fmt.Printf("\nПодтверждено движение мыши — выход.\n")
						printStats(iterations, time.Since(startTime))
						return
					}
				} else {
					consecutiveDetections = 0
				}

				lastCheckTime = currentTime
			} else {
				moveMouse(x, y)
			}

			time.Sleep(1 * time.Millisecond)
			actualX, actualY := getMousePos()
			deviation := distance(actualX, actualY, x, y)

			if deviation > config.DeviationLimit {
				measurementCounter++
				if !userMovementDetected {
					userMovementDetected = true
				}
			} else {
				measurementCounter = 0
				userMovementDetected = false
			}

			if measurementCounter >= 5 {
				fmt.Printf("\n⚠️  Обнаружено вмешательство (большое отклонение) — выход.\n")
				printStats(iterations, time.Since(startTime))
				return
			}

			x += dx
			y += dy

			if x <= 0 || x >= screenWidth {
				dx = -dx
				x = clamp(x, 0, screenWidth)
				if config.Verbose {
					fmt.Printf("\n💥 Отражение по X на границе %d\n", x)
				}
			}
			if y <= 0 || y >= screenHeight {
				dy = -dy
				y = clamp(y, 0, screenHeight)
				if config.Verbose {
					fmt.Printf("\n💥 Отражение по Y на границе %d\n", y)
				}
			}

			time.Sleep(15 * time.Millisecond)
		}
	}
}

func printBanner(config *Config) {
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║      🎬 DVD Screen Saver Effect 🎬        ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Printf("\n⚙️  Настройки:\n")
	fmt.Printf("   • Скорость: %d пикселей/шаг\n", config.Speed)
	fmt.Printf("   • Чувствительность: %.1f px\n", config.Sensitivity)
	fmt.Printf("   • Интервал проверки: %d мс\n", config.CheckInterval)
	fmt.Printf("   • Лимит отклонения: %.1f px\n", config.DeviationLimit)
	if config.RunMinutes > 0 {
		fmt.Printf("   • Таймер: %d мин\n", config.RunMinutes)
	} else {
		fmt.Printf("   • Таймер: выключен\n")
	}
	fmt.Println("\n🚀 Запуск... (Ctrl+C или пошевелите мышью для выхода)")
	fmt.Println()
}

func printStats(iterations int, duration time.Duration) {
	fmt.Println("\n📊 Статистика:")
	fmt.Printf("   • Итераций: %d\n", iterations)

	totalSeconds := int(duration.Seconds())
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	milliseconds := duration.Milliseconds() % 1000

	if hours > 0 {
		fmt.Printf("   • Время работы: %d ч %d мин %d сек\n", hours, minutes, seconds)
	} else if minutes > 0 {
		fmt.Printf("   • Время работы: %d мин %d сек\n", minutes, seconds)
	} else if seconds > 0 {
		fmt.Printf("   • Время работы: %d сек %d мс\n", seconds, milliseconds)
	} else {
		fmt.Printf("   • Время работы: %d мс\n", milliseconds)
	}

	fmt.Println("\n👋 До встречи!")
}

func distance(x1, y1, x2, y2 int) float64 {
	dx := float64(x1 - x2)
	dy := float64(y1 - y2)
	return math.Sqrt(dx*dx + dy*dy)
}

func clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
