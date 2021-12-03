package database

type Comment struct {
}

func (Comment) TableName() string {
	return "comment"
}

type CommentVote struct {
}

func (CommentVote) TableName() string {
	return "comment_vote"
}
