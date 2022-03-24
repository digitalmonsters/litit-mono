package comments

import (
	"fmt"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/content_comments_counter"
	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/user_comments_counter"
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

	if request.ResourceId == 0 {
		return nil, errors.New("resource id should be set")
	}

	query := db.Model(comments)

	if request.ResourceId > 0 {
		switch resourceType {
		case ResourceTypeContent:
			query = query.Where("content_id = ?", request.ResourceId).Where("parent_id is null")

		case ResourceTypeProfile:
			query = query.Where("profile_id = ?", request.ResourceId).Where("parent_id is null")
		case ResourceTypeParentComment:
			query = query.Where("parent_id = ?", request.ResourceId)
		}
	}
	var paginatorRules []paginator.Rule

	switch strings.ToLower(request.SortOrder) {
	case "newest":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "CreatedAt",
			Order: paginator.DESC,
		},
			paginator.Rule{
				Key:   "Id",
				Order: paginator.ASC,
			})
	case "oldest":
		paginatorRules = append(paginatorRules, paginator.Rule{
			Key:   "CreatedAt",
			Order: paginator.ASC,
		},
			paginator.Rule{
				Key:   "Id",
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
		return nil, errors.WithStack(result.Error)
	}

	var resultComments []*Comment

	for _, comment := range comments {
		item := MapDbCommentToComment(comment)
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
	apmTransaction *apm.Transaction) (*CommentWithCursor, error) {
	var comment database.Comment

	if err := db.Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	resourceType := ResourceTypeContent
	resourceTypeString := "content_id"
	resourceId := comment.ContentId.ValueOrZero()

	if comment.ProfileId.Valid {
		resourceType = ResourceTypeProfile
		resourceId = comment.ProfileId.ValueOrZero()
		resourceTypeString = "profile_id"
	}

	if comment.ParentId.Valid {
		resourceType = ResourceTypeParentComment
		resourceId = comment.ParentId.ValueOrZero()
		resourceTypeString = "parent_id"
	}

	var index int64

	indexQuery := fmt.Sprintf("select count(*) from comment where %v = ?", resourceTypeString)
	if resourceType != ResourceTypeParentComment {
		indexQuery += " and parent_id is null"
	}
	indexQuery += " and id >= ?;"

	if err := db.Raw(indexQuery, resourceId, commentId).Scan(&index).Error; err != nil {
		return nil, err
	}

	commentsResp, err := GetCommentsByResourceId(GetCommentsByTypeWithResourceRequest{
		ResourceId: resourceId,
		Count:      index,
		SortOrder:  "newest",
	}, currentUserId, db, userWrapper, nil, resourceType)

	if err != nil {
		return nil, err
	}
	resultComment := MapDbCommentToComment(comment)

	extenders := []chan error{
		extendWithAuthor(userWrapper, apmTransaction, &resultComment),
		extendWithLikedByMe(db, currentUserId, &resultComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	cursor := ""

	if commentsResp != nil {
		cursor = commentsResp.Paging.After
	}

	return &CommentWithCursor{
		SimpleComment: resultComment.SimpleComment,
		Author:        resultComment.Author,
		Content:       resultComment.Content,
		Cursor:        cursor,
		ParentId:      comment.ParentId,
	}, nil
}

func isBlocked(userBlockWrapper user_block.IUserBlockWrapper, apmTransaction *apm.Transaction,
	blockedTo int64, blockedBy int64) (*user_block.BlockedUserType, error) {
	responseData := <-userBlockWrapper.GetUserBlock(blockedTo, blockedBy, apmTransaction, false)

	if responseData.Error != nil {
		return nil, errors.New(fmt.Sprintf("invalid response from user block service [%v]", responseData.Error.Message))
	}

	return responseData.Data.Type, nil
}

func updateUserStatsComments(tx *gorm.DB, authorId int64, contentId int64, isContentComment bool,
	userCommentsNotifier *user_comments_counter.Notifier, contentCommentsNotifier *content_comments_counter.Notifier,
	decrease bool) error {
	var userStatsActionComments int64
	var userStatsContentComments int64

	increaseValue := 1

	if decrease {
		increaseValue = -1
	}

	if err := tx.Raw("insert into user_stats_action(id, comments) values (?, 1) on conflict (id) "+
		"do update set comments = user_stats_action.comments + ? returning comments", authorId, increaseValue).
		Scan(&userStatsActionComments).Error; err != nil {
		return err
	}

	if isContentComment {
		if err := tx.Raw("insert into user_stats_content(id, comments) values (?, 1) on conflict (id) "+
			"do update set comments = user_stats_content.comments + ? returning comments", contentId, increaseValue).
			Scan(&userStatsContentComments).Error; err != nil {
			return err
		}
	}

	if userCommentsNotifier != nil {
		userCommentsNotifier.Enqueue(authorId, userStatsActionComments)
	}

	if isContentComment && contentCommentsNotifier != nil {
		contentCommentsNotifier.Enqueue(contentId, userStatsContentComments)
	}

	return nil
}
