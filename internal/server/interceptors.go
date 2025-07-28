package server

import (
	"context"
	"strings"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// ErrUnauthenticated is returned when authentication fails or a token is missing.
var ErrUnauthenticated = errors.New("unauthenticated")

// contextKey is a custom type used to avoid key collisions in context values.
type contextKey string

// userIDKey is the context key used to store the authenticated user's ID.
const userIDKey contextKey = "user_id"

// bearerPrefix is the prefix used in the Authorization header for Bearer tokens.
const bearerPrefix = "Bearer "

// ChainUnaryInterceptors chains multiple gRPC unary interceptors into a single interceptor.
func ChainUnaryInterceptors(interceptors ...grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		chain := handler

		for i := len(interceptors) - 1; i >= 0; i-- {
			interceptor := interceptors[i]

			next := chain

			chain = func(currentCtx context.Context, currentReq interface{}) (interface{}, error) {
				return interceptor(currentCtx, currentReq, info, next)
			}
		}

		return chain(ctx, req)
	}
}

// LogUnaryInterceptor logs request metadata including method name, duration, and error if any.
func (s *Server) LogUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		latency := time.Since(start)

		s.log.Debugf("%s took %s | method=%s | error=%v", req, latency, info.FullMethod, err)

		return resp, err
	}
}

// AuthInterceptor validates JWT tokens and injects user ID into the request context.
func (s *Server) AuthInterceptor(excludedMethods map[string]bool) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		// Пропустить проверку токена для публичных методов
		if excludedMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, ErrUnauthenticated
		}

		authHeader := md.Get("authorization")
		if len(authHeader) == 0 {
			return nil, ErrUnauthenticated
		}

		token := strings.TrimPrefix(authHeader[0], bearerPrefix)
		userID, err := parseJWT(token)
		if err != nil {
			return nil, ErrUnauthenticated
		}

		// передаём user_id дальше
		ctx = ContextWithUserID(ctx, userID)
		return handler(ctx, req)
	}
}

// ContextWithUserID returns a new context with the given user ID.
func ContextWithUserID(ctx context.Context, uid uint64) context.Context {
	return context.WithValue(ctx, userIDKey, uid)
}

// UserIDFromContext extracts the user ID from the given context.
func UserIDFromContext(ctx context.Context) (uint64, error) {
	uid, ok := ctx.Value(userIDKey).(uint64)
	if !ok {
		return 0, ErrUnauthenticated
	}
	return uid, nil
}
