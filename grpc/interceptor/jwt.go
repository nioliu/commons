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
type VerifyJwtFailedFunc func(ctx context.Context, req interface{}, err error) (bool, error)

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
	KeyFunc             jwt.Keyfunc
}

type Option func(impl *VerifyJwtImpl)

func apply(impl *VerifyJwtImpl, option ...Option) {
	for _, o := range option {
		o(impl)
	}
}

func WithSecretKey(secretKey []byte) Option {
	if secretKey == nil {
		log.Fatal("SecretKey cannot be empty")
	}
	return func(impl *VerifyJwtImpl) {
		impl.SecretKey = secretKey
	}
}

func WithVerifyJwtFailedFunc(verifyJwtFailedFunc VerifyJwtFailedFunc) Option {
	if verifyJwtFailedFunc == nil {
		log.Fatal("VerifyJwtFailedFunc cannot be empty")
	}
	return func(impl *VerifyJwtImpl) {
		impl.VerifyJwtFailedFunc = verifyJwtFailedFunc
	}
}

func WithExtractTokenFunc(extractTokenFunc ExtractTokenFunc) Option {
	if extractTokenFunc == nil {
		log.Fatal("ExtractTokenFunc cannot be empty")
	}
	return func(impl *VerifyJwtImpl) {
		impl.ExtractTokenFunc = extractTokenFunc
	}
}

func WithParser(parser *jwt.Parser) Option {
	if parser == nil {
		log.Fatal("Parser cannot be empty")
	}
	return func(impl *VerifyJwtImpl) {
		impl.Parser = parser
	}
}

func WithKeyFunc(keyFunc jwt.Keyfunc) Option {
	if keyFunc == nil {
		log.Fatal("Keyfunc cannot be empty")
	}
	return func(impl *VerifyJwtImpl) {
		impl.KeyFunc = keyFunc
	}
}

func NewDefaultJwtImpl(secretKey []byte, option ...Option) *VerifyJwtImpl {
	verifyJwtFailedFunc := VerifyJwtFailedFunc(func(ctx context.Context,
		req interface{}, err error) (bool, error) {
		return false, errors.New("verify jwt failed, err: " + err.Error())
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

	v := &VerifyJwtImpl{
		SecretKey:           secretKey,
		VerifyJwtFailedFunc: verifyJwtFailedFunc,
		ExtractTokenFunc:    extractTokenFunc,
		Parser:              parser,
		KeyFunc:             keyFunc,
	}

	// overwrite default value
	apply(v, option...)

	return v
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
		KeyFunc:             keyFunc,
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
	if obj.KeyFunc == nil {
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
		var parsed *jwt.Token
		token, err := verifyJwt.ExtractTokenFunc(ctx, req)
		if err != nil {
			pass, err := verifyJwt.VerifyJwtFailedFunc(ctx, req, err)
			if err != nil {
				return nil, err
			}
			if pass {
				goto doHandle
			}
		}

		parsed, err = verifyJwt.Parser.Parse(token, verifyJwt.KeyFunc)
		if err != nil {
			pass, err := verifyJwt.VerifyJwtFailedFunc(ctx, req, err)
			if err != nil {
				return nil, err
			}
			if pass {
				goto doHandle
			}
		}

		if !parsed.Valid {
			pass, err := verifyJwt.VerifyJwtFailedFunc(ctx, req, err)
			if err != nil {
				return nil, err
			}
			if pass {
				goto doHandle
			}
		}

	doHandle:
		return handler(ctx, req)
	}
}
