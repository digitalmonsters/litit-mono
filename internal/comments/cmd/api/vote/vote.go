package vote

import (
	"encoding/json"
	"net/http"

	"github.com/digitalmonsters/comments/cmd/api/comments/notifiers/comment"
	vote2 "github.com/digitalmonsters/comments/cmd/api/vote/notifiers/vote"
	"github.com/digitalmonsters/comments/pkg/vote"
	"github.com/digitalmonsters/comments/utils"
	"github.com/digitalmonsters/go-common/error_codes"
	"github.com/digitalmonsters/go-common/router"
	"github.com/digitalmonsters/go-common/wrappers/content"
	"github.com/pkg/errors"
	"gorm.io/gorm"
)

func Init(httpRouter *router.HttpRouter, db *gorm.DB, commentNotifier *comment.Notifier,
	voteNotifier *vote2.Notifier, contentWrapper content.IContentWrapper) error {

	if err := httpRouter.RegisterRestCmd(router.NewRestCommand(func(request []byte,
		executionData router.MethodExecutionData) (interface{}, *error_codes.ErrorWithCode) {
		commentId := utils.ExtractInt64(executionData.GetUserValue, "comment_id", 0, 0)
		if executionData.IsGuest {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("registration required"), error_codes.RegistrationRequiredError)
		}
		if commentId <= 0 {
			return nil, error_codes.NewErrorWithCodeRef(errors.New("invalid comment_id"), error_codes.GenericValidationError)
		}

		var reportRequest voteRequest

		if err := json.Unmarshal(request, &reportRequest); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericMappingError)
		}

		if _, err := vote.VoteComment(db.WithContext(executionData.Context), commentId,
			reportRequest.VoteUp, executionData.UserId, commentNotifier, voteNotifier,
			executionData.Context, contentWrapper); err != nil {
			return nil, error_codes.NewErrorWithCodeRef(err, error_codes.GenericServerError)
		} else {
			return successResponse{
				Success: true,
			}, nil
		}
	}, "/{comment_id}/vote", http.MethodPost).RequireIdentityValidation().Build()); err != nil {
		return err
	}

	return nil
}
