//go:build unit
// +build unit

package jenkins

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SAP/jenkins-library/pkg/jenkins/mocks"
	"github.com/bndr/gojenkins"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestInterfaceCompatibility(t *testing.T) {
	var _ Jenkins = new(gojenkins.Jenkins)
	var _ Build = new(gojenkins.Build)
}

func TestTriggerJob(t *testing.T) {
	ctx := context.Background()
	jobParameters := map[string]string{}

	t.Run("error - job not updated", func(t *testing.T) {
		// init
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(404, fmt.Errorf("%s", mock.Anything))
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, fmt.Sprintf("failed to load job: %s", mock.Anything))
		assert.Nil(t, build)
	})
	t.Run("error - task not started", func(t *testing.T) {
		// init
		queueID := int64(0)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, map[string]string{}).
			Return(queueID, fmt.Errorf("%s", mock.Anything))
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, build)
	})
	t.Run("error - task already queued", func(t *testing.T) {
		// init
		queueID := int64(0)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, jobParameters).
			Return(queueID, nil)
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, "unable to queue build")
		assert.Nil(t, build)
	})
	t.Run("error - task not queued", func(t *testing.T) {
		// init
		queueID := int64(43)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, jobParameters).
			Return(queueID, nil).
			On("GetJob").
			Return(&gojenkins.Job{})
		jenkins.
			On("GetBuildFromQueueID", ctx, mock.Anything, queueID).
			Return(nil, fmt.Errorf("%s", mock.Anything))
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		job.AssertExpectations(t)
		jenkins.AssertExpectations(t)
		assert.EqualError(t, err, mock.Anything)
		assert.Nil(t, build)
	})
	t.Run("success", func(t *testing.T) {
		// init
		queueID := int64(43)
		jenkins := &mocks.Jenkins{}
		jenkins.Test(t)
		job := &mocks.Job{}
		job.Test(t)
		job.
			On("Poll", ctx).Return(200, nil).
			On("InvokeSimple", ctx, jobParameters).
			Return(queueID, nil).
			On("GetJob").
			Return(&gojenkins.Job{})
		jenkins.
			On("GetBuildFromQueueID", ctx, mock.Anything, queueID).
			Return(&gojenkins.Build{}, nil)
		// test
		build, err := TriggerJob(ctx, jenkins, job, jobParameters)
		// asserts
		jenkins.AssertExpectations(t)
		assert.NoError(t, err)
		assert.NotNil(t, build)
	})
}

