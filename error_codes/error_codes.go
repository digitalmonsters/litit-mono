package error_codes

import (
	"fmt"
	"github.com/pkg/errors"
)

type ErrorCode int

const (
	None                        ErrorCode = 200
	GenericValidationError      ErrorCode = 400
	GenericServerError          ErrorCode = 500
	Timeout                     ErrorCode = 408
	MissingJwtToken             ErrorCode = 401
	ExpiredJwtToken             ErrorCode = 401
	InvalidJwtToken             ErrorCode = 401
	InvalidMethodPermission     ErrorCode = 401
	Forbidden                   ErrorCode = 403
	NotAllowed                  ErrorCode = 405
	GenericMappingError         ErrorCode = -32700
	GenericDuplicateError       ErrorCode = 409
	GenericNotFoundError        ErrorCode = 404
	CommandNotFoundError        ErrorCode = -32601
	GenericTimeoutError         ErrorCode = 502
	GenericPanicError           ErrorCode = -32603
	UserHasNoPermissionToMethod ErrorCode = 601
	KYCRequiredError                      = ErrorCode(602)
	RegistrationRequiredError             = ErrorCode(603)
)

const (
	AdminAuthAdminNotFoundError ErrorCode = iota + 701
	AdminAuthInvalidPasswordError
	AdminAuthSendSmsCodeError
	AdminAuthInvalidMobileAuthCodeError
	AdminAuthInvalidMFATokenError
	AdminAuthJWTError
	AdminAuthNotUniqueEmailError
	AdminAuthNotUniqueMobilePhoneError
	AdminAuthPasswordsDoNotMatchError
	AdminAuthInactiveError
)

const (
	TokenomicsNotEnoughBalance        ErrorCode = 1500
	TokenomicsReceivingUserWithoutKyc ErrorCode = 1501
)

var TokenomicsNotEnoughBalanceError = errors.New("user doesn't have enough money to execute operation")
var TokenomicsErrorCannotProceedWithoutKyc = errors.New("You cannot proceed without KYC")
var TokenomicsErrorReceivingUserWithoutKyc = errors.New("You cannot tip user without KYC")

type ErrorWithCode struct {
	error error
	code  ErrorCode
}

func NewErrorWithCode(err error, code ErrorCode) ErrorWithCode {
	if err == nil {
		err = errors.New("that should not be happen, error is NIL !!!")
	}

	return ErrorWithCode{
		error: err,
		code:  code,
	}
}

func NewErrorWithCodeRef(err error, code ErrorCode) *ErrorWithCode {
	val := NewErrorWithCode(err, code)

	return &val
}

func (e *ErrorWithCode) GetCode() ErrorCode {
	return e.code
}
func (e *ErrorWithCode) GetMessage() string {
	if e.error == nil {
		return ""
	}

	return fmt.Sprintf("%v", e.error.Error())
}

func (e *ErrorWithCode) GetStack() string {
	if e.error == nil {
		return ""
	}

	return fmt.Sprintf("%+v", e.error)
}

func (e *ErrorWithCode) GetError() error {
	return e.error
}

type SimpleException struct {
	Message        string           `json:"message"`
	StackTrace     string           `json:"stack_trace"`
	InnerException *SimpleException `json:"inner_exception"`
	Extra          interface{}      `json:"extra"`
}

func (e *ErrorWithCode) ToSimpleException() SimpleException {
	return SimpleException{
		Message:        e.GetMessage(),
		StackTrace:     e.GetStack(),
		InnerException: nil,
		Extra:          nil,
	}
}
