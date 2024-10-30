package main

import (
	"fmt"
	"os"
	"github.com/forsvunnet/project-sync-tool/cmd/pst"
)

func main() {
    if err := pst.Execute(); err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}

