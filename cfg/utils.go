package cfg

import (
	"fmt"
	"io"
	"os"
)

func Print(nodes []Node) {
	printNodes(nodes, os.Stdout)
}

func PrintToWriter(nodes []Node, w io.Writer) error {
	printNodes(nodes, w)
	return nil
}

func printNodes(nodes []Node, w io.Writer) {
	for _, n := range nodes {
		fmt.Fprintln(w, ToString(n))
	}
}
