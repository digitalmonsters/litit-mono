package deeplink

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/digitalmonsters/go-common/http_client"
)

type Provider string

const (
	ProviderBranch   Provider = "branch"
	ProviderFirebase Provider = "firebase"
)

type Service struct {
	httpClient *http_client.HttpClient
	ctx        context.Context
	config     Config
	url        string
}

func NewService(config Config, httpClient *http_client.HttpClient, ctx context.Context) IService {
	url := "https://firebasedynamiclinks.googleapis.com/v1/shortLinks?key=" + config.Key
	return &Service{
		config:     config,
		url:        url,
		httpClient: httpClient,
		ctx:        ctx,
	}
}

func (s *Service) generateDeeplink(link string, provider Provider) (string, error) {
	switch provider {
	case ProviderFirebase:
		return s.generateFirebaseDeeplink(link)
	case ProviderBranch:
		return s.generateBranchDeeplink(link)
	default:
		return "", fmt.Errorf("unsupported provider:")
	}
}

func (s *Service) generateBranchDeeplink(link string) (string, error) {
	requestBody := BranchLinkRequest{
		BranchKey: s.config.Key,
		Data: map[string]string{
			"$deeplink_path": link,
			"$canonical_url": link,
		},
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false, // Disable HTTP/2 and use HTTP/1.1
		},
	}

	// Send the POST request
	req, err := http.NewRequest("POST", "https://api.branch.io/v1/url", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check for a valid response and ensure body exists
	if resp.Body == nil {
		return "", fmt.Errorf("response body is nil")
	}

	// Read the body only once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Print the response body for debugging
	fmt.Println("Response body: ", string(bodyBytes), "Status code:", resp.StatusCode)

	// Check for non-200 response status codes
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the response into the BranchLinkResponse struct
	var branchResponse BranchLinkResponse
	err = json.Unmarshal(bodyBytes, &branchResponse)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	return branchResponse.URL, nil
}

func (s *Service) generateFirebaseDeeplink(link string) (string, error) {
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
	resp, err := s.httpClient.NewRequest(s.ctx).SetContentType("application/json").SetBody(requestJsonBody).Post(s.url)

	if err != nil {
		return "", err
	}

	var firebaseResp firebaseCreateDeeplinkResponse

	if err := json.Unmarshal(resp.Bytes(), &firebaseResp); err != nil {
		return "", err
	}
	return firebaseResp.ShortLink, nil
}

func (s *Service) GetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string) (string, error) {
	link := fmt.Sprintf("%v/%v/%v", s.config.URI, s.getShareType(contentType), contentId)
	shareCode := fmt.Sprintf("%v%v%v", contentId, userId, time.Now().Unix())
	link += fmt.Sprintf("?sharerId=%v&referredByType=shared_content&shareCode=%v&selectedtype=%v", userId, shareCode, contentType)
	if len(referralCode) > 0 {
		link += fmt.Sprintf("&referralCode=%v", referralCode)
	}
	return s.generateDeeplink(link, ProviderBranch)
}

func (s *Service) GetPetVideoShareLink(contentId int64, contentType eventsourcing.ContentType, userId int64, referralCode string, petType int64, petId int64, petName string) (string, error) {
	link := fmt.Sprintf("%v/%v/%v", s.config.URI, s.getShareType(contentType), contentId)
	shareCode := fmt.Sprintf("%v%v%v", contentId, userId, time.Now().Unix())
	link += fmt.Sprintf("?sharerId=%v&referredByType=shared_content&shareCode=%v&petType=%v&petId=%v&petName=%s", userId, shareCode, petType, petId, petName)
	if len(referralCode) > 0 {
		link += fmt.Sprintf("&referralCode=%v", referralCode)
	}
	return s.generateDeeplink(link, ProviderBranch)
}

func (s *Service) GetPreviewShareLink(contentId int64, contentType eventsourcing.ContentType, uri string, userId int64, referralCode string) (string, error) {
	link := fmt.Sprintf("%v/%v/%v", uri, s.getShareType(contentType), contentId)
	shareCode := fmt.Sprintf("%v%v%v", contentId, userId, time.Now().Unix())
	link += fmt.Sprintf("?sharerId=%v&referredByType=shared_content&shareCode=%v", userId, shareCode)
	if len(referralCode) > 0 {
		link += fmt.Sprintf("&referralCode=%v", referralCode)
	}
	return s.generateDeeplink(link, ProviderBranch)
}

func (s *Service) getShareType(contentType eventsourcing.ContentType) string {
	switch contentType {
	case eventsourcing.ContentTypeMusic:
		return "music"
	case eventsourcing.ContentTypeSpot:
		return "spot"
	case eventsourcing.ContentTypePreview:
		return "preview"
	default:
		return "video"
	}
}

type BranchService struct {
	ctx    context.Context
	config Config
}

func NewBranchService(ctx context.Context, config Config) *BranchService {
	return &BranchService{
		ctx:    ctx,
		config: config,
	}
}

func (s *BranchService) GenerateBranchDeepLinkWithMeta(link string, title string, description string, previewThumbnail string) (string, error) {
	requestBody := BranchLinkRequest{
		BranchKey: s.config.Key,
		Data: map[string]string{
			"$deeplink_path":  link,
			"$canonical_url":  link,
			"$og_title":       title,
			"$og_description": description,
			"$og_image_url":   previewThumbnail,
		},
	}
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling JSON: %v", err)
	}

	httpClient := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false, // Disable HTTP/2 and use HTTP/1.1
		},
	}

	req, err := http.NewRequest("POST", "https://api.branch.io/v1/url", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating HTTP request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	// Check for a valid response and ensure body exists
	if resp.Body == nil {
		return "", fmt.Errorf("response body is nil")
	}

	// Read the body only once
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	// Print the response body for debugging
	fmt.Println("Response body: ", string(bodyBytes), "Status code:", resp.StatusCode)

	// Check for non-200 response status codes
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("non-200 status code: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// Decode the response into the BranchLinkResponse struct
	var branchResponse BranchLinkResponse
	err = json.Unmarshal(bodyBytes, &branchResponse)
	if err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}
	return branchResponse.URL, nil
}
