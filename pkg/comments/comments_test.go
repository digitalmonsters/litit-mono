package comments

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers"
	"github.com/digitalmonsters/go-common/wrappers/comment"
	"github.com/digitalmonsters/go-common/wrappers/content"
	user "github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
	"time"
)

var db *gorm.DB
var userWrapperMock user.IUserGoWrapper
var mockUserRecord = user.UserRecord{
	UserId:                     1074240,
	Avatar:                     null.String{},
	Username:                   "testusername",
	Firstname:                  "testFirstname",
	Lastname:                   "testLastName",
	Verified:                   false,
	EnableAgeRestrictedContent: false,
}

func TestMain(m *testing.M) {
	db = database.GetDb()

	userWrapperMock = &user.UserGoWrapperMock{
		GetUsersFn: func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user.UserRecord] {
			ch := make(chan wrappers.GenericResponseChan[map[int64]user.UserRecord], 2)
			go func() {
				defer func() {
					close(ch)
				}()

				if userIds[0] == mockUserRecord.UserId {
					ch <- wrappers.GenericResponseChan[map[int64]user.UserRecord]{
						Error: nil,
						Response: map[int64]user.UserRecord{
							mockUserRecord.UserId: mockUserRecord,
						},
					}
				}
			}()

			return ch
		},
		GetUserBlockFn: func(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[user.UserBlockData] {
			ch := make(chan wrappers.GenericResponseChan[user.UserBlockData], 2)
			go func() {
				defer func() {
					close(ch)
				}()

				if blockedTo == 1 || blockedBy == 2 {
					ch <- wrappers.GenericResponseChan[user.UserBlockData]{
						Error:    nil,
						Response: mockBlockRecord,
					}
				} else {
					ch <- wrappers.GenericResponseChan[user.UserBlockData]{
						Error:    nil,
						Response: mockNotBlockRecord,
					}
				}
			}()

			return ch
		},
	}

	contentWrapperMock = &content.ContentWrapperMock{
		GetInternalFn: func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]content.SimpleContent] {
			ch := make(chan wrappers.GenericResponseChan[map[int64]content.SimpleContent], 2)
			go func() {
				defer func() {
					close(ch)
				}()

				if contentIds[0] == mockContentRecord.Id {
					ch <- wrappers.GenericResponseChan[map[int64]content.SimpleContent]{
						Error: nil,
						Response: map[int64]content.SimpleContent{
							mockContentRecord.Id: mockContentRecord,
						},
					}
				}
			}()

			return ch
		},
	}
	os.Exit(m.Run())
}

func baseSetup(t *testing.T) {
	cfg := configs.GetConfig()

	if err := boilerplate_testing.FlushPostgresAllTables(cfg.Db, nil, nil); err != nil {
		t.Fatal(err)
	}
	if err := utils.PollutePostgresDatabase(db, "./test_data/seed.json"); err != nil {
		t.Fatal(err)
	}

}

func TestGetCommentsByContent(t *testing.T) {
	baseSetup(t)

	result, err := GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 1017738,
		After:      "",
		Count:      2,
		SortOrder:  "",
	}, 0, db, userWrapperMock, context.TODO(), ResourceTypeContent)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9693), result.Comments[0].Id)
	assert.Equal(t, int64(9699), result.Comments[1].Id)

	result, err = GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 1017738,
		After:      result.Paging.After,
		Count:      2,
		SortOrder:  "",
	}, 0, db, userWrapperMock, context.TODO(), ResourceTypeContent)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9705), result.Comments[0].Id)
	assert.Equal(t, int64(9691), result.Comments[1].Id)

	result, err = GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 1017738,
		After:      result.Paging.After,
		Count:      9999,
		SortOrder:  "",
	}, 0, db, userWrapperMock, context.TODO(), ResourceTypeContent)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 27, len(result.Comments))
}

