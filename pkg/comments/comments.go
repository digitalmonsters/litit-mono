package comments

import (
	"fmt"
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/digitalmonsters/go-common/wrappers/user_block"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gorm.io/gorm"
	"strings"
)

func GetCommentsByResourceId(request GetCommentsByTypeWithResourceRequest, currentUserId int64, db *gorm.DB,
	userWrapper user.IUserWrapper, apmTransaction *apm.Transaction, resourceType ResourceType) (*GetCommentsByTypeWithResourceResponse, error) {
	var comments []database.Comment

	if request.ParentId == 0 && request.ResourceId == 0 {
		return nil, errors.New("parent id or content id should be set")
	}

	query := db.Model(comments)

	if request.ResourceId > 0 {
		switch resourceType {
		case ResourceTypeContent:
			query = query.Where("content_id = ?", request.ResourceId)

		case ResourceTypeProfile:
			query = query.Where("profile_id = ?", request.ResourceId)
		case ResourceTypeParentComment:
			query = query.Where("parent_id = ?", request.ParentId)
		}
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
		},
			paginator.Rule{
				Key:   "Id",
				Order: paginator.ASC,
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
		}, paginator.Rule{
			Key:   "Id",
			Order: paginator.ASC,
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
		},
			paginator.Rule{
				Key:   "Id",
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
		},
			paginator.Rule{
				Key:   "Id",
				Order: paginator.ASC,
			},
		)
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

	if len(request.Before) > 0 {
		p.SetBeforeCursor(request.Before)
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
		pagingResult.After = *cursor.After
	}

	if cursor.Before != nil {
		pagingResult.Before = *cursor.Before
	}

	finalResponse := GetCommentsByTypeWithResourceResponse{
		Paging: pagingResult,
	}

	for _, c := range resultComments {
		finalResponse.Comments = append(finalResponse.Comments, *c)
	}

	return &finalResponse, nil
}

func GetCommentById(db *gorm.DB, commentId int64, currentUserId int64, userWrapper user.IUserWrapper,
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

func isBlocked(userBlockWrapper user_block.IUserBlockWrapper, apmTransaction *apm.Transaction,
	blockedTo int64, blockedBy int64) (*user_block.BlockedUserType, error) {
	responseData := <-userBlockWrapper.GetUserBlock(blockedTo, blockedBy, apmTransaction, false)

	if responseData.Error != nil {
		return nil, errors.New(fmt.Sprintf("invalid response from user block service [%v]", responseData.Error.Message))
	}

	return responseData.Data.Type, nil
}

func updateUserStatsComments(tx *gorm.DB, authorId int64, contentId int64) error {
	if err := tx.Exec("insert into user_stats_action(id, comments) values (?, 1) on conflict (id) do update set comments = excluded.comments + 1", authorId).Error; err != nil {
		return err
	}

	if err := tx.Exec("insert into user_stats_content(id, comments) values (?, 1) on conflict (id) do update set comments = excluded.comments + 1", contentId).Error; err != nil {
		return err
	}

	return nil
}

func updateContentCommentsCounter(tx *gorm.DB, contentId int64, isIncrement bool) error {
	var incrementStm string

	if isIncrement {
		incrementStm = "comments_count + 1"
	} else {
		incrementStm = "comments_count - 1"
	}

	if err := tx.Model(database.Content{}).Where("id = ?", contentId).
		Update("comments_count", gorm.Expr(incrementStm)).Error; err != nil {
		return err
	}

	return nil
}
