package main

import "fmt"

// Значения переменных сборки задаются флагами компоновщика:
//
//	go build -ldflags "-X main.buildVersion=v1.0.0 -X 'main.buildDate=$(date)' -X main.buildCommit=$(git rev-parse HEAD)"
var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

// printBuildInfo выводит сведения о сборке. Незаданные значения
// заменяются на N/A.
func printBuildInfo() {
	fmt.Printf("Build version: %s\n", orNA(buildVersion))
	fmt.Printf("Build date: %s\n", orNA(buildDate))
	fmt.Printf("Build commit: %s\n", orNA(buildCommit))
}

// orNA возвращает значение или N/A, если значение пустое.
func orNA(value string) string {
	if value == "" {
		return "N/A"
	}

	return value
}
