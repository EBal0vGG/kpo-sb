package cmdutil

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
)

func AskInt(r *bufio.Reader, prompt string) int {
	for {
		fmt.Print(prompt)
		line, _ := r.ReadString('\n')
		line = strings.TrimSpace(line)
		if v, err := strconv.Atoi(line); err == nil {
			return v
		}
		fmt.Println("Enter integer, please")
	}
}

func AskIntMin(r *bufio.Reader, prompt string, min int) int {
	for {
		v := AskInt(r, prompt)
		if v >= min {
			return v
		}
		fmt.Printf("Value must be >= %d\n", min)
	}
}

func AskIntRange(r *bufio.Reader, prompt string, min, max int) int {
	for {
		v := AskInt(r, prompt)
		if v >= min && v <= max {
			return v
		}
		fmt.Printf("Value must be in [%d..%d]\n", min, max)
	}
}
