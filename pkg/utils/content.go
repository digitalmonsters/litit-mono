package utils

import (
	"fmt"
	"github.com/digitalmonsters/notification-handler/configs"
)

func GetAnimUrl(videoId string) string {
	return fmt.Sprintf("%v/%v/%v/a.webp", configs.CDN_BASE, configs.PREFIX_CONTENT, videoId)
}

func GetThumbnailUrl(videoId string) string {
	return fmt.Sprintf("%v/%v/%v/t.jpg", configs.CDN_BASE, configs.PREFIX_CONTENT, videoId)
}

func GetVideoUrl(videoId string) string {
	return fmt.Sprintf("%v/%v/%v/v.mp4", configs.CDN_BASE, configs.PREFIX_CONTENT, videoId)
}

func GetIsVertical(width, height int) bool {
	if width == 0 || height == 0 {
		return false
	}

	return height > width
}
