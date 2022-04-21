package print

import (
	"bufio"
	"fmt"
	"os"
	"runtime"
	"strconv"
	"strings"
)

func Printf(format string, a ...any) {
	// https://stackoverflow.com/a/25322013/6309 , https://stackoverflow.com/questions/8605446/appending-items-to-a-variadic-function-wrapper-without-reallocating-a-new-slice
	// https://go.dev/play/p/xXEO58BJoA
	a = append(a, "a")
	copy(a[1:], a[0:])
	a[0] = goid()
	msg := fmt.Sprintf("[%d] "+format, a...)
	// https://github.com/golang/go/issues/36619
	bufStdout := bufio.NewWriter(os.Stdout)
	_, err := bufStdout.WriteString(msg)
	defer bufStdout.Flush()
	if err == nil {
		fmt.Fprintf(bufStdout, "\n")
	} else {
		fmt.Printf("Error on Printf '%s': %+v\n", msg, err)
	}
}

// https://gist.github.com/metafeather/3615b23097836bc36579100dac376906
// https://blog.sgmansfield.com/2015/12/goroutine-ids/
func goid() int {
	var buf [64]byte
	n := runtime.Stack(buf[:], false)
	idField := strings.Fields(strings.TrimPrefix(string(buf[:n]), "goroutine "))[0]
	id, err := strconv.Atoi(idField)
	if err != nil {
		panic(fmt.Sprintf("cannot get goroutine id: %v", err))
	}
	return id
}
