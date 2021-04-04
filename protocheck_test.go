package protochecks

import (
	"testing"

	"golang.org/x/tools/go/analysis/analysistest"
)

func TestBasicCase(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, Analyzer, "basic")
}
