package global

import "github.com/digitalmonsters/music/pkg/database"

type INotifier interface {
	Enqueue(userId int64, data *database.Creator)
	Flush() []error
	Close() []error
}
