package publicapi

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"strings"
)

func GetCommendById(db *gorm.DB, commentId int64, currentUserId int64, userWrapper user.IUserWrapper,
	apmTransaction *apm.Transaction) (*Comment, error) {
	var comment database.Comment

	if err := db.Find(&comment).Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	resultComment := mapDbCommentToComment(comment)

	extenders := []chan error{
		extendWithAuthor(userWrapper, apmTransaction, &resultComment),
		extendWithLikedByMe(db, currentUserId, &resultComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	return &resultComment, nil
}

func DeleteCommentById(commentId int64, db *gorm.DB) (interface{}, error) {

}

func UpdateCommentById(commentId int64, comment string, db *gorm.DB) (interface{}, error) {

}

func GetRepliesByCommentId(commentId int64, db *gorm.DB) (interface{}, error) {

}

func VoteComment(commentId int64, voteUp bool, db *gorm.DB) (interface{}, error) {

}

func ReportComment(commentId int64, details string, db *gorm.DB) (interface{}, error) {

}

func GetCommentByTypeWithResourceId(request GetCommentsByTypeWithResourceRequest, currentUserId int64, db *gorm.DB,
	userWrapper user.IUserWrapper, apmTransaction *apm.Transaction) (*GetCommentsByTypeWithResourceResponse, error) {
	var comments []database.Comment

	query := db.Model(comments).Where("content_id = ?", request.ContentId)

	if request.ParentId > 0 {
		query = query.Where("parent_id = ?", request.ParentId)
	}

	var paginatorRules []paginator.Rule

	switch strings.ToLower(request.SortOrder) {
	case "newest":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "created_at",
			Order: paginator.DESC,
		})
	case "oldest":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "created_at",
			Order: paginator.ASC,
		})
	case "most_replied":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "num_replies",
			Order: paginator.DESC,
		})
	case "top_reactions":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "num_replies",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "num_upvotes",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "num_downvotes",
			Order: paginator.DESC,
		})
	case "least_popular":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "num_replies",
			Order: paginator.ASC,
		}, paginator.Rule{
			Key:   "num_upvotes",
			Order: paginator.ASC,
		}, paginator.Rule{
			Key:   "num_downvotes",
			Order: paginator.ASC,
		})
	default:
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "num_replies",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "num_upvotes",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "num_downvotes",
			Order: paginator.DESC,
		})
	}

	p := paginator.New(
		&paginator.Config{
			Rules: paginatorRules,
			Limit: request.Count,
		},
	)

	if len(request.After) > 0 {
		p.SetAfterCursor(request.After)
	}

	result, cursor, err := p.Paginate(query, &comments)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if result.Error != nil {
		return nil, errors.WithStack(err)
	}

	var resultComments []*Comment

	for _, comment := range comments {
		item := mapDbCommentToComment(comment)
		resultComments = append(resultComments, &item)
	}

	if len(comments) > 0 {
		extenders := []chan error{
			extendWithAuthor(userWrapper, apmTransaction, resultComments...),
			extendWithLikedByMe(db, currentUserId, resultComments...),
		}

		for _, e := range extenders {
			if err = <-e; err != nil {
				apm_helper.CaptureApmError(err, apmTransaction)
			}
		}
	}

	pagingResult := CursorPaging{}

	if cursor.After != nil {
		pagingResult.HasNext = true
		pagingResult.Next = *cursor.After
	}

	finalResponse := GetCommentsByTypeWithResourceResponse{
		Paging: pagingResult,
	}

	for _, c := range resultComments {
		finalResponse.Comments = append(finalResponse.Comments, *c)
	}

	return &finalResponse, nil
}

func SendComment(commentType string, resourceId int64, db *gorm.DB) (interface{}, error) {

}
