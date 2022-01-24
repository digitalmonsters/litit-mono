package common

type AccessLevel byte

const (
	AccessLevelPublic  = AccessLevel(0)
	AccessLevelRead    = AccessLevel(1)
	AccessLevelWrite   = AccessLevel(2)
	AccessLevelService = AccessLevel(3)
)

func (a AccessLevel) ToString() string {
	switch a {
	case AccessLevelPublic:
		return "public"
	case AccessLevelRead:
		return "read"
	case AccessLevelWrite:
		return "write"
	case AccessLevelService:
		return "service"
	}

	return "unk"
}
