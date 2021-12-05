package comments

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

func GetCommentsByContent(request GetCommentsByTypeWithResourceRequest, currentUserId int64, db *gorm.DB,
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
			Key:   "CreatedAt",
			Order: paginator.DESC,
		})
	case "oldest":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "CreatedAt",
			Order: paginator.ASC,
		})
	case "most_replied":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "NumReplies",
			Order: paginator.DESC,
		})
	case "top_reactions":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "NumReplies",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "NumUpvotes",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "NumDownvotes",
			Order: paginator.DESC,
		})
	case "least_popular":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "NumReplies",
			Order: paginator.ASC,
		}, paginator.Rule{
			Key:   "NumUpvotes",
			Order: paginator.ASC,
		}, paginator.Rule{
			Key:   "NumDownvotes",
			Order: paginator.ASC,
		})
	default:
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "NumReplies",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "NumUpvotes",
			Order: paginator.DESC,
		}, paginator.Rule{
			Key:   "NumDownvotes",
			Order: paginator.DESC,
		})
	}

	p := paginator.New(
		&paginator.Config{
			Rules: paginatorRules,
			Limit: int(request.Count),
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

func GetCommendById(db *gorm.DB, commentId int64, currentUserId int64, userWrapper user.IUserWrapper,
	apmTransaction *apm.Transaction) (*Comment, error) {
	var comment database.Comment

	if err := db.Take(&comment, commentId).Error; err != nil {
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

func GetRepliesByCommentId(commentId int64, db *gorm.DB, transaction *apm.Transaction, count int64, after string) (interface{}, error) {

}

func isBlocked(blockBy int64, blockTo int64) (*BlockedUserType, error) {
	// TODO: need implementation
	return nil, nil
}

func updateUserStatsComments(authorId int64) error {
	// TODO: need implementation
	return nil
}
