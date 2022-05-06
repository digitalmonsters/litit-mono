package eventsourcing

type ConfigEvent struct {
	Key   string
	Value string
}

func (c ConfigEvent) GetPublishKey() string {
	return c.Key
}
