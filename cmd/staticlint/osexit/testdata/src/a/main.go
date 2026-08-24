package main

import (
	"fmt"
	"os"
)

func helper() {
	os.Exit(1)
}

func main() {
	defer fmt.Println("не выполнится")

	if len(os.Args) > 1 {
		os.Exit(1) // want "прямой вызов os.Exit в функции main запрещён"
	}

	helper()

	func() {
		os.Exit(2) // want "прямой вызов os.Exit в функции main запрещён"
	}()

	os.Exit(0) // want "прямой вызов os.Exit в функции main запрещён"
}
