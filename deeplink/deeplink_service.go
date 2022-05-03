package deeplink

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/digitalmonsters/go-common/http_client"
)

type Service struct {
	httpClient *http_client.HttpClient
	ctx        context.Context
	config     Config
	url        string
}

func NewService(config Config, httpClient *http_client.HttpClient, ctx context.Context) *Service {
	url := "https://firebasedynamiclinks.googleapis.com/v1/shortLinks?key=" + config.Key
	return &Service{
		config:     config,
		url:        url,
		httpClient: httpClient,
		ctx:        ctx,
	}
}

func (s *Service) generateDeeplink(link string) (string, error) {
	requestBody := firebaseCreateDeeplinkRequest{
		DynamicLinkInfo: dynamicLinkInfo{
			DomainUriPrefix: s.config.DomainURIPrefix,
			Link:            link,
			AndroidInfo: androidInfo{
				AndroidPackageName: s.config.AndroidPackageName,
			},
			IosInfo: iosInfo{
				IosBundleId:   s.config.IOSBundleId,
				IosAppStoreId: s.config.IOSAppStoreId,
			},
		},
	}
	requestJsonBody, err := json.Marshal(requestBody)

	if err != nil {
		return "", err
	}
	resp, err := s.httpClient.NewRequest(s.ctx).SetBody(requestJsonBody).Post(s.url)

	if err != nil {
		return "", err
	}

	var firebaseResp firebaseCreateDeeplinkResponse

	if err := json.Unmarshal(resp.Bytes(), &firebaseResp); err != nil {
		return "", err
	}
	return firebaseResp.ShortLink, nil
}

func (s *Service) GetVideoShareLink(contentId int64, userId int64, referralCode string) (string, error) {
	link := fmt.Sprintf("%v/video/%v", s.config.URI, contentId)
	link += fmt.Sprintf("?sharerId=%v&referredByType=shared_content", userId)
	if len(referralCode) > 0 {
		link += fmt.Sprintf("&referralCode=%v", referralCode)
	}
	return s.generateDeeplink(link)
}
