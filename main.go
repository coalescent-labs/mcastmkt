package main

import (
	"fmt"
	"github.com/coalescent-labs/mcastmkt/cmd"
)

func main() {
	err := cmd.Execute()
	if err != nil && err.Error() != "" {
		fmt.Println(err)
	}
}
