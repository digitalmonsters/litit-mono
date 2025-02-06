package deeplink

import "github.com/digitalmonsters/go-common/eventsourcing"

type ServiceMock struct {
	GetVideoShareLinkFn              func(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error)
	GetPreviewShareLinkFn            func(contentId int64, contentType eventsourcing.ContentType, uri string, userId int64, referralCode string) (string, error)
	GetPetVideoShareLinkFn           func(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, petType int64, petId int64, petName string) (string, error)
	GetVideoShareLinkWithMetaFn      func(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, title string, description string, previewThumbnail string) (string, error)
	GenerateBranchDeeplinkWithMetaFn func(link string, title string, description string, previewThumbnail string) (string, error)
}

func (m ServiceMock) GenerateBranchDeeplinkWithMeta(link string, title string, description string, previewThumbnail string) (string, error) {
	return m.GenerateBranchDeeplinkWithMetaFn(link, title, description, previewThumbnail)
}

func (m ServiceMock) GetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error) {
	return m.GetVideoShareLinkFn(contentId, contentType, userId, referralCode)
}

func (m ServiceMock) GetPreviewShareLink(contentId int64, contentType eventsourcing.ContentType, uri string, userId int64, referralCode string) (string, error) {
	return m.GetPreviewShareLinkFn(contentId, contentType, uri, userId, referralCode)
}

func (m ServiceMock) GetPetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, petType int64, petId int64, petName string) (string, error) {
	return m.GetPetVideoShareLinkFn(contentId, contentType, userId, referralCode, petType, petId, petName)
}

func (m ServiceMock) GetVideoShareLinkWithMeta(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, title string, description string, previewThumbnail string) (string, error) {
	return m.GetVideoShareLinkWithMetaFn(contentId, contentType, userId, referralCode, title, description, previewThumbnail)
}

func GetMock() IService {
	return &ServiceMock{}
}
