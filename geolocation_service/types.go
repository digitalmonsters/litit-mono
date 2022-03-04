package geolocation_service

type GeolocationServiceI interface {
	GetLocationInfo(ip string) (*LocationInfo, error)
}

type LocationInfo struct {
	Status  ResponseStatus `json:"status"`
	Country string         `json:"country"`
	City    string         `json:"city"`
}

type ResponseStatus string

const (
	Success ResponseStatus = "success"
	Fail    ResponseStatus = "fail"
)

