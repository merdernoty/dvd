package main

import (
	"flag"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

const (
	version = "1.0.0"
)

type Config struct {
	Speed          int
	Sensitivity    float64
	CheckInterval  int
	DeviationLimit float64
	ShowVersion    bool
	Verbose        bool
	RandomStart    bool
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

	flag.Float64Var(&config.Sensitivity, "sensitivity", 5.0, "Чувствительность обнаружения движения мыши")
	flag.Float64Var(&config.Sensitivity, "sens", 5.0, "Чувствительность (короткая версия)")

	flag.IntVar(&config.CheckInterval, "interval", 50, "Интервал проверки движения (миллисекунды)")
	flag.IntVar(&config.CheckInterval, "i", 50, "Интервал проверки (короткая версия)")

	flag.Float64Var(&config.DeviationLimit, "deviation", 100.0, "Лимит отклонения для обнаружения вмешательства")
	flag.Float64Var(&config.DeviationLimit, "d", 100.0, "Лимит отклонения (короткая версия)")

	flag.BoolVar(&config.ShowVersion, "version", false, "Показать версию программы")
	flag.BoolVar(&config.ShowVersion, "v", false, "Показать версию (короткая версия)")

	flag.BoolVar(&config.Verbose, "verbose", false, "Подробный вывод")

	flag.BoolVar(&config.RandomStart, "random", true, "Случайная начальная позиция")
	flag.BoolVar(&config.RandomStart, "r", true, "Случайная позиция (короткая версия)")

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

	screenWidth, screenHeight := robotgo.GetScreenSize()

	var x, y int
	if config.RandomStart {
		x, y = rand.Intn(screenWidth), rand.Intn(screenHeight)
	} else {
		x, y = screenWidth/2, screenHeight/2
	}

	dx, dy := config.Speed, config.Speed

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	printBanner(config)

	measurementCounter := 0
	userMovementDetected := false
	checkInterval := time.Duration(config.CheckInterval) * time.Millisecond
	lastCheckTime := time.Now()
	iterations := 0

	for {
		select {
		case <-stop:
			fmt.Println("\n🛑 Завершено вручную (Ctrl+C).")
			printStats(iterations, time.Since(lastCheckTime))
			return

		default:
			currentTime := time.Now()
			iterations++

			if currentTime.Sub(lastCheckTime) >= checkInterval {
				beforeX, beforeY := robotgo.GetMousePos()
				time.Sleep(30 * time.Millisecond)
				afterX, afterY := robotgo.GetMousePos()

				if beforeX != afterX || beforeY != afterY {
					dist := distance(beforeX, beforeY, afterX, afterY)
					if dist > config.Sensitivity {
						fmt.Printf("\nОбнаружено движение мыши (%.1f px) — выход.\n", dist)
						printStats(iterations, time.Since(lastCheckTime))
						return
					}
				}

				lastCheckTime = currentTime
			}

			targetX, targetY := x, y

			robotgo.MoveMouse(x, y)
			time.Sleep(1 * time.Millisecond)

			actualX, actualY := robotgo.GetMousePos()
			deviation := distance(actualX, actualY, targetX, targetY)

			if config.Verbose && iterations%100 == 0 {
				fmt.Printf("\rПозиция: (%4d, %4d) | Отклонение: %.1f px | Итераций: %d",
					x, y, deviation, iterations)
			}

			if deviation > config.DeviationLimit {
				measurementCounter++
				if !userMovementDetected {
					userMovementDetected = true
				}
			} else {
				measurementCounter = 0
				userMovementDetected = false
			}

			if measurementCounter >= 3 {
				fmt.Printf("\nОбнаружено вмешательство (отклонение: %.1f px) — выход.\n", deviation)
				printStats(iterations, time.Since(lastCheckTime))
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
	fmt.Printf("\nНастройки:\n")
	fmt.Printf("   • Скорость: %d пикселей/шаг\n", config.Speed)
	fmt.Printf("   • Чувствительность: %.1f px\n", config.Sensitivity)
	fmt.Printf("   • Интервал проверки: %d мс\n", config.CheckInterval)
	fmt.Printf("   • Лимит отклонения: %.1f px\n", config.DeviationLimit)
	fmt.Println("\nЗапуск... (Ctrl+C или пошевелите мышью для выхода)")
	fmt.Println()
}

func printStats(iterations int, duration time.Duration) {
	fmt.Println("\nСтатистика:")
	fmt.Printf("   • Итераций: %d\n", iterations)
	fmt.Printf("   • Время работы: %v\n", duration.Round(time.Millisecond))
	fmt.Println("\nДо встречи!")
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
