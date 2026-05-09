package workcell

import "errors"

var (
	ErrInvalidProfile = errors.New("invalid_profile")
	ErrInvalidCommand = errors.New("invalid_command")
)

func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrInvalidProfile):
		return "invalid_profile"
	case errors.Is(err, ErrInvalidCommand):
		return "invalid_command"
	default:
		return "workcell_error"
	}
}
