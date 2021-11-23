package common

type AccessLevel byte

const (
	AccessLevelPublic = AccessLevel(0)
	AccessLevelRead   = AccessLevel(1)
)

func (a AccessLevel) ToString() string {
	switch a {
	case AccessLevelPublic:
		return "public"
	case AccessLevelRead:
		return "read"
	}

	return "unk"
}
