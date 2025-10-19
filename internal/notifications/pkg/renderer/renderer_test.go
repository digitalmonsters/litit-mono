package renderer

import (
	"fmt"
	"github.com/digitalmonsters/go-common/translation"
	"github.com/digitalmonsters/notification-handler/pkg/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRender(t *testing.T) {
	contentLikeTemplate := database.RenderTemplate{
		Id:        "content_like",
		Kind:      "default",
		IsGrouped: true,
	}

	var (
		title            string
		body             string
		headline         string
		titleMultiple    string
		bodyMultiple     string
		headlineMultiple string
		err              error
	)

	title, body, headline, titleMultiple, bodyMultiple, headlineMultiple, err = Render(contentLikeTemplate, database.RenderingVariables{
		"firstname":          "firstname1",
		"lastname":           "lastname1",
		"notificationsCount": "2",
	}, translation.LanguageEn)
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)

	a.Equal("Lit.it", title)
	a.Equal("Lit.it", titleMultiple)
	a.Equal("firstname1 lastname1 liked your video", body)
	a.Equal("firstname1 lastname1 and 2 others liked your video", bodyMultiple)
	a.Equal("", headline)
	a.Equal("", headlineMultiple)

	title, body, headline, titleMultiple, bodyMultiple, headlineMultiple, err = Render(contentLikeTemplate, database.RenderingVariables{
		"firstname":          "firstname1",
		"lastname":           "lastname1",
		"notificationsCount": "2",
	}, translation.LanguageDe)
	if err != nil {
		t.Fatal(err)
	}

	a.Equal("Lit.it", title)
	a.Equal("", titleMultiple)
	a.Equal("firstname1 lastname1 gefällt dein Video", body)
	a.Equal("", bodyMultiple)
	a.Equal("", headline)
	a.Equal("", headlineMultiple)
}

func TestRenderText(t *testing.T) {
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
