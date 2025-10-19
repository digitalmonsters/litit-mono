package translation

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestGetTranslation(t *testing.T) {
	value, fromRequestedLanguage := GetTranslation(DefaultUserLanguage, LanguageEn, "notifications", "increase_reward_stage_2_title")

	a := assert.New(t)

	a.True(value.Valid)
	a.True(strings.Contains(value.String, "{{.percent_multiplier}}% increased rewards for friends invitations."))
	a.True(fromRequestedLanguage)

	value, fromRequestedLanguage = GetTranslation(DefaultUserLanguage, "unknown", "notifications", "increase_reward_stage_2_title")

	a.True(value.Valid)
	a.True(strings.Contains(value.String, "{{.percent_multiplier}}% increased rewards for friends invitations."))
	a.False(fromRequestedLanguage)

	value, fromRequestedLanguage = GetTranslation(DefaultUserLanguage, LanguageEn, "notifications", "unknown_notification_key")

	a.False(value.Valid)
	a.True(fromRequestedLanguage)
}
