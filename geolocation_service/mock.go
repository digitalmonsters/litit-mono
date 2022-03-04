package geolocation_service

type GeolocationServiceMock struct {}

func (m *GeolocationServiceMock) GetLocationInfo(ip string) (*LocationInfo, error) {
	return &LocationInfo{
		Status:  Success,
		Country: "test",
		City:    "test",
	}, nil
}
