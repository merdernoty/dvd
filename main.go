package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-vgo/robotgo"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	screenWidth, screenHeight := robotgo.GetScreenSize()
	x, y := rand.Intn(screenWidth), rand.Intn(screenHeight)
	dx, dy := 10, 10

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	fmt.Println("Запуск эффекта DVD... (Ctrl+C для выхода)")
	fmt.Println("Дёрни мышь, чтобы остановить программу.")

	measurementCounter := 0
	userMovementDetected := false

	checkInterval := 50 * time.Millisecond
	lastCheckTime := time.Now()

	for {
		select {
		case <-stop:
			fmt.Println("\nЗавершено вручную (Ctrl+C).")
			return

		default:
			currentTime := time.Now()

			if currentTime.Sub(lastCheckTime) >= checkInterval {
				beforeX, beforeY := robotgo.GetMousePos()
				time.Sleep(30 * time.Millisecond)
				afterX, afterY := robotgo.GetMousePos()
				if beforeX != afterX || beforeY != afterY {
					dist := distance(beforeX, beforeY, afterX, afterY)
					if dist > 5 {
						fmt.Printf("\nОбнаружено движение мыши пользователем, выход.\n")
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

			if deviation > 100 {
				measurementCounter++
				if !userMovementDetected {
					userMovementDetected = true
				}
			} else {
				measurementCounter = 0
				userMovementDetected = false
			}
			if measurementCounter >= 3 {
				fmt.Printf("\nОбнаружено вмешательство пользователя — выход.\n")
				return
			}

			x += dx
			y += dy

			if x <= 0 || x >= screenWidth {
				dx = -dx
				x = clamp(x, 0, screenWidth)
			}
			if y <= 0 || y >= screenHeight {
				dy = -dy
				y = clamp(y, 0, screenHeight)
			}

			time.Sleep(15 * time.Millisecond)
		}
	}
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
