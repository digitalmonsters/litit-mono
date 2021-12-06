package comments

import (
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/wrappers/user"
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

	os.Exit(m.Run())
}

func baseSetup(t *testing.T) {
	cfg := configs.GetConfig()

	if err := boilerplate_testing.FlushPostgresTables(cfg.Db.ToBoilerplate(),
		[]string{"public.comment", "public.comment_vote", "public.content"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := utils.PollutePostgresDatabase(db, "./test_data/seed.json"); err != nil {
		t.Fatal(err)
	}
}

func TestGetCommentsByContent(t *testing.T) {
	baseSetup(t)

	result, err := GetCommentsByContent(GetCommentsByTypeWithResourceRequest{
		ContentId: 1017738,
		ParentId:  0,
		After:     "",
		Count:     2,
		SortOrder: "",
	}, 0, db, userWrapperMock, nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.True(t, len(result.Paging.Next) > 0)
	assert.Equal(t, true, result.Paging.HasNext)

	assert.Equal(t, int64(9699), result.Comments[0].Id)
	assert.Equal(t, int64(9693), result.Comments[1].Id)

	result, err = GetCommentsByContent(GetCommentsByTypeWithResourceRequest{
		ContentId: 1017738,
		ParentId:  0,
		After:     result.Paging.Next,
		Count:     2,
		SortOrder: "",
	}, 0, db, userWrapperMock, nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, int64(9691), result.Comments[0].Id)
	assert.Equal(t, int64(9707), result.Comments[1].Id)

	assert.Equal(t, true, result.Paging.HasNext)

	result, err = GetCommentsByContent(GetCommentsByTypeWithResourceRequest{
		ContentId: 1017738,
		ParentId:  0,
		After:     result.Paging.Next,
		Count:     9999,
		SortOrder: "",
	}, 0, db, userWrapperMock, nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 28, len(result.Comments))
	assert.Equal(t, false, result.Paging.HasNext)
	assert.Equal(t, "", result.Paging.Next)
}

func TestGetCommendById(t *testing.T) {
	baseSetup(t)

	data, err := GetCommendById(db, 9694, 0, userWrapperMock, nil)

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
			Comment:      "Testing comment reply notification",
			ContentId:    1017738,
		},
		Author: Author{
			Id:        mockUserRecord.UserId,
			Username:  mockUserRecord.Username,
			Avatar:    mockUserRecord.Avatar,
			Firstname: mockUserRecord.Firstname,
			Lastname:  mockUserRecord.Lastname,
		},
		Content: SimpleContent{},
	}, *data)
}
