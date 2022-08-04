package answer

import (
	"fmt"
	"strings"
)

type Answer int

const (
	NO Answer = iota
	YES
	YES_TO_ALL
)

func Ask() Answer {
	fmt.Printf("Please enter Y (Yes) or N (No) or A (Yes to All): ")

	var answer string
	if _, err := fmt.Scanln(&answer); err == nil && len(answer) > 0 {
		switch strings.ToUpper(string(answer[0])) {
		case "Y":
			return YES
		case "A":
			return YES_TO_ALL
		}
	}

	return NO
}
