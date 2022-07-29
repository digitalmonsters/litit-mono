package feed

type CursorPaging struct {
	Before string `json:"before"`
	After  string `json:"after"`
}
