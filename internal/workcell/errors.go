package workcell

import "errors"

var (
	ErrInvalidProfile    = errors.New("invalid_profile")
	ErrInvalidCommand    = errors.New("invalid_command")
	ErrInvalidValidation = errors.New("invalid_validation_request")
)

func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrInvalidProfile):
		return "invalid_profile"
	case errors.Is(err, ErrInvalidCommand):
		return "invalid_command"
	case errors.Is(err, ErrInvalidValidation):
		return "invalid_validation_request"
	default:
		return "workcell_error"
	}
}
