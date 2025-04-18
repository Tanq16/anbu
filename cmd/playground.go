package cmd

import (
	"fmt"
	"os"
	"strconv"

	"github.com/tanq16/anbu/utils"
)

func playground() {
	fmt.Println(utils.OutError("Playground mode executed"))
	variable := os.Getenv("COLUMNS")
	size, _ := strconv.Atoi(variable)
	fmt.Println(utils.OutDetail(fmt.Sprintf("Terminal size: %d", size)))
}
