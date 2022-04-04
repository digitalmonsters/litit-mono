package comments

import (
	user "github.com/digitalmonsters/go-common/wrappers/user_go"
	"strings"
	"testing"

	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"
)

var contentWrapperMock content.IContentWrapper
var mockContentRecord = content.SimpleContent{
	Id:            1017738,
	Duration:      10,
	AgeRestricted: false,
	AuthorId:      1,
	CategoryId:    null.Int{},
	Hashtags:      nil,
}

var userType = user.BlockedUser

var mockBlockRecord = user.UserBlockData{
	Type:      &userType,
	IsBlocked: true,
}
var mockNotBlockRecord = user.UserBlockData{
	Type:      nil,
	IsBlocked: false,
}

func TestCreateComment(t *testing.T) {
	baseSetup(t)
	comment, err := CreateComment(db, 1017738, "test_create_comment", null.NewInt(0, false),
		contentWrapperMock, userWrapperMock, nil, 10, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var c1 database.Comment
	if err := db.Where("id = ?", comment.Id).First(&c1).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(int64(1017738), c1.ContentId.ValueOrZero())
	a.Equal("test_create_comment", c1.Comment)
	a.Equal(false, c1.ParentId.Valid)

	comment2, err := CreateComment(db, 1017738, "test_create_comment2", null.IntFrom(comment.Id),
		contentWrapperMock, userWrapperMock, nil, 10, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var c2 database.Comment
	if err := db.Where("id = ?", comment2.Id).First(&c2).Error; err != nil {
		t.Fatal(err)
	}

	a.Equal(int64(1017738), c2.ContentId.ValueOrZero())
	a.Equal("test_create_comment2", c2.Comment)
	a.Equal(c1.Id, c2.ParentId.ValueOrZero())

	var content database.Content
	if err := db.Where("id = 1017738").First(&content).Error; err != nil {
		t.Fatal(err)
	}
	a.Equal(int64(2), content.CommentsCount)
	_, err = CreateComment(db, 1017738, "test_create_comment2", null.IntFrom(comment.Id), contentWrapperMock, userWrapperMock,
		nil, 1, nil, nil, nil)
	a.NotEqual(nil, err)

}

func TestUpdateCommentById(t *testing.T) {
	baseSetup(t)
	updatedComment, err := UpdateCommentById(db, 9700, "updated comment", 1074240,
		nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var comment database.Comment
	if err := db.Where("id = 9700").First(&comment).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(updatedComment.Comment, comment.Comment)

	updatedComment, err = UpdateCommentById(db, 1, "updated comment", 1074240,
		nil, nil, nil)
	a.Nil(updatedComment)
	a.NotNil(err)
	a.True(strings.Contains(err.Error(), "record not found"))

	updatedComment, err = UpdateCommentById(db, 9700, "updated comment", 1074241,
		nil, nil, nil)
	a.Nil(updatedComment)
	a.NotNil(err)
	a.True(strings.Contains(err.Error(), "record not found"))
}

func TestDeleteCommentById(t *testing.T) {
	baseSetup(t)
	_, err := DeleteCommentById(db, 9713, 1074247, contentWrapperMock, nil, nil,
		nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	var deletedComment []database.Comment
	if err := db.Where("id = 9713").Find(&deletedComment).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(0, len(deletedComment))

	_, err = DeleteCommentById(db, 9714, 1074247, contentWrapperMock, nil, nil,
		nil, nil)

	a.NotNil(err)
	a.True(strings.Contains(err.Error(), "no comments to delete"))

	_, err = DeleteCommentById(db, 9712, 1074248, contentWrapperMock, nil, nil,
		nil, nil)

	a.NotNil(err)
	a.True(strings.Contains(err.Error(), "delete operation not permitted"))

	_, err = DeleteCommentById(db, 9710, 1074248, contentWrapperMock, nil, nil,
		nil, nil)

	a.NotNil(err)
	a.True(strings.Contains(err.Error(), "delete operation not permitted 2"))
}

func TestCreateCommentOnProfile(t *testing.T) {
	baseSetup(t)
	comment, err := CreateCommentOnProfile(db, 11108, "test_create_comment_on_profile",
		null.NewInt(0, false), userWrapperMock, nil, 10, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var c1 database.Comment
	if err := db.Where("id = ?", comment.Id).First(&c1).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(int64(11108), c1.ProfileId.ValueOrZero())
	a.Equal("test_create_comment_on_profile", c1.Comment)
	a.Equal(false, c1.ParentId.Valid)

	comment2, err := CreateCommentOnProfile(db, 11108, "test_create_comment2_on_profile",
		null.IntFrom(comment.Id), userWrapperMock, nil, 10, nil, nil, nil)
	if err != nil {
		t.Fatal(err)
	}

	var c2 database.Comment
	if err := db.Where("id = ?", comment2.Id).First(&c2).Error; err != nil {
		t.Fatal(err)
	}

	a.Equal(int64(11108), c2.ProfileId.ValueOrZero())
	a.Equal("test_create_comment2_on_profile", c2.Comment)
	a.Equal(c1.Id, c2.ParentId.ValueOrZero())

	_, err = CreateCommentOnProfile(db, 11108, "test_create_comment2_on_profile",
		null.IntFrom(comment.Id), userWrapperMock, nil, 1, nil, nil, nil)
	a.NotEqual(nil, err)

}
