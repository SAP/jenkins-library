// Code generated by mockery v2.10.4. DO NOT EDIT.

package mocks

import (
	context "context"

	github "github.com/google/go-github/v45/github"
	mock "github.com/stretchr/testify/mock"

	os "os"
)

// GithubRepoClient is an autogenerated mock type for the GithubRepoClient type
type GithubRepoClient struct {
	mock.Mock
}

// CreateRelease provides a mock function with given fields: ctx, owner, repo, release
func (_m *GithubRepoClient) CreateRelease(ctx context.Context, owner string, repo string, release *github.RepositoryRelease) (*github.RepositoryRelease, *github.Response, error) {
	ret := _m.Called(ctx, owner, repo, release)

	var r0 *github.RepositoryRelease
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *github.RepositoryRelease) *github.RepositoryRelease); ok {
		r0 = rf(ctx, owner, repo, release)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.RepositoryRelease)
		}
	}

	var r1 *github.Response
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *github.RepositoryRelease) *github.Response); ok {
		r1 = rf(ctx, owner, repo, release)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*github.Response)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string, *github.RepositoryRelease) error); ok {
		r2 = rf(ctx, owner, repo, release)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DeleteReleaseAsset provides a mock function with given fields: ctx, owner, repo, id
func (_m *GithubRepoClient) DeleteReleaseAsset(ctx context.Context, owner string, repo string, id int64) (*github.Response, error) {
	ret := _m.Called(ctx, owner, repo, id)

	var r0 *github.Response
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64) *github.Response); ok {
		r0 = rf(ctx, owner, repo, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, int64) error); ok {
		r1 = rf(ctx, owner, repo, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatestRelease provides a mock function with given fields: ctx, owner, repo
func (_m *GithubRepoClient) GetLatestRelease(ctx context.Context, owner string, repo string) (*github.RepositoryRelease, *github.Response, error) {
	ret := _m.Called(ctx, owner, repo)

	var r0 *github.RepositoryRelease
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *github.RepositoryRelease); ok {
		r0 = rf(ctx, owner, repo)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.RepositoryRelease)
		}
	}

	var r1 *github.Response
	if rf, ok := ret.Get(1).(func(context.Context, string, string) *github.Response); ok {
		r1 = rf(ctx, owner, repo)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*github.Response)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string) error); ok {
		r2 = rf(ctx, owner, repo)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ListReleaseAssets provides a mock function with given fields: ctx, owner, repo, id, opt
func (_m *GithubRepoClient) ListReleaseAssets(ctx context.Context, owner string, repo string, id int64, opt *github.ListOptions) ([]*github.ReleaseAsset, *github.Response, error) {
	ret := _m.Called(ctx, owner, repo, id, opt)

	var r0 []*github.ReleaseAsset
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64, *github.ListOptions) []*github.ReleaseAsset); ok {
		r0 = rf(ctx, owner, repo, id, opt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*github.ReleaseAsset)
		}
	}

	var r1 *github.Response
	if rf, ok := ret.Get(1).(func(context.Context, string, string, int64, *github.ListOptions) *github.Response); ok {
		r1 = rf(ctx, owner, repo, id, opt)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*github.Response)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string, int64, *github.ListOptions) error); ok {
		r2 = rf(ctx, owner, repo, id, opt)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// UploadReleaseAsset provides a mock function with given fields: ctx, owner, repo, id, opt, file
func (_m *GithubRepoClient) UploadReleaseAsset(ctx context.Context, owner string, repo string, id int64, opt *github.UploadOptions, file *os.File) (*github.ReleaseAsset, *github.Response, error) {
	ret := _m.Called(ctx, owner, repo, id, opt, file)

	var r0 *github.ReleaseAsset
	if rf, ok := ret.Get(0).(func(context.Context, string, string, int64, *github.UploadOptions, *os.File) *github.ReleaseAsset); ok {
		r0 = rf(ctx, owner, repo, id, opt, file)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*github.ReleaseAsset)
		}
	}

	var r1 *github.Response
	if rf, ok := ret.Get(1).(func(context.Context, string, string, int64, *github.UploadOptions, *os.File) *github.Response); ok {
		r1 = rf(ctx, owner, repo, id, opt, file)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*github.Response)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, string, string, int64, *github.UploadOptions, *os.File) error); ok {
		r2 = rf(ctx, owner, repo, id, opt, file)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
