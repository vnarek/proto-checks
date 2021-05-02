package cfg

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
)

func PrintNodes(n Node) {
	SuccPrint(n, 0, make(map[Node]struct{}), os.Stdout)
}

func PrintToWriter(n Node, r io.Writer) error {
	SuccPrint(n, 0, make(map[Node]struct{}), r)
	return nil
}

func SuccPrint(n Node, depth int, printed map[Node]struct{}, r io.Writer) {
	printed[n] = struct{}{}
	fmt.Fprint(r, strings.Repeat("  ", depth))
	fmt.Fprintln(r, ToString(n))
	succArr := make([]Node, 0, len(n.Succ()))

	for i := range n.Succ() {
		succArr = append(succArr, i)
	}
	sort.Slice(succArr, func(i, j int) bool {
		return succArr[i].Id() < succArr[j].Id()
	})

	for _, k := range succArr {
		if _, ok := k.Pred()[n]; !ok {
			panic("panic")
		}
		if _, ok := printed[k]; ok {
			fmt.Fprint(r, strings.Repeat("  ", depth+1))
			fmt.Fprintln(r, "[connects to: "+ToString(k)+"]")
		} else {
			SuccPrint(k, depth+1, printed, r)
		}
	}
}
