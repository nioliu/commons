package errs

import (
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

// GetDefaultErrRsp default
func GetDefaultErrRsp() ErrRsp {
	return ErrRsp{Code: 0, Description: "Unknown"}
}

// GetInnerSystemStandardError get common error response.
func GetInnerSystemStandardError(err error) (errRsp ErrRsp) {
	switch err {
	case redis.Nil:
		errRsp = ErrRsp{
			Code:        1000,
			Description: "not found in storage",
		}
	}

	return GetDefaultErrRsp()
}
