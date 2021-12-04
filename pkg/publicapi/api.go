package publicapi

import (
	"github.com/digitalmonsters/comments/pkg/database"
	"github.com/digitalmonsters/go-common/apm_helper"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/digitalmonsters/go-common/wrappers/user"
	"github.com/pilagod/gorm-cursor-paginator/v2/paginator"
	"github.com/pkg/errors"
	"go.elastic.co/apm"
	"gopkg.in/guregu/null.v4"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func DeleteCommentById(db *gorm.DB, commentId int64, currentUserId int64, contentWrapper content.IContentWrapper,
	apmTransaction *apm.Transaction) (interface{}, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var comment database.Comment

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	mappedComment := mapDbCommentForDeleteToCommentForDelete(comment)

	extenders := []chan error{
		extendWithContentId(contentWrapper, apmTransaction, &mappedComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	if mappedComment.AuthorId != currentUserId && mappedComment.Content.AuthorId != currentUserId {
		return nil, errors.WithStack(errors.New("not allowed"))
	}

	if comment.ParentId.ValueOrZero() > 0 {
		if err := tx.Model(database.Comment{}).Where("id = ?", comment.ParentId.ValueOrZero()).
			Update("num_replies", gorm.Expr("num_replies - 1")).Error; err != nil {
			return nil, err
		}
	}

	if err := tx.Delete(&comment, commentId).Error; err != nil {
		return nil, err
	}

	return nil, tx.Commit().Error
}

func UpdateCommentById(db *gorm.DB, commentId int64, updatedComment string, currentUserId int64) (*database.Comment, error) {
	var comment database.Comment

	tx := db.Begin()
	defer tx.Rollback()

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("author_id = ?", currentUserId).
		Take(&comment, commentId).Error; err != nil {
		return nil, err
	}

	if err := tx.Model(&comment).Update("comment", updatedComment).Error; err != nil {
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		return nil, errors.WithStack(err)
	}

	return &comment, nil
}

func GetRepliesByCommentId(commentId int64, db *gorm.DB, transaction *apm.Transaction, count int64, after string) (interface{}, error) {

}

func VoteComment(db *gorm.DB, commentId int64, voteUp null.Bool, currentUserId int64) (interface{}, error) {
	tx := db.Begin()
	defer tx.Rollback()

	var comment database.Comment

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Take(&comment, commentId).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	previousVoteValue := null.NewBool(false, false)
	var previousVote database.CommentVote

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Find(&previousVote, commentId, currentUserId).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if previousVote.UserId > 0 { // if 0, then its new vote
		previousVoteValue = null.BoolFrom(previousVote.VoteUp)
	}

	previousVote.CommentId = commentId
	previousVote.UserId = currentUserId

	if previousVote.VoteUp == voteUp.ValueOrZero() { // nothing to do here, as status in db is already valid
		return nil, nil
	}

	previousVote.VoteUp = voteUp.ValueOrZero()

	if previousVote.VoteUp {
		if err := tx.Model(&comment).Update("num_upvotes", gorm.Expr("num_upvotes + 1")).Error; err != nil {
			return nil, errors.WithStack(err)
		}
	} else {
		if err := tx.Model(&comment).Update("num_downvotes", gorm.Expr("num_upvotes + 1")).Error; err != nil {
			return nil, errors.WithStack(err)
		}
	}

	if previousVoteValue.Valid {
		if previousVoteValue.Bool {
			if err := tx.Model(&comment).Update("num_upvotes", gorm.Expr("num_upvotes - 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		} else {
			if err := tx.Model(&comment).Update("num_downvotes", gorm.Expr("num_upvotes - 1")).Error; err != nil {
				return nil, errors.WithStack(err)
			}
		}
	}

	if err := tx.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Save(previousVote).Error; err != nil {
		return nil, errors.WithStack(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return nil, nil
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

func SendContentComment(db *gorm.DB, resourceId int64, commentStr string, parentId null.Int, contentWrapper content.IContentWrapper,
	apmTransaction *apm.Transaction, currentUserId int64) (*SendCommentResponse, error) {
	var parentComment database.Comment

	if !parentId.IsZero() {
		if err := db.Take(&parentComment, parentId.ValueOrZero()).Error; err != nil {
			return nil, err
		}
	}

	mappedComment := mapDbCommentForSendToCommentForSend(parentComment)

	extenders := []chan error{
		extendWithContentForSend(contentWrapper, apmTransaction, &mappedComment),
	}

	for _, e := range extenders {
		if err := <-e; err != nil {
			apm_helper.CaptureApmError(err, apmTransaction)
		}
	}

	blockedUserType, err := isBlocked(currentUserId, mappedComment.AuthorId)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	if blockedUserType != nil {
		return nil, errors.WithStack(errors.New(string(*blockedUserType)))
	}

	tx := db.Begin()
	defer tx.Rollback()
	var comment database.Comment

	comment.ContentId = resourceId
	comment.Comment = commentStr
	comment.AuthorId = currentUserId
	comment.ParentId = parentId

	if err := tx.Omit("created_at").Create(&comment).Error; err != nil {
		return nil, err
	}

	if !parentId.IsZero() {
		if err := tx.Model(&parentComment).Update("num_replies", parentComment.NumReplies+1).Error; err != nil {
			return nil, err
		}
	}

	if err := updateUserStatsComments(request.AuthorId); err != nil {
		return nil, err
	}

	// TODO: send notify 'comment'

	if err := tx.Commit().Error; err != nil {
		return nil, err
	}

	return &SendCommentResponse{
		Id:        comment.Id,
		Comment:   comment.Comment,
		AuthorId:  comment.AuthorId,
		ContentId: comment.ContentId,
	}, nil
}

func isBlocked(blockBy int64, blockTo int64) (*BlockedUserType, error) {
	// TODO: need implementation
	return nil, nil
}

func updateUserStatsComments(authorId int64) error {
	// TODO: need implementation
	return nil
}
