package cfg

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/andreyvit/diff"
	"golang.org/x/tools/go/cfg"
)

var (
	update = flag.Bool("u", false, "update the golden files of this test")
)

func TestDesugar(t *testing.T) {
	flag.Parse()
	file, err := os.Open("./testdata/")
	if err != nil {
		t.Fatal(err)
	}

	names, err := file.Readdirnames(-1)
	if err != nil {
		t.Fatal(err)
	}

	for _, n := range names {
		func(name string) {
			testname := "./testdata/" + name

			if strings.HasSuffix(testname, ".golden") {
				return
			}

			f, err := os.Open(testname)
			if err != nil {
				t.Fatal(err)
			}
			defer f.Close()

			// Create the AST by parsing src.
			fset := token.NewFileSet() // positions are relative to fset
			pf, err := parser.ParseFile(fset, "", f, 0)
			if err != nil {
				panic(err)
			}

			funDecl, ok := pf.Decls[0].(*ast.FuncDecl)
			if !ok {
				panic("not funDecl")
			}

			c := cfg.New(
				funDecl.Body,
				func(ce *ast.CallExpr) bool { return true },
			)

			b := NewBuilder()
			b.Build(c.Blocks[0])
			start := b.GetCfg()

			var bf bytes.Buffer

			err = PrintToWriter(start, &bf)
			if err != nil {
				t.Fatal(err)
			}

			goldenFilename := "./testdata/" + name + ".golden"
			if *update {
				f, err := os.Create(goldenFilename)
				if err != nil {
					t.Fatal(err)
				}
				defer f.Close()
				_, err = f.Write(bf.Bytes())
				if err != nil {
					t.Fatal(err)
				}
			} else {
				f, err := os.Open(goldenFilename)
				if err != nil {
					fmt.Println(bf.String())
					t.Log("golden file missing")
					t.Fail()
					return
				}
				defer f.Close()
				b, err := ioutil.ReadAll(f)
				if err != nil {
					t.Fatal(err)
				}
				if string(b) != bf.String() {
					t.Log(goldenFilename + " != " + testname)
					t.Log(diff.LineDiff(string(b), bf.String()))
					t.Fail()
				}
			}

		}(n)
	}
}
