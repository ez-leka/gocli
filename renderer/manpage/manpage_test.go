package manpage

import (
	"os"
	"testing"

	"github.com/russross/blackfriday/v2"
)

func TestManpage_Render(t *testing.T) {
	tests := []struct {
		name string
		file string
	}{
		{
			name: "test1",
			file: "../markdown_test1.md",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			buf, err := os.ReadFile(tt.file)
			if err != nil {
				return
			}
			renderer := TRoffRenderer("RENDERER_TEST")

			output := blackfriday.Run(buf, blackfriday.WithRenderer(renderer))

			err = os.WriteFile("test.1", output, 0644)
			if err != nil {
				panic(err)
			}

		})
	}
}
