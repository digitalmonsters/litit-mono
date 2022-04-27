package eventsourcing

// see BaseChangeEvent in user topic

type KycStatusType string

const (
	KycStatusNone        = KycStatusType("none")
	KycStatusPending     = KycStatusType("pending")
	KycStatusRejected    = KycStatusType("rejected")
	KycStatusVerified    = KycStatusType("verified")
	KycStatusRobotReview = KycStatusType("robot_review")
)

type KycReason string

const (
	KycReasonDocument  = KycReason("document")
	KycReasonAge       = KycReason("age")
	KycReasonPhoto     = KycReason("photo")
	KycReasonSuspicion = KycReason("suspicion")
)
