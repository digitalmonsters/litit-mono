package eventsourcing

import (
	"fmt"
	"gopkg.in/guregu/null.v4"
)

type ProfileEvent struct {
	UserId          int64       `json:"user_id"`
	Bio             null.String `json:"bio"`
	AddressCity     null.String `json:"address_city"`
	SocialReddit    null.String `json:"social_reddit"`
	SocialQuora     null.String `json:"social_quora"`
	SocialMedium    null.String `json:"social_medium"`
	SocialLinkedin  null.String `json:"social_linkedin"`
	SocialDiscord   null.String `json:"social_discord"`
	SocialTelegram  null.String `json:"social_telegram"`
	SocialViber     null.String `json:"social_viber"`
	SocialWhatsapp  null.String `json:"social_whatsapp"`
	SocialFacebook  null.String `json:"social_facebook"`
	SocialInstagram null.String `json:"social_instagram"`
	SocialTwitter   null.String `json:"social_twitter"`
	SocialWebsite   null.String `json:"social_website"`
	SocialYoutube   null.String `json:"social_youtube"`
	SocialTiktok    null.String `json:"social_tiktok"`
	SocialClubhouse null.String `json:"social_clubhouse"`
	SocialTwitch    null.String `json:"social_twitch"`
	BaseChangeEvent
}

func (c ProfileEvent) GetPublishKey() string {
	return fmt.Sprint(c.UserId)
}
