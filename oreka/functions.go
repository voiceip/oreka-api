package oreka

import (
	"runtime"
	"fmt"
	"github.com/fatih/color"
	"bufio"
	"os"
)

func Check(e error) {
	if e != nil {
		panic(e)
	}
}

func Die(msg string, e error) {
	if runtime.GOOS == "windows" {
		fmt.Println("ERROR:", msg)
	} else {
		fmt.Println(color.RedString(msg))
	}
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	Check(e)
	os.Exit(1)
}
