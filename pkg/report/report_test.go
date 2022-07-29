package report

import (
	"context"
	"fmt"
	"github.com/digitalmonsters/comments/configs"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/boilerplate_testing"
	"github.com/digitalmonsters/go-common/frontend"
	"github.com/digitalmonsters/go-common/wrappers"
	content2 "github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user_go"
	"github.com/stretchr/testify/assert"
	"go.elastic.co/apm"
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
		[]string{"public.comment", "public.comment_vote", "public.content", "public.report", "public.profile"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	if err := utils.PollutePostgresDatabase(db, "../comments/test_data/seed.json"); err != nil {
		t.Fatal(err)
	}
}

func TestReportComment(t *testing.T) {
	baseSetup(t)
	report, err := ReportComment(9700, "spam", db, 1, "type")
	if err != nil {
		t.Fatal(err)
	}
	var dbReport *database.Report
	if err := db.Where("id = ?", report.Id).First(&dbReport).Error; err != nil {
		t.Fatal(err)
	}

	a := assert.New(t)
	a.Equal(int64(9700), report.CommentId)
	a.Equal(int64(1), report.ReporterId)
	a.Equal(int64(1017738), report.ContentId.ValueOrZero())
	a.Equal(int64(0), report.UserId.ValueOrZero())
	a.Equal("comment", report.ReportType)
	a.Equal("type", report.Type)
	a.Equal("spam", report.Detail)

	secondReport, err := ReportComment(9700, "spam", db, 1, "type")
	if err != nil {
		t.Fatal(err)
	}

	a.Equal(report.Id, secondReport.Id)

	reportOnProfile, err := ReportComment(9713, "violence", db, 1, "profile type")

	if err != nil {
		t.Fatal(err)
	}

	a.Equal(int64(9713), reportOnProfile.CommentId)
	a.Equal(int64(1), reportOnProfile.ReporterId)
	a.Equal(int64(0), reportOnProfile.ContentId.ValueOrZero())
	a.Equal(int64(11108), reportOnProfile.UserId.ValueOrZero())
	a.Equal("comment", reportOnProfile.ReportType)
	a.Equal("profile type", reportOnProfile.Type)
	a.Equal("violence", reportOnProfile.Detail)
}

func TestGetReportedUserProfileComments(t *testing.T) {
	cfg := configs.GetConfig()
	if err := boilerplate_testing.FlushPostgresTables(cfg.Db,
		[]string{"public.comment", "public.comment_vote", "public.content", "public.report", "public.profile"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	userGoWrapper := user_go.UserGoWrapperMock{
		GetUsersFn: func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord] {
			respMap := make(map[int64]user_go.UserRecord)
			ch := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord], 2)
			go func() {
				defer close(ch)
				for _, id := range userIds {
					respMap[id] = user_go.UserRecord{
						UserId:   id,
						Avatar:   null.StringFrom("avatar" + fmt.Sprint(id)),
						Username: "username" + fmt.Sprint(id),
					}
				}
				ch <- wrappers.GenericResponseChan[map[int64]user_go.UserRecord]{
					Response: respMap,
				}
			}()
			return ch
		},
	}

	okComments := make([]database.Comment, 0)

	for i := int64(1); i <= 10; i++ {
		comment := database.Comment{
			AuthorId:   i,
			Comment:    "test" + fmt.Sprint(i),
			ProfileId:  null.IntFrom(i + 10),
			NumReports: i,
		}
		okComments = append(okComments, comment)
	}

	if err := db.Create(&okComments).Error; err != nil {
		t.Fatal(err)
	}

	notOkComments := make([]database.Comment, 0)

	content := database.Content{}
	if err := db.Create(&content).Error; err != nil {
		t.Fatal(err)
	}

	for i := int64(100); i <= 150; i++ {
		comment1 := database.Comment{
			AuthorId:  i,
			Comment:   "test" + fmt.Sprint(i),
			ProfileId: null.IntFrom(i + 10),
		}
		notOkComments = append(notOkComments, comment1)

		comment2 := database.Comment{
			AuthorId:  i,
			Comment:   "test" + fmt.Sprint(i),
			ContentId: null.IntFrom(content.Id),
		}
		notOkComments = append(notOkComments, comment2)
	}

	if err := db.Create(&notOkComments).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := GetReportedUserProfileComments(GetReportedUserProfileCommentsRequest{
		Limit: 7,
	}, db, &userGoWrapper, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 7, len(resp.Items))
	assert.Equal(t, int64(10), resp.TotalCount)
	assert.Equal(t, int64(10), resp.Items[0].CommenterId)
	assert.Equal(t, "username10", resp.Items[0].CommenterUsername)
	assert.Equal(t, "avatar10", resp.Items[0].CommenterAvatar.ValueOrZero())
	assert.Equal(t, int64(20), resp.Items[0].UserId)
	assert.Equal(t, "username20", resp.Items[0].UserUsername)
	assert.Equal(t, "avatar20", resp.Items[0].UserAvatar.ValueOrZero())

	resp, err = GetReportedUserProfileComments(GetReportedUserProfileCommentsRequest{
		Limit:        10,
		CommenterIds: []int64{1, 2, 5},
	}, db, &userGoWrapper, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(resp.Items))
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, int64(5), resp.Items[0].CommenterId)
	assert.Equal(t, "username5", resp.Items[0].CommenterUsername)
	assert.Equal(t, "avatar5", resp.Items[0].CommenterAvatar.ValueOrZero())
	assert.Equal(t, int64(15), resp.Items[0].UserId)
	assert.Equal(t, "username15", resp.Items[0].UserUsername)
	assert.Equal(t, "avatar15", resp.Items[0].UserAvatar.ValueOrZero())

	resp, err = GetReportedUserProfileComments(GetReportedUserProfileCommentsRequest{
		Limit:   10,
		UserIds: []int64{11, 12, 15},
	}, db, &userGoWrapper, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(resp.Items))
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, int64(5), resp.Items[0].CommenterId)
	assert.Equal(t, "username5", resp.Items[0].CommenterUsername)
	assert.Equal(t, "avatar5", resp.Items[0].CommenterAvatar.ValueOrZero())
	assert.Equal(t, int64(15), resp.Items[0].UserId)
	assert.Equal(t, "username15", resp.Items[0].UserUsername)
	assert.Equal(t, "avatar15", resp.Items[0].UserAvatar.ValueOrZero())

}

