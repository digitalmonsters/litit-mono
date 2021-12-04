package publicapi



type BlockedUserType string

const (
	BlockedUser   BlockedUserType = "BLOCKED USER"
	BlockedByUser BlockedUserType = "YOUR PROFILE IS BLOCKED BY USER"
)
