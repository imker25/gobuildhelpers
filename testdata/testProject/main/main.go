package main

import (
	"fmt"
	"os"
)

func main() {
	res := Add(1, 2)

	fmt.Println(fmt.Sprintf("The result is '%d'", res))

	os.Exit(0)

}
