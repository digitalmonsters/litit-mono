package utils

import (
	"fmt"

	"github.com/digitalmonsters/ads-manager/configs"
)

func GetAnimUrl(videoId string) string {
	return fmt.Sprintf("%v/%v/preview.webp", configs.CDN_BASE, videoId)
}

func GetThumbnailUrl(videoId string) string {
	return fmt.Sprintf("%v/%v/thumbnail.jpg", configs.CDN_BASE, videoId)
}

func GetVideoUrl(videoId string) string {
	return fmt.Sprintf("%v/%v/playlist.m3u8", configs.CDN_BASE, videoId)
}

func GetIsVertical(width, height int) bool {
	if width == 0 || height == 0 {
		return false
	}

	return height > width
}