func TestGetBuildFromQueueID(t *testing.T) {
	ctx := context.Background()
	const queueID = int64(42)
	const buildNumber = int64(99)
	const buildBasePath = "/job/myFolder/job/myJob/99" // path without trailing slash, used for handler registration
	const buildPath = buildBasePath + "/"              // Base set on Build (Jenkins URLs end with /)

	// newTestServer returns a httptest.Server whose handler serves minimal
	// Jenkins API responses. queueBody is returned for /queue/item/42/api/json;
	// buildBody for /job/myFolder/job/myJob/99/api/json.
	newTestServer := func(queueBody, buildBody string) *httptest.Server {
		mux := http.NewServeMux()
		mux.HandleFunc(fmt.Sprintf("/queue/item/%d/api/json", queueID), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, queueBody)
		})
		mux.HandleFunc(buildBasePath+"/api/json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, buildBody)
		})
		return httptest.NewServer(mux)
	}

	queueReady := func(buildURL string) string {
		b, _ := json.Marshal(map[string]interface{}{
			"executable": map[string]interface{}{
				"number": buildNumber,
				"url":    buildURL,
			},
		})
		return string(b)
	}
	buildReady := `{"number":99,"result":"SUCCESS","url":"http://ignored/job/myFolder/job/myJob/99/"}`

	t.Run("success - constructs build from path, bypassing host mismatch", func(t *testing.T) {
		// The queue item returns a URL with a different host than the server
		// (simulating the DNS-migration scenario). getBuildFromQueueID must use
		// only the path component, so the build Poll succeeds against the test server.
		buildURL := "https://old-host.example.com:8080" + buildPath
		srv2 := newTestServer(queueReady(buildURL), buildReady)
		defer srv2.Close()

		j := gojenkins.CreateJenkins(srv2.Client(), srv2.URL, "user", "token")
		build, err := getBuildFromQueueID(ctx, j, &gojenkins.Job{}, queueID)

		assert.NoError(t, err)
		assert.NotNil(t, build)
		assert.Equal(t, buildPath, build.Base)
		assert.Equal(t, int64(buildNumber), build.Raw.Number)
	})

	t.Run("success - polls until build number is assigned", func(t *testing.T) {
		// First queue response has Executable.Number==0 (not yet started).
		// Second response has the build number and URL populated.
		callCount := 0
		buildURL := "https://old-host.example.com" + buildPath
		mux := http.NewServeMux()
		mux.HandleFunc(fmt.Sprintf("/queue/item/%d/api/json", queueID), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			callCount++
			if callCount == 1 {
				fmt.Fprint(w, `{"executable":{"number":0,"url":""}}`)
			} else {
				fmt.Fprint(w, queueReady(buildURL))
			}
		})
		mux.HandleFunc(buildBasePath+"/api/json", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, buildReady)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		j := gojenkins.CreateJenkins(srv.Client(), srv.URL, "user", "token")
		build, err := getBuildFromQueueID(ctx, j, &gojenkins.Job{}, queueID)

		assert.NoError(t, err)
		assert.NotNil(t, build)
		assert.Equal(t, buildPath, build.Base)
		assert.GreaterOrEqual(t, callCount, 2)
	})

	t.Run("error - GetQueueItem HTTP failure", func(t *testing.T) {
		// Close the server immediately so the HTTP call gets a connection-refused error.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		j := gojenkins.CreateJenkins(srv.Client(), srv.URL, "user", "token")
		srv.Close()

		build, err := getBuildFromQueueID(ctx, j, &gojenkins.Job{}, queueID)

		assert.Error(t, err)
		assert.Nil(t, build)
	})

	t.Run("error - queue item executable URL is empty", func(t *testing.T) {
		// Executable.Number is non-zero but URL is blank.
		queueBody := `{"executable":{"number":99,"url":""}}`
		mux := http.NewServeMux()
		mux.HandleFunc(fmt.Sprintf("/queue/item/%d/api/json", queueID), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, queueBody)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		j := gojenkins.CreateJenkins(srv.Client(), srv.URL, "user", "token")
		build, err := getBuildFromQueueID(ctx, j, &gojenkins.Job{}, queueID)

		assert.ErrorContains(t, err, "unexpected build URL")
		assert.Nil(t, build)
	})

	t.Run("error - build Poll returns non-200", func(t *testing.T) {
		buildURL := "https://old-host.example.com" + buildPath
		mux := http.NewServeMux()
		mux.HandleFunc(fmt.Sprintf("/queue/item/%d/api/json", queueID), func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, queueReady(buildURL))
		})
		mux.HandleFunc(buildBasePath+"/api/json", func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		})
		srv := httptest.NewServer(mux)
		defer srv.Close()

		j := gojenkins.CreateJenkins(srv.Client(), srv.URL, "user", "token")
		build, err := getBuildFromQueueID(ctx, j, &gojenkins.Job{}, queueID)

		assert.ErrorContains(t, err, "unexpected HTTP status 404")
		assert.Nil(t, build)
	})

	t.Run("fallback - non-gojenkins Jenkins uses GetBuildFromQueueID mock", func(t *testing.T) {
		mockJenkins := &mocks.Jenkins{}
		mockJenkins.Test(t)
		expectedBuild := &gojenkins.Build{}
		mockJenkins.
			On("GetBuildFromQueueID", ctx, mock.Anything, queueID).
			Return(expectedBuild, nil)

		build, err := getBuildFromQueueID(ctx, mockJenkins, &gojenkins.Job{}, queueID)

		mockJenkins.AssertExpectations(t)
		assert.NoError(t, err)
		assert.Equal(t, expectedBuild, build)
	})
}
