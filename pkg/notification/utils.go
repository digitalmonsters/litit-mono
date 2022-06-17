package notification

func GetUserNotificationsReadClusterKey(userId int64) int64 {
	return userId / 10000
}
