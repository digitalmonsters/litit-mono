package comments

import (
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/digitalmonsters/go-common/wrappers/user_block"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"os"
	"testing"
)

var db *gorm.DB
var userWrapperMock user.IUserWrapper
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

	userWrapperMock = &user.UserWrapperMock{
		GetUsersFn: func(userIds []int64, apmTransaction *apm.Transaction, forceLog bool) chan user.GetUsersResponseChan {
			ch := make(chan user.GetUsersResponseChan, 2)
			go func() {
				defer func() {
					close(ch)
				}()

				if userIds[0] == mockUserRecord.UserId {
					ch <- user.GetUsersResponseChan{
						Error: nil,
						Items: map[int64]user.UserRecord{
							mockUserRecord.UserId: mockUserRecord,
						},
					}
				}
			}()

			return ch
		},
	}

	contentWrapperMock = &content.ContentWrapperMock{
		GetInternalFn: func(contentIds []int64, includeDeleted bool, apmTransaction *apm.Transaction, forceLog bool) chan content.ContentGetInternalResponseChan {
			ch := make(chan content.ContentGetInternalResponseChan, 2)
			go func() {
				defer func() {
					close(ch)
				}()

				if contentIds[0] == mockContentRecord.Id {
					ch <- content.ContentGetInternalResponseChan{
						Error: nil,
						Items: map[int64]content.SimpleContent{
							mockContentRecord.Id: mockContentRecord,
						},
					}
				}
			}()

			return ch
		},
	}

	blockWrapperMock = &user_block.UserBlockWrapperMock{
		GetUserBlockFn: func(blockedTo int64, blockedBy int64, apmTransaction *apm.Transaction, forceLog bool) chan user_block.GetUserBlockResponseChan {
			ch := make(chan user_block.GetUserBlockResponseChan, 2)
			go func() {
				defer func() {
					close(ch)
				}()

				if blockedTo == 1 || blockedBy == 2 {
					ch <- user_block.GetUserBlockResponseChan{
						Error: nil,
						Data:  mockBlockRecord,
					}
				} else {
					ch <- user_block.GetUserBlockResponseChan{
						Error: nil,
						Data:  mockNotBlockRecord,
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

	if err := boilerplate_testing.FlushPostgresTables(cfg.Db.ToBoilerplate(),
		[]string{"public.comment", "public.comment_vote", "public.content", "public.profile"}, nil, nil); err != nil {
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
	}, 0, db, userWrapperMock, nil, ResourceTypeContent)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9694), result.Comments[0].Id)
	assert.Equal(t, int64(9693), result.Comments[1].Id)

	result, err = GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 1017738,
		After:      result.Paging.After,
		Count:      2,
		SortOrder:  "",
	}, 0, db, userWrapperMock, nil, ResourceTypeContent)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9699), result.Comments[0].Id)
	assert.Equal(t, int64(9705), result.Comments[1].Id)

	result, err = GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 1017738,
		After:      result.Paging.After,
		Count:      9999,
		SortOrder:  "",
	}, 0, db, userWrapperMock, nil, ResourceTypeContent)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 30, len(result.Comments))
}

func TestGetCommentById(t *testing.T) {
	baseSetup(t)

	userId := int64(10500)
	commentId := int64(9694)

	db.Create(&database.CommentVote{
		UserId:    userId,
		CommentId: commentId,
		VoteUp:    null.BoolFrom(true),
	})

	data, err := GetCommentById(db, commentId, userId, userWrapperMock, nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, Comment{
		SimpleComment: SimpleComment{
			Id:           9694,
			AuthorId:     1074240,
			NumReplies:   55,
			NumDownvotes: 77,
			NumUpvotes:   66,
			CreatedAt:    data.CreatedAt,
			CreatedAtTs:  data.CreatedAtTs,
			Comment:      "Testing comment reply notification",
			ContentId:    null.IntFrom(1017738),
			MyVoteUp:     null.BoolFrom(true),
		},
		Author: Author{
			Id:        mockUserRecord.UserId,
			Username:  mockUserRecord.Username,
			Avatar:    mockUserRecord.Avatar,
			Firstname: mockUserRecord.Firstname,
			Lastname:  mockUserRecord.Lastname,
		},
		Content: content.SimpleContent{},
	}, *data)
}

func TestGetCommentsByProfile(t *testing.T) {
	baseSetup(t)

	result, err := GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: 11108,
		After:      "",
		Count:      2,
		SortOrder:  "",
	}, 0, db, userWrapperMock, nil, ResourceTypeProfile)

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
	}, 0, db, userWrapperMock, nil, ResourceTypeProfile)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9711), result.Comments[0].Id)

}
