package common

type DeviceType string

const (
	DeviceTypeIos     = DeviceType("ios")
	DeviceTypeAndroid = DeviceType("android")
	DeviceTypeWeb     = DeviceType("web")
)
