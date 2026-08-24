package osexit

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestOSExitInMain(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "a")
}

func TestOSExitOutsideMainPackage(t *testing.T) {
	analysistest.Run(t, analysistest.TestData(), Analyzer, "b")
}
