package terminal

import (
	"fmt"
	"os"
	"testing"

	"github.com/russross/blackfriday/v2"
)

func TestTerminal_Render(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{
			name: "test1",
			file: "../markdown_test.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf, err := os.ReadFile(tt.file)
			if err != nil {
				return
			}
			renderer := TerminalRenderer(0)

			output := blackfriday.Run(buf, blackfriday.WithRenderer(renderer))

			fmt.Println(string(output))

		})
	}
}
