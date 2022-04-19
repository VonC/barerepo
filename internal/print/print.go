package print

import (
	"bufio"
	"fmt"
	"os"
)

func Printf(msg string) {
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
