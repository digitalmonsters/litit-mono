package renderer

import (
	"fmt"
	"testing"
)

func TestRender(t *testing.T) {
	rawTemplate := "You earned your first 100 LIT points for joining {{.website}}"

	resultText, resultErr := RenderText("test", rawTemplate,
		map[string]string{
			"point_amount": "100",
			"points_name":  "LIT",
			"website":      "Lit.it",
		})

	fmt.Println(resultText)

	if resultErr != nil {
		t.Fatal(resultErr)
	}

}