func TestGetCommentById(t *testing.T) {
	baseSetup(t)

	userId := int64(10500)
	commentId := int64(9713)

	db.Create(&database.CommentVote{
		UserId:    userId,
		CommentId: commentId,
		VoteUp:    null.BoolFrom(true),
	})

	_, err := GetCommentById(db, commentId, userId, userWrapperMock, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

}

func TestGetCommentById_ChildrenComments(t *testing.T) {
	cfg := configs.GetConfig()

	if err := boilerplate_testing.FlushPostgresTables(cfg.Db,
		[]string{"public.comment"}, nil, nil); err != nil {
		t.Fatal(err)
	}
	userId := int64(10511)
	profileId := int64(9716)

	var parentComments []database.Comment

	for i := 0; i < 4; i++ {
		var minute = 2 * (i + 1)
		parentComments = append(parentComments, database.Comment{
			AuthorId:  userId,
			Comment:   fmt.Sprintf("test parent comments_%v", i+1),
			ProfileId: null.IntFrom(profileId),
			CreatedAt: time.Now().UTC().Add(time.Duration(minute) * time.Minute),
		})
	}
	if err := db.Create(&parentComments).Error; err != nil {
		t.Fatal(err)
	}

	var comments []database.Comment

	for i := 0; i < 3; i++ {
		var minute = -2 * (i + 1)
		comments = append(comments, database.Comment{
			AuthorId:  userId,
			Comment:   "test comments",
			ProfileId: null.IntFrom(profileId),
			ParentId:  null.IntFrom(parentComments[0].Id),
			CreatedAt: time.Now().UTC().Add(time.Duration(minute) * time.Minute),
		})
	}

	if err := db.Create(&comments).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := GetCommentById(db, parentComments[0].Id, userId, userWrapperMock, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	result, err := GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: parentComments[0].ProfileId.Int64,
		After:      resp.Cursor,
		Count:      2,
		SortOrder:  "newest",
	}, userId, db, userWrapperMock, nil, ResourceTypeProfile)

	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result.Comments), 2)
	assert.NotEqual(t, result.Paging.After, "")

	for _, c := range result.Comments {
		var isFound = false

		for _, p := range parentComments {
			if p.Id == c.Id {
				isFound = true
			}
		}
		assert.True(t, isFound)
	}
	resp, err = GetCommentById(db, comments[1].Id, userId, userWrapperMock, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	result, err = GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		After:      resp.Cursor,
		Count:      2,
		ResourceId: comments[1].ParentId.Int64,
		SortOrder:  "newest",
	}, userId, db, userWrapperMock, nil, ResourceTypeParentComment)

	if err != nil {
		t.Fatal(err)
	}
	assert.Equal(t, len(result.Comments), 1)
	assert.Equal(t, result.Paging.After, "")

	for _, c := range result.Comments {
		var isFound = false

		for _, p := range comments {
			if p.Id == c.Id {
				isFound = true
			}
		}
		assert.True(t, isFound)
	}
}

func TestGetCommentsByProfile(t *testing.T) {
	baseSetup(t)

	result, err := GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 11108,
		After:      "",
		Count:      2,
		SortOrder:  "",
	}, 0, db, userWrapperMock, context.TODO(), ResourceTypeProfile)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9713), result.Comments[0].Id)
	assert.Equal(t, int64(9712), result.Comments[1].Id)

	result, err = GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 11108,
		After:      result.Paging.After,
		Count:      2,
		SortOrder:  "",
	}, 0, db, userWrapperMock, context.TODO(), ResourceTypeProfile)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9711), result.Comments[0].Id)

}

func TestGetCommentsInfoById(t *testing.T) {
	baseSetup(t)

	result, err := GetCommentsInfoById(comment.GetCommentsInfoByIdRequest{CommentIds: []int64{9694, 9712}}, db)
	if err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	for commentId, info := range result {
		if commentId == 9694 {
			a.Equal(int64(1074241), info.ParentAuthorId.Int64)
		} else if commentId == 9712 {
			a.False(info.ParentAuthorId.Valid)
		} else {
			t.Fatal(errors.New("unexpected value"))
		}
	}
}
