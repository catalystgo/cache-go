package cache_test

import (
	"context"
	"testing"

	"github.com/catalystgo/cache-go/cache"
	"github.com/catalystgo/cache-go/mock"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

type testRequest string

func (t testRequest) String() string {
	return string(t)
}

type testRequestNotStringer struct{}

func TestInterceptor(t *testing.T) {
	t.Parallel()

	const testMethodName = "my-test-method"
	const testResponse = "my-test-response"

	cases := []struct {
		name    string
		prepare func(ctrl *gomock.Controller, registry *mock.MockRegistry)
		request interface{}
		handler func(ctx context.Context, req interface{}) (interface{}, error)
		assert  func(resp interface{}, err error)
	}{
		{
			name:    "request is not a stringer",
			request: testRequestNotStringer{},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, assert.AnError
			},
			assert: func(resp interface{}, err error) {
				require.Nil(t, resp)
				require.Equal(t, assert.AnError, err)
			},
		},
		{
			name:    "there is no cache for the given method",
			request: testRequest("my-test-request"),
			prepare: func(ctrl *gomock.Controller, registry *mock.MockRegistry) {
				registry.EXPECT().GetByName(testMethodName).Return(nil, false)
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, assert.AnError
			},
			assert: func(resp interface{}, err error) {
				require.Nil(t, resp)
				require.Equal(t, assert.AnError, err)
			},
		},
		{
			name:    "not found in cache, successful response",
			request: testRequest("my-test-request"),
			prepare: func(ctrl *gomock.Controller, registry *mock.MockRegistry) {
				cache := mock.NewMockNamedCache(ctrl)
				registry.EXPECT().GetByName(testMethodName).Return(cache, true)
				cache.EXPECT().Put("my-test-request", testResponse).Return()
				cache.EXPECT().Get("my-test-request").Return(nil, false)
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return testResponse, nil
			},
			assert: func(resp interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, testResponse, resp)
			},
		},
		{
			name:    "not found in cache, failed response",
			request: testRequest("my-test-request"),
			prepare: func(ctrl *gomock.Controller, registry *mock.MockRegistry) {
				cache := mock.NewMockNamedCache(ctrl)
				registry.EXPECT().GetByName(testMethodName).Return(cache, true)
				cache.EXPECT().Get("my-test-request").Return(nil, false)
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, assert.AnError
			},
			assert: func(resp interface{}, err error) {
				require.Error(t, err)
			},
		},
		{
			name:    "found in cache",
			request: testRequest("my-test-request"),
			prepare: func(ctrl *gomock.Controller, registry *mock.MockRegistry) {
				cache := mock.NewMockNamedCache(ctrl)
				registry.EXPECT().GetByName(testMethodName).Return(cache, true)
				cache.EXPECT().Get("my-test-request").Return(testResponse, true)
			},
			handler: func(ctx context.Context, req interface{}) (interface{}, error) {
				require.FailNow(t, "should not be called")
				return nil, nil
			},
			assert: func(resp interface{}, err error) {
				require.NoError(t, err)
				require.Equal(t, testResponse, resp)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			registry := mock.NewMockRegistry(ctrl)
			if tc.prepare != nil {
				tc.prepare(ctrl, registry)
			}
			intercept := cache.NewInterceptor(registry)
			serverInfo := &grpc.UnaryServerInfo{FullMethod: testMethodName}

			// act
			resp, err := intercept(
				context.Background(),
				tc.request,
				serverInfo,
				tc.handler,
			)

			// assert
			tc.assert(resp, err)
		})
	}
}
