package api

import "errors"

// ErrInvalidMsgType is returned by an Option in case the Option is supplied
// an unexpected proto.Message
var ErrInvalidMsgType = errors.New("invalid message type")

// ErrInvalidValue is returned by a Option in case the Option is supplied
// an unexpected value.
var ErrInvalidValue = errors.New("invalid value")