func TestGetReportsForComment(t *testing.T) {
	cfg := configs.GetConfig()
	if err := boilerplate_testing.FlushPostgresTables(cfg.Db,
		[]string{"public.comment", "public.comment_vote", "public.content", "public.report", "public.profile"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	userGoWrapper := user_go.UserGoWrapperMock{
		GetUsersFn: func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord] {
			respMap := make(map[int64]user_go.UserRecord)
			ch := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord], 2)
			go func() {
				defer close(ch)
				for _, id := range userIds {
					respMap[id] = user_go.UserRecord{
						UserId:   id,
						Avatar:   null.StringFrom("avatar" + fmt.Sprint(id)),
						Username: "username" + fmt.Sprint(id),
					}
				}
				ch <- wrappers.GenericResponseChan[map[int64]user_go.UserRecord]{
					Response: respMap,
				}
			}()
			return ch
		},
	}

	comment1 := database.Comment{
		AuthorId:   1,
		Comment:    "test_comment1",
		ProfileId:  null.IntFrom(2),
		NumReports: 10,
	}

	if err := db.Create(&comment1).Error; err != nil {
		t.Fatal(err)
	}

	comment2 := database.Comment{
		AuthorId:   1,
		Comment:    "test_comment2",
		ProfileId:  null.IntFrom(2),
		NumReports: 10,
	}

	if err := db.Create(&comment2).Error; err != nil {
		t.Fatal(err)
	}

	for i := 1; i <= 10; i++ {
		iString := fmt.Sprint(i)
		report1 := database.Report{
			Type:       "type" + iString,
			ReportType: "report_type" + iString,
			ReporterId: int64(i),
			CommentId:  comment1.Id,
			Detail:     "detail" + iString,
		}
		if err := db.Create(&report1).Error; err != nil {
			t.Fatal(err)
		}
		report2 := database.Report{
			Type:       "type_extra" + iString,
			ReportType: "report_type_extra" + iString,
			ReporterId: int64(i),
			CommentId:  comment2.Id,
			Detail:     "detail_extra" + iString,
		}
		if err := db.Create(&report2).Error; err != nil {
			t.Fatal(err)
		}
	}

	resp, err := GetReportsForComment(GetReportsForCommentRequest{
		Limit:     7,
		CommentId: comment1.Id,
	}, db, &userGoWrapper, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 7, len(resp.Items))
	assert.Equal(t, int64(10), resp.TotalCount)
	assert.Equal(t, int64(10), resp.Items[0].ReporterId)
	assert.Equal(t, "username10", resp.Items[0].ReporterUsername)
	assert.Equal(t, "type10", resp.Items[0].Type)
	assert.Equal(t, "report_type10", resp.Items[0].ReportType)
	assert.Equal(t, "detail10", resp.Items[0].Detail)

	resp, err = GetReportsForComment(GetReportsForCommentRequest{
		Limit:     10,
		CommentId: comment1.Id,
		Sorting: []ReportsSorting{
			{
				Field:       ReportsSortFieldDate,
				IsAscending: true,
			},
		},
	}, db, &userGoWrapper, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 10, len(resp.Items))
	assert.Equal(t, int64(10), resp.TotalCount)
	assert.Equal(t, int64(1), resp.Items[0].ReporterId)
	assert.Equal(t, "username1", resp.Items[0].ReporterUsername)
	assert.Equal(t, "type1", resp.Items[0].Type)
	assert.Equal(t, "report_type1", resp.Items[0].ReportType)
	assert.Equal(t, "detail1", resp.Items[0].Detail)

	resp, err = GetReportsForComment(GetReportsForCommentRequest{
		Limit:         10,
		CommentId:     comment1.Id,
		ReportedByIds: []int64{5, 6, 1},
	}, db, &userGoWrapper, context.TODO())

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(resp.Items))
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, int64(6), resp.Items[0].ReporterId)
	assert.Equal(t, "username6", resp.Items[0].ReporterUsername)
	assert.Equal(t, "type6", resp.Items[0].Type)
	assert.Equal(t, "report_type6", resp.Items[0].ReportType)
	assert.Equal(t, "detail6", resp.Items[0].Detail)
}

