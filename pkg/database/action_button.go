package database

import "time"

type ActionButton struct {
	Id        int64
	Name      string
	Type      ButtonType
	CreatedAt time.Time
	DeletedAt time.Time
}

func (ActionButton) TableName() string {
	return "action_buttons"
}

type ButtonType byte

const (
	LinkButtonType    ButtonType = 1
	ContentButtonType ButtonType = 2
)
