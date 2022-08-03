package eventsourcing

import (
	"fmt"
	"time"
)

type UserTotalWatchTimeEvent struct { // local.total_watch_time
	UserId         int64     `json:"user_id"`
	TotalWatchTime int64     `json:"total_watch_time"`
	CreatedAt      time.Time `json:"created_at"`
}

func (v UserTotalWatchTimeEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", v.UserId)
}

type UserTodayWatchTimeEvent struct { // local.today_watch_time
	UserId              int64     `json:"user_id"`
	TotalTodayWatchTime int64     `json:"total_today_watch_time"`
	CreatedAt           time.Time `json:"created_at"`
}

func (v UserTodayWatchTimeEvent) GetPublishKey() string {
	return fmt.Sprintf("%v", v.UserId)
}

type SourceView int

const (
	SourceViewDirectlyById  = SourceView(0)  // content/{id}
	SourceViewFeedCountry   = SourceView(1)  // content/feed/country
	SourceViewFeedTop       = SourceView(2)  // content/spots/top/week
	SourceViewFeedFollowing = SourceView(3)  // content/feed/following
	SourceViewFeedInterests = SourceView(4)  // content/feed/interests
	SourceViewFeedSpots     = SourceView(5)  // content/feed/spots
	SourceViewSpotsMe       = SourceView(6)  // content/spots/me
	SourceViewSpotsTopDay   = SourceView(7)  // content/spots/top/day
	SourceViewSpotsTopWeek  = SourceView(8)  // content/spots/top/week
	SourceViewSearchVideos  = SourceView(9)  // content/search?type=videos
	SourceViewSearchTop     = SourceView(10) // content/search?type=top
	SourceViewMe            = SourceView(11) // content/me
	SourceViewMeSearch      = SourceView(12) // content/me/search
	SourceViewProfile       = SourceView(13) // content/profile

	SourceViewDiscovery         = SourceView(14) // discovery/v2/discovery
	SourceViewDiscoveryCategory = SourceView(15) // discovery/v2/discovery/category
	SourceViewDiscoveryHashtag  = SourceView(16) // discovery/v2/discovery/hashtag
	SourceViewDiscoveryNoFollow = SourceView(17) // discovery/v2/discovery/no_follow
	SourceViewDiscoveryTopUsers = SourceView(18) // discovery/v2/discovery/top_users

	SourceViewNotificationCommentReply   = SourceView(19) // push.comment.reply
	SourceViewPushCommentReply           = SourceView(20) // comment_reply
	SourceViewNotificationCommentContent = SourceView(21) // push.content.comment
	SourceViewPushCommentContent         = SourceView(22) // comment_content_resource_create
	SourceViewNotificationCommentProfile = SourceView(23) // push.profile.comment
	SourceViewPushCommentProfile         = SourceView(24) // comment_profile_resource_create

	SourceViewNotificationCommentVoteLike    = SourceView(25) // push.comment.vote
	SourceViewPushCommentVoteLike            = SourceView(26) // comment_vote_like
	SourceViewNotificationCommentVoteDislike = SourceView(27) // push.comment.vote
	SourceViewPushCommentVoteDislike         = SourceView(28) // comment_vote_dislike

	SourceViewNotificationContentUpload    = SourceView(29) // push.content.successful-upload
	SourceViewPushContentUpload            = SourceView(30) // content_upload
	SourceViewNotificationSpotUpload       = SourceView(31) // push.spot.successful-upload
	SourceViewPushSpotUpload               = SourceView(32) // spot_upload
	SourceViewNotificationContentRejected  = SourceView(33) // push.content.rejected
	SourceViewPushContentRejected          = SourceView(34) // content_reject
	SourceViewNotificationContentNewPosted = SourceView(35) // push.content.new-posted
	SourceViewPushContentNewPosted         = SourceView(36) // content_posted

	SourceViewNotificationContentLike = SourceView(37) // push.content.like
	SourceViewPushContentLike         = SourceView(38) // content_like

	SourceViewNotificationAdmin = SourceView(39) // push.admin.bulk
	SourceViewPushAdmin         = SourceView(40) // push_admin
)
