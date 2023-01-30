package errs

import (
	"encoding/json"
	"github.com/go-redis/redis/v8"
)

// Standard:
// Inner system error 1000 - 1999
// Outer error  2000 - 2999

// ErrRsp common err rsp
type ErrRsp struct {
	Code        int    `json:"code"`
	Description string `json:"description"`
}

func (e *ErrRsp) Error() string {
	e2 := &ErrRsp{}
	*e2 = *e
	if e2 == nil {
		e2 = GetDefaultErrRsp()
	}
	marshal, _ := json.Marshal(e2)
	return string(marshal)
}

func (e *ErrRsp) WithDescription(dsp string) *ErrRsp {
	e.Description = dsp
	return e
}

func (e *ErrRsp) Err() *ErrRsp {
	return e
}

// GetDefaultErrRsp default
func GetDefaultErrRsp() *ErrRsp {
	return &ErrRsp{Code: 0, Description: "Unknown"}
}

func GetDatabaseError(description string) *ErrRsp {
	return &ErrRsp{Code: 1000, Description: description}
}

func NewError(code int, description string) *ErrRsp {
	return &ErrRsp{Code: code, Description: description}
}

// GetInnerSystemStandardError get common error response.
func GetInnerSystemStandardError(err error) (errRsp *ErrRsp) {
	switch err {
	case redis.Nil:
		errRsp = &ErrRsp{
			Code:        1000,
			Description: "not found in storage",
		}
	}

	return GetDefaultErrRsp()
}