func TestGetReportedVideoComments(t *testing.T) {
	cfg := configs.GetConfig()
	if err := boilerplate_testing.FlushPostgresTables(cfg.Db,
		[]string{"public.comment", "public.comment_vote", "public.content", "public.report", "public.profile"}, nil, nil); err != nil {
		t.Fatal(err)
	}

	userGoWrapper := user_go.UserGoWrapperMock{
		GetUsersFn: func(userIds []int64, ctx context.Context, forceLog bool) chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord] {
			respMap := make(map[int64]user_go.UserRecord)
			ch := make(chan wrappers.GenericResponseChan[map[int64]user_go.UserRecord], 2)
			go func() {
				defer close(ch)
				for _, id := range userIds {
					respMap[id] = user_go.UserRecord{
						UserId:   id,
						Avatar:   null.StringFrom("avatar" + fmt.Sprint(id)),
						Username: "username" + fmt.Sprint(id),
					}
				}
				ch <- wrappers.GenericResponseChan[map[int64]user_go.UserRecord]{
					Response: respMap,
				}
			}()
			return ch
		},
	}

	contentWrapper := content2.ContentWrapperMock{
		GetInternalAdminModelsFn: func(contentIds []int64, apmTx *apm.Transaction, forceLog bool) chan wrappers.GenericResponseChan[map[int64]frontend.ContentModel] {
			respMap := make(map[int64]frontend.ContentModel)
			ch := make(chan wrappers.GenericResponseChan[map[int64]frontend.ContentModel], 2)
			go func() {
				defer close(ch)
				for _, id := range contentIds {
					respMap[id] = frontend.ContentModel{
						Id:            id,
						AnimUrl:       "test_anim" + fmt.Sprint(id),
						CommentsCount: id,
						UserId:        id,
						VideoId:       "test_video_id" + fmt.Sprint(id),
						Thumbnail:     "test_thumbnail" + fmt.Sprint(id),
						User: frontend.VideoUserModel{
							Avatar:   "avatar" + fmt.Sprint(id),
							UserName: "username" + fmt.Sprint(id),
						},
					}
				}
				ch <- wrappers.GenericResponseChan[map[int64]frontend.ContentModel]{
					Response: respMap,
				}
			}()
			return ch
		},
	}

	okComments := make([]database.Comment, 0)

	for i := int64(1); i <= 10; i++ {
		content := database.Content{
			Id: i + 10,
		}
		if err := db.Create(&content).Error; err != nil {
			t.Fatal(err)
		}
		comment := database.Comment{
			AuthorId:   i,
			Comment:    "test" + fmt.Sprint(i),
			ContentId:  null.IntFrom(content.Id),
			NumReports: i,
		}
		okComments = append(okComments, comment)
	}

	if err := db.Create(&okComments).Error; err != nil {
		t.Fatal(err)
	}

	notOkComments := make([]database.Comment, 0)

	for i := int64(100); i <= 150; i++ {
		content := database.Content{
			Id: i + 10,
		}
		if err := db.Create(&content).Error; err != nil {
			t.Fatal(err)
		}
		comment1 := database.Comment{
			AuthorId:  i,
			Comment:   "test" + fmt.Sprint(i),
			ContentId: null.IntFrom(content.Id),
		}
		notOkComments = append(notOkComments, comment1)

		comment2 := database.Comment{
			AuthorId:   i,
			Comment:    "test" + fmt.Sprint(i),
			ProfileId:  null.IntFrom(content.Id),
			NumReports: i,
		}
		notOkComments = append(notOkComments, comment2)
	}

	if err := db.Create(&notOkComments).Error; err != nil {
		t.Fatal(err)
	}

	resp, err := GetReportedVideoComments(GetReportedVideoCommentsRequest{
		Limit: 7,
	}, db, &userGoWrapper, &contentWrapper, context.TODO(), nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 7, len(resp.Items))
	assert.Equal(t, int64(10), resp.TotalCount)
	assert.Equal(t, int64(10), resp.Items[0].CommenterId)
	assert.Equal(t, "username10", resp.Items[0].CommenterUsername)
	assert.Equal(t, "avatar10", resp.Items[0].CommenterAvatar.ValueOrZero())
	assert.Equal(t, int64(20), resp.Items[0].ContentId)
	assert.Equal(t, "username20", resp.Items[0].Content.User.UserName)
	assert.Equal(t, "avatar20", resp.Items[0].Content.User.Avatar)

	resp, err = GetReportedVideoComments(GetReportedVideoCommentsRequest{
		Limit:        10,
		CommenterIds: []int64{1, 2, 5},
	}, db, &userGoWrapper, &contentWrapper, context.TODO(), nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(resp.Items))
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, int64(5), resp.Items[0].CommenterId)
	assert.Equal(t, "username5", resp.Items[0].CommenterUsername)
	assert.Equal(t, "avatar5", resp.Items[0].CommenterAvatar.ValueOrZero())
	assert.Equal(t, int64(15), resp.Items[0].ContentId)
	assert.Equal(t, "username15", resp.Items[0].Content.User.UserName)
	assert.Equal(t, "avatar15", resp.Items[0].Content.User.Avatar)

	resp, err = GetReportedVideoComments(GetReportedVideoCommentsRequest{
		Limit:      10,
		ContentIds: []int64{11, 12, 15},
	}, db, &userGoWrapper, &contentWrapper, context.TODO(), nil)

	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 3, len(resp.Items))
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, int64(5), resp.Items[0].CommenterId)
	assert.Equal(t, "username5", resp.Items[0].CommenterUsername)
	assert.Equal(t, "avatar5", resp.Items[0].CommenterAvatar.ValueOrZero())
	assert.Equal(t, int64(15), resp.Items[0].ContentId)
	assert.Equal(t, "username15", resp.Items[0].Content.User.UserName)
	assert.Equal(t, "avatar15", resp.Items[0].Content.User.Avatar)
}
