package deeplink

type ServiceMock struct {
	GetVideoShareLinkFn func(contentId int64, userId int64, referralCode string) (string, error)
}

func (m ServiceMock) GetVideoShareLink(contentId int64, userId int64, referralCode string) (string, error) {
	return m.GetVideoShareLinkFn(contentId, userId, referralCode)
}

func GetMock() IService {
	return &ServiceMock{}
}
