package vote

import (
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
)

var db *gorm.DB

func TestMain(m *testing.M) {
	db = database.GetDb()
	os.Exit(m.Run())
}

func baseSetup(t *testing.T) {
	cfg := configs.GetConfig()

	if err := boilerplate_testing.FlushPostgresTables(cfg.Db,
		[]string{"public.comment", "public.comment_vote", "public.content", "public.comment_vote", "public.profile"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := utils.PollutePostgresDatabase(db, "../comments/test_data/seed.json"); err != nil {
		t.Fatal(err)
	}
}

func TestVoteComment(t *testing.T) {
	baseSetup(t)
	vote, err := VoteComment(db, 9700, null.BoolFrom(true), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(int64(9700), vote.CommentId)
	a.Equal(int64(1), vote.UserId)
	a.Equal(true, vote.VoteUp.ValueOrZero())

	var comment database.Comment
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(1), comment.NumUpvotes)
	a.Equal(int64(0), comment.NumDownvotes)

	_, err = VoteComment(db, 9700, null.BoolFrom(true), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(1), comment.NumUpvotes)
	a.Equal(int64(0), comment.NumDownvotes)

	_, err = VoteComment(db, 9700, null.BoolFrom(false), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(0), comment.NumUpvotes)
	a.Equal(int64(1), comment.NumDownvotes)

	_, err = VoteComment(db, 9700, null.NewBool(false, false), 1, nil, nil, nil, &content.ContentWrapperMock{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Where("id = ?", 9700).First(&comment).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(0), comment.NumUpvotes)
	a.Equal(int64(0), comment.NumDownvotes)
}
