package favorites

type AddToFavoritesRequest struct {
	UserId int64
	SongId string
}

type RemoveFromFavoritesRequest struct {
	UserId int64
	SongId string
}
