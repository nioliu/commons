package interceptor

import (
	"context"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"log"
)

// VerifyJwtFailedFunc handle error if verify jwt failed, if this func return error,
//the grpc trace will be stopped and return
type VerifyJwtFailedFunc func(ctx context.Context, req interface{}, err error) error

// ExtractTokenFunc extract token from ctx or req
type ExtractTokenFunc func(ctx context.Context, req interface{}) (string, error)

var TokenNotValid = errors.New("token is not valid")

//type TokenNotValid struct {
//}
//
//func (t *TokenNotValid) Error() string {
//	return "token is not valid"
//}

type VerifyJwtImpl struct {
	SecretKey           []byte // secret key
	VerifyJwtFailedFunc VerifyJwtFailedFunc
	ExtractTokenFunc    ExtractTokenFunc
	Parser              *jwt.Parser
	Keyfunc             jwt.Keyfunc
}

func NewDefaultJwtImpl(secretKey []byte) *VerifyJwtImpl {
	verifyJwtFailedFunc := VerifyJwtFailedFunc(func(ctx context.Context, req interface{}, err error) error {
		return errors.New("verify jwt failed, err: " + err.Error())
	})

	extractTokenFunc := ExtractTokenFunc(func(ctx context.Context, req interface{}) (string, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok || md.Get("token") == nil || len(md.Get("token")) == 0 {
			return "", errors.New("can't find metadata from grpc context")
		} else {
			return md.Get("token")[0], nil
		}
	})

	parser := jwt.NewParser(jwt.WithValidMethods([]string{"HS256"}))

	keyFunc := jwt.Keyfunc(func(token *jwt.Token) (interface{}, error) {
		return secretKey, nil
	})

	return &VerifyJwtImpl{
		SecretKey:           secretKey,
		VerifyJwtFailedFunc: verifyJwtFailedFunc,
		ExtractTokenFunc:    extractTokenFunc,
		Parser:              parser,
		Keyfunc:             keyFunc,
	}
}

func NewVerifyJwtImpl(
	secretKey []byte,
	verifyJwtFailedFunc VerifyJwtFailedFunc,
	extractTokenFunc ExtractTokenFunc,
	parser *jwt.Parser,
	keyFunc jwt.Keyfunc,
) *VerifyJwtImpl {
	return &VerifyJwtImpl{
		SecretKey:           secretKey,
		VerifyJwtFailedFunc: verifyJwtFailedFunc,
		ExtractTokenFunc:    extractTokenFunc,
		Parser:              parser,
		Keyfunc:             keyFunc,
	}
}

func checkVerifyJwtImpl(obj *VerifyJwtImpl) error {
	if obj.SecretKey == nil {
		return errors.New("SecretKey cannot be empty")
	}
	if obj.VerifyJwtFailedFunc == nil {
		return errors.New("VerifyJwtFailedFunc cannot be empty")
	}
	if obj.ExtractTokenFunc == nil {
		return errors.New("ExtractTokenFunc cannot be empty")
	}
	if obj.Parser == nil {
		return errors.New("Parser cannot be empty")
	}
	if obj.Keyfunc == nil {
		return errors.New("Keyfunc cannot be empty")
	}
	return nil
}

func GetJwtVerifyInterceptor(ctx context.Context, verifyJwt *VerifyJwtImpl) grpc.UnaryServerInterceptor {
	if err := checkVerifyJwtImpl(verifyJwt); err != nil {
		log.Fatal(err)
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (resp interface{}, err error) {
		token, err := verifyJwt.ExtractTokenFunc(ctx, req)
		if err != nil {
			if err := verifyJwt.VerifyJwtFailedFunc(ctx, req, err); err != nil {
				return nil, err
			}
		}

		parsed, err := verifyJwt.Parser.Parse(token, verifyJwt.Keyfunc)
		if err != nil {
			if err := verifyJwt.VerifyJwtFailedFunc(ctx, req, err); err != nil {
				return nil, err
			}
		}

		if !parsed.Valid {
			if err := verifyJwt.VerifyJwtFailedFunc(ctx, req, TokenNotValid); err != nil {
				return nil, err
			}
		}

		return handler(ctx, req)
	}
}
