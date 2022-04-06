package database

import (
	"github.com/digitalmonsters/go-common/eventsourcing"
	"github.com/lib/pq"
	"github.com/shopspring/decimal"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"time"
)

type Playlist struct {
	Id         int64          `json:"id" gorm:"primaryKey"`
	Name       string         `json:"name"`
	SortOrder  int            `json:"sort_order"`
	Color      string         `json:"color"`
	SongsCount int            `json:"songs_count"`
	IsActive   bool           `json:"is_active"`
	CreatedAt  time.Time      `json:"created_at"`
	DeletedAt  gorm.DeletedAt `json:"deleted_at"`
}

func (Playlist) TableName() string {
	return "playlists"
}

type Song struct {
	Id           int64          `json:"id" gorm:"primaryKey"`
	Source       SongSource     `json:"source"`
	ExternalId   string         `json:"external_id"`
	Title        string         `json:"title"`
	Artist       string         `json:"artist"`
	ImageUrl     string         `json:"image_url"`
	Genre        string         `json:"genre"`
	Duration     float64        `json:"duration"`
	ListenAmount int            `json:"listen_amount"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at"`
}

type SongSource int

const (
	SongSourceOwnStorage  = SongSource(1)
	SongSourceSoundStripe = SongSource(2)
)

func (Song) TableName() string {
	return "songs"
}

type PlaylistSongRelations struct {
	PlaylistId int64
	SongId     int64
	SortOrder  int
}

type Favorite struct {
	UserId    int64
	SongId    int64
	CreatedAt time.Time
}

type MusicStorage struct {
	Id          int64          `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Artist      string         `json:"artist"`
	ImageUrl    string         `json:"image_url"`
	Genre       string         `json:"genre"`
	Duration    float64        `json:"duration"`
	Url         string         `json:"url"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at"`
}

func (MusicStorage) TableName() string {
	return "music_storage"
}

type Creator struct {
	Id           int64                       `json:"id"`
	UserId       int64                       `json:"user_id"`
	Status       eventsourcing.CreatorStatus `json:"status"`
	RejectReason null.Int                    `json:"reject_reason"`
	LibraryUrl   string                      `json:"library_url"`
	SongsCount   int                         `json:"songs_count"`
	CreatedAt    time.Time                   `json:"created_at"`
	ApprovedAt   null.Time                   `json:"approved_at"`
	DeletedAt    gorm.DeletedAt              `json:"deleted_at"`

	Reason *CreatorRejectReasons `json:"-" gorm:"foreignKey:reject_reason"`
}

func (Creator) TableName() string {
	return "creators"
}

type ReasonType int

const (
	ReasonTypeNone         = ReasonType(0)
	ReasonTypeMusicCreator = ReasonType(1)
	ReasonTypeCreatorSong  = ReasonType(2)
)

type CreatorRejectReasons struct {
	Id        int64
	Type      ReasonType
	Reason    string
	CreatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (CreatorRejectReasons) TableName() string {
	return "creator_reject_reasons"
}

type Category struct {
	Id         int64
	Name       string
	SortOrder  int
	IsActive   bool
	SongsCount int
	CreatedAt  time.Time
	UpdatedAt  null.Time
	DeletedAt  gorm.DeletedAt
}

func (Category) TableName() string {
	return "categories"
}

type CreatorSong struct {
	Id                int64             `json:"id"`
	UserId            int64             `json:"user_id"`
	Name              string            `json:"name"`
	Status            CreatorSongStatus `json:"status"`
	LyricAuthor       null.String       `json:"lyric_author"`
	MusicAuthor       string            `json:"music_author"`
	CategoryId        int64             `json:"category_id"`
	MoodId            int64             `json:"mood_id"`
	FullSongUrl       string            `json:"full_song_url"`
	FullSongDuration  float64           `json:"full_song_duration"`
	ShortSongUrl      string            `json:"short_song_url"`
	ShortSongDuration float64           `json:"short_song_duration"`
	ImageUrl          string            `json:"image_url"`
	Hashtags          pq.StringArray    `gorm:"type:text[]" json:"hashtags"`
	ShortListens      int               `json:"short_listens"`
	FullListens       int               `json:"full_listens"`
	Likes             int               `json:"likes"`
	Comments          int               `json:"comments"`
	UsedInVideo       int               `json:"used_in_video"`
	PointsEarned      decimal.Decimal   `json:"points_earned"`
	RejectReason      null.Int          `json:"reject_reason"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         null.Time         `json:"updated_at"`
	DeletedAt         gorm.DeletedAt    `json:"deleted_at"`

	Category *Category             `gorm:"foreignKey:category_id" json:"-"`
	Mood     *Mood                 `gorm:"foreignKey:mood_id" json:"-"`
	Reject   *CreatorRejectReasons `gorm:"foreignKey:reject_reason" json:"-"`
}

func (CreatorSong) TableName() string {
	return "creator_songs"
}

type CreatorSongStatus int

const (
	CreatorSongStatusNone      = CreatorSongStatus(0)
	CreatorSongStatusPublished = CreatorSongStatus(1)
	CreatorSongStatusRejected  = CreatorSongStatus(2)
	CreatorSongStatusApproved  = CreatorSongStatus(3)
)

type Mood struct {
	Id         int64
	Name       string
	SortOrder  int
	SongsCount int
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  null.Time
	DeletedAt  gorm.DeletedAt
}

func (Mood) TableName() string {
	return "moods"
}
