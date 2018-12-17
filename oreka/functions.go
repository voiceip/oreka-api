package oreka

import (
	"runtime"
	"fmt"
	"github.com/fatih/color"
	"bufio"
	"os"
	"io"
	"crypto/md5"
	"encoding/hex"
)

func MD5(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

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

type deleteCloser struct {
	io.ReadCloser
	path string
	f *os.File
}

func (d *deleteCloser) Close() error {
	err := d.ReadCloser.Close()
	if err != nil {
		return err
	}
	err = os.Remove(d.path)
	if err != nil {
		fmt.Println("ERROR DeleteOnCloseReader", d.path)
		return err
	}
	return nil
}

func (d *deleteCloser) Size() int64 {
	fi, _ := d.f.Stat()
	return fi.Size()
}

func DeleteOnCloseReader(path string) (*deleteCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return &deleteCloser{f, path, f}, nil
}
