package cerror

import "errors"

func UnwrapError(err error) error {
	if err == nil {
		return nil
	}
	// cast to WithStack
	var e *WithStack
	if errors.As(err, &e) {
		return e.error
	}
	return err
}

func ErrorWithStack(err error) error {
	if err == nil {
		return nil
	}
	return &WithStack{
		err,
		callers(),
	}
}

type WithStack struct {
	error
	*stack
}

func (w *WithStack) Cause() error { return w.error }

func (w *WithStack) Unwrap() error { return w.error }

func ErrorWithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	return &WithMessage{
		cause: err,
		msg:   message,
	}
}

type WithMessage struct {
	cause error
	msg   string
}

func (w *WithMessage) Error() string { return w.msg + ": " + w.cause.Error() }

func (w *WithMessage) Cause() error { return w.cause }

func (w *WithMessage) Unwrap() error { return w.cause }

func WrapWithMessage(err error, message string) error {
	if err == nil {
		return nil
	}
	err = &WithMessage{
		cause: err,
		msg:   message,
	}
	return &WithStack{
		err,
		callers(),
	}
}
