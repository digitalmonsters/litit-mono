package eventsourcing

// see BaseChangeEvent in user topic

type KycStatusType string

const (
	KycStatusNone     = KycStatusType("none")
	KycStatusPending  = KycStatusType("pending")
	KycStatusRejected = KycStatusType("rejected")
	KycStatusVerified = KycStatusType("verified")
)

type KycType string

const (
	KycTypeStatusUpdated = KycType("kyc_status_updated")
)

type KycReason string

const (
	KycReasonDocument  = KycReason("document")
	KycReasonAge       = KycReason("age")
	KycReasonPhoto     = KycReason("photo")
	KycReasonSuspicion = KycReason("suspicion")
)
