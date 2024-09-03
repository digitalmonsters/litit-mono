package deeplink

import "github.com/digitalmonsters/go-common/eventsourcing"

type ServiceMock struct {
	GetVideoShareLinkFn   func(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error)
	GetPreviewShareLinkFn func(contentId int64, contentType eventsourcing.ContentType, uri string, userId int64, referralCode string) (string, error)
}

func (m ServiceMock) GetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error) {
	return m.GetVideoShareLinkFn(contentId, contentType, userId, referralCode)
}

func (m ServiceMock) GetPreviewShareLink(contentId int64, contentType eventsourcing.ContentType, uri string, userId int64, referralCode string) (string, error) {
	return m.GetPreviewShareLinkFn(contentId, contentType, uri, userId, referralCode)
}

func (m ServiceMock) GetPetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, petType int64, petId int64, petName string) (string, error) {
	return m.GetPetVideoShareLink(contentId, contentType, userId, referralCode, petType, petId, petName)
}

func GetMock() IService {
	return &ServiceMock{}
}
