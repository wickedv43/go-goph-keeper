package server

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestChainUnaryInterceptors(t *testing.T) {
	var calls []string

	i1 := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		calls = append(calls, "i1:before")
		resp, err := handler(ctx, req)
		calls = append(calls, "i1:after")
		return resp, err
	}

	i2 := func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		calls = append(calls, "i2:before")
		resp, err := handler(ctx, req)
		calls = append(calls, "i2:after")
		return resp, err
	}

	finalHandler := func(ctx context.Context, req interface{}) (interface{}, error) {
		calls = append(calls, "handler")
		return "ok", nil
	}

	// Оборачиваем
	chain := ChainUnaryInterceptors(i1, i2)

	// Вызываем цепочку
	resp, err := chain(
		context.Background(),
		"request",
		&grpc.UnaryServerInfo{FullMethod: "/test.Method"},
		finalHandler,
	)

	require.NoError(t, err)
	require.Equal(t, "ok", resp)

	// Проверка порядка вызовов
	expected := []string{
		"i1:before",
		"i2:before",
		"handler",
		"i2:after",
		"i1:after",
	}
	require.Equal(t, expected, calls)
}

func TestServer_LogUnaryInterceptor(t *testing.T) {
	var handlerCalled bool

	s := &Server{
		log: zap.NewNop().Sugar(), // не паникует и не пишет
	}

	interceptor := s.LogUnaryInterceptor()

	fakeReq := "test-request"
	expectedResp := "test-response"
	expectedErr := errors.New("handler error")

	t.Run("success: logs and returns handler result", func(t *testing.T) {
		handlerCalled = false

		resp, err := interceptor(
			context.Background(),
			fakeReq,
			&grpc.UnaryServerInfo{FullMethod: "/gk.TestMethod"},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return expectedResp, nil
			},
		)

		require.True(t, handlerCalled)
		require.NoError(t, err)
		require.Equal(t, expectedResp, resp)
	})

	t.Run("error: handler returns error", func(t *testing.T) {
		handlerCalled = false

		resp, err := interceptor(
			context.Background(),
			fakeReq,
			&grpc.UnaryServerInfo{FullMethod: "/gk.TestError"},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return nil, expectedErr
			},
		)

		require.True(t, handlerCalled)
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, resp)
	})
}

func TestAuthInterceptor_WithJWT(t *testing.T) {
	s := &Server{
		log: zap.NewNop().Sugar(),
	}

	t.Run("excluded method bypasses auth", func(t *testing.T) {
		called := false

		interceptor := s.AuthInterceptor(map[string]bool{
			"/gk.Public": true,
		})

		_, err := interceptor(
			context.Background(),
			nil,
			&grpc.UnaryServerInfo{FullMethod: "/gk.Public"},
			func(ctx context.Context, req interface{}) (interface{}, error) {
				called = true
				return "ok", nil
			},
		)

		require.NoError(t, err)
		require.True(t, called)
	})

	t.Run("missing metadata returns unauthenticated", func(t *testing.T) {
		interceptor := s.AuthInterceptor(nil)

		resp, err := interceptor(context.Background(), nil, &grpc.UnaryServerInfo{
			FullMethod: "/gk/Secured",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		})

		require.ErrorIs(t, err, ErrUnauthenticated)
		require.Nil(t, resp)
	})

	t.Run("missing authorization header returns unauthenticated", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{})

		interceptor := s.AuthInterceptor(nil)

		resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
			FullMethod: "/gk/Secured",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		})

		require.ErrorIs(t, err, ErrUnauthenticated)
		require.Nil(t, resp)
	})

	t.Run("invalid JWT returns unauthenticated", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			"authorization": {"Bearer not-a-token"},
		})

		interceptor := s.AuthInterceptor(nil)

		resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
			FullMethod: "/gk/Secured",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			t.Fatal("handler should not be called")
			return nil, nil
		})

		require.ErrorIs(t, err, ErrUnauthenticated)
		require.Nil(t, resp)
	})

	t.Run("valid JWT sets userID in context", func(t *testing.T) {
		token, err := generateJWT(123)
		require.NoError(t, err)

		ctx := metadata.NewIncomingContext(context.Background(), metadata.MD{
			"authorization": {"Bearer " + token},
		})

		interceptor := s.AuthInterceptor(nil)

		called := false

		resp, err := interceptor(ctx, nil, &grpc.UnaryServerInfo{
			FullMethod: "/gk/Secured",
		}, func(ctx context.Context, req interface{}) (interface{}, error) {
			called = true
			uid, err := UserIDFromContext(ctx)
			require.NoError(t, err)
			require.Equal(t, uint64(123), uid)
			return "ok", nil
		})

		require.NoError(t, err)
		require.Equal(t, "ok", resp)
		require.True(t, called)
	})
}
