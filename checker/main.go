package main

import (
	protochecks "github.com/vnarek/proto-checks"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(protochecks.Analyzer)
}
