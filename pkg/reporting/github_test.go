package reporting

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

type scanReportlMock struct {
	markdown       []byte
	text           string
	title          string
	failToMarkdown bool
}

func (s *scanReportlMock) Title() string {
	return s.title
}

func (s *scanReportlMock) ToMarkdown() ([]byte, error) {
	if s.failToMarkdown {
		return s.markdown, fmt.Errorf("toMarkdown failure")
	}
	return s.markdown, nil
}

func (s *scanReportlMock) ToTxt() string {
	return s.text
}

type ghServicesMock struct {
	issues               []*github.Issue
	createError          error
	comment              *github.IssueComment
	createCommmentNumber int
	createCommentError   error
	editError            error
	editNumber           int
	editRequest          *github.IssueRequest
	searchError          error
	searchOpts           *github.SearchOptions
	searchQuery          string
	searchResult         []*github.Issue
}

func (g *ghServicesMock) Create(ctx context.Context, owner string, repo string, issueRequest *github.IssueRequest) (*github.Issue, *github.Response, error) {
	if g.issues == nil {
		g.issues = []*github.Issue{}
	}
	number := len(g.issues) + 1
	id := int64(number)
	assignees := []*github.User{}
	if issueRequest.Assignees != nil {
		for _, userName := range *issueRequest.Assignees {
			// cannot use userName directly since variable is re-used in the loop and thus name in assignees would be pointing to final value only.
			user := userName
			assignees = append(assignees, &github.User{Name: &user})
		}
	}

	theIssue := github.Issue{
		ID:        &id,
		Number:    &number,
		Title:     issueRequest.Title,
		Body:      issueRequest.Body,
		Assignees: assignees,
	}
	g.issues = append(g.issues, &theIssue)
	if g.createError != nil {
		return &theIssue, &github.Response{}, g.createError
	}
	return &theIssue, &github.Response{}, nil
}

func (g *ghServicesMock) CreateComment(ctx context.Context, owner string, repo string, number int, comment *github.IssueComment) (*github.IssueComment, *github.Response, error) {
	g.createCommmentNumber = number
	g.comment = comment
	if g.createCommentError != nil {
		return &github.IssueComment{}, &github.Response{}, g.createCommentError
	}
	return &github.IssueComment{}, &github.Response{}, nil
}

func (g *ghServicesMock) Edit(ctx context.Context, owner string, repo string, number int, issueRequest *github.IssueRequest) (*github.Issue, *github.Response, error) {
	g.editNumber = number
	g.editRequest = issueRequest
	if g.editError != nil {
		return &github.Issue{}, &github.Response{}, g.editError
	}
	return &github.Issue{}, &github.Response{}, nil
}

func (g *ghServicesMock) Issues(ctx context.Context, query string, opts *github.SearchOptions) (*github.IssuesSearchResult, *github.Response, error) {
	g.searchOpts = opts
	g.searchQuery = query

	if g.searchError != nil {
		return &github.IssuesSearchResult{Issues: g.searchResult}, &github.Response{}, g.searchError
	}
	return &github.IssuesSearchResult{Issues: g.searchResult}, &github.Response{}, nil
}

var (
	owner      string = "testOwner"
	repository string = "testRepository"
)

func TestUploadSingleReport(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{}
		s := scanReportlMock{title: "The Title", markdown: []byte("# The Markdown")}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		err := gh.UploadSingleReport(ctx, &s)

		assert.NoError(t, err)
		assert.Equal(t, s.title, ghMock.issues[0].GetTitle())
		assert.Equal(t, string(s.markdown), ghMock.issues[0].GetBody())
	})

	t.Run("error case", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{createError: fmt.Errorf("create failed")}
		s := scanReportlMock{title: "The Title"}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		err := gh.UploadSingleReport(ctx, &s)

		assert.EqualError(t, err, "failed to upload results for 'The Title' into GitHub issue: failed to create issue: create failed")
	})
}

func TestUploadMultipleReports(t *testing.T) {
	t.Parallel()

	t.Run("success case", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{}
		s1 := scanReportlMock{title: "The Title 1", markdown: []byte("# The Markdown 1")}
		s2 := scanReportlMock{title: "The Title 2", markdown: []byte("# The Markdown 2")}
		s := []IssueDetail{&s1, &s2}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		err := gh.UploadMultipleReports(ctx, &s)

		assert.NoError(t, err)
		assert.Equal(t, s1.title, ghMock.issues[0].GetTitle())
		assert.Equal(t, string(s1.markdown), ghMock.issues[0].GetBody())
		assert.Equal(t, s2.title, ghMock.issues[1].GetTitle())
		assert.Equal(t, string(s2.markdown), ghMock.issues[1].GetBody())
	})

	t.Run("error case", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{createError: fmt.Errorf("create failed")}
		s1 := scanReportlMock{title: "The Title 1", markdown: []byte("# The Markdown 1")}
		s2 := scanReportlMock{title: "The Title 2", markdown: []byte("# The Markdown 2")}
		s := []IssueDetail{&s1, &s2}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		err := gh.UploadMultipleReports(ctx, &s)

		assert.EqualError(t, err, "failed to upload results for 'The Title 1' into GitHub issue: failed to create issue: create failed")
	})
}

func TestCreateIssueOrUpdateIssueComment(t *testing.T) {
	t.Parallel()

	title := "The Title"
	assignees := []string{"assignee1", "assignee2"}

	t.Run("success case - new issue", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			Assignees:     &assignees,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}
		markdown := "# The Markdown"

		err := gh.createIssueOrUpdateIssueComment(ctx, title, markdown)

		assert.NoError(t, err)
		assert.Equal(t, title, ghMock.issues[0].GetTitle())
		assert.Equal(t, markdown, ghMock.issues[0].GetBody())
		assert.Equal(t, assignees[0], ghMock.issues[0].Assignees[0].GetName())
		assert.Equal(t, assignees[1], ghMock.issues[0].Assignees[1].GetName())
	})

	t.Run("success case - update issue", func(t *testing.T) {
		ctx := context.Background()
		number := 1
		state := "open"
		title := "The Title"
		body := "the body of the issue"
		issue := github.Issue{Number: &number, State: &state, Title: &title, Body: &body}
		ghMock := ghServicesMock{searchResult: []*github.Issue{&issue}}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}
		markdown := "# The Markdown"

		err := gh.createIssueOrUpdateIssueComment(ctx, title, markdown)
		assert.NoError(t, err)
		assert.Equal(t, number, ghMock.editNumber)
		assert.Equal(t, markdown, ghMock.editRequest.GetBody())
		assert.Equal(t, number, ghMock.createCommmentNumber)
		assert.Equal(t, "issue content has been updated", ghMock.comment.GetBody())
	})

	t.Run("success case - no update", func(t *testing.T) {
		ctx := context.Background()
		number := 1
		state := "open"
		title := "The Title"
		body := "the body of the issue"
		issue := github.Issue{Number: &number, State: &state, Title: &title, Body: &body}
		ghMock := ghServicesMock{searchResult: []*github.Issue{&issue}}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}
		markdown := "the body of the issue"

		err := gh.createIssueOrUpdateIssueComment(ctx, title, markdown)
		assert.NoError(t, err)
		assert.Nil(t, ghMock.editRequest)
		assert.Nil(t, ghMock.comment)
	})

	t.Run("error case - lookup failed", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{searchError: fmt.Errorf("search failed")}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}
		markdown := "# The Markdown"

		err := gh.createIssueOrUpdateIssueComment(ctx, title, markdown)

		assert.EqualError(t, err, "error when looking up issue: error occurred when looking for existing issue: search failed")
	})

	t.Run("error case - issue creation failed", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{createError: fmt.Errorf("creation failed")}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}
		markdown := "# The Markdown"

		err := gh.createIssueOrUpdateIssueComment(ctx, title, markdown)

		assert.EqualError(t, err, "failed to create issue: creation failed")
	})

	t.Run("error case - issue editing failed", func(t *testing.T) {
		ctx := context.Background()
		number := 1
		state := "open"
		title := "The Title"
		body := "the body of the issue"
		issue := github.Issue{Number: &number, State: &state, Title: &title, Body: &body}
		ghMock := ghServicesMock{searchResult: []*github.Issue{&issue}, editError: fmt.Errorf("edit failed")}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}
		markdown := "# The Markdown"

		err := gh.createIssueOrUpdateIssueComment(ctx, title, markdown)
		assert.EqualError(t, err, "failed to edit issue: edit failed")
	})

	t.Run("error case - edit comment creation failed", func(t *testing.T) {
		ctx := context.Background()
		number := 1
		state := "open"
		title := "The Title"
		body := "the body of the issue"
		issue := github.Issue{Number: &number, State: &state, Title: &title, Body: &body}
		ghMock := ghServicesMock{searchResult: []*github.Issue{&issue}, createCommentError: fmt.Errorf("comment failed")}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}
		markdown := "# The Markdown"

		err := gh.createIssueOrUpdateIssueComment(ctx, title, markdown)
		assert.EqualError(t, err, "failed to create comment: comment failed")

	})
}

func TestFindExistingIssue(t *testing.T) {
	t.Parallel()

	t.Run("success case - issue found", func(t *testing.T) {
		ctx := context.Background()
		number := 1
		state := "open"
		title := "The Title"
		body := "the body of the issue"
		issue := github.Issue{Number: &number, State: &state, Title: &title, Body: &body}
		ghMock := ghServicesMock{searchResult: []*github.Issue{&issue}}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		i, b, err := gh.findExistingIssue(ctx, title)

		assert.NoError(t, err)
		assert.Equal(t, fmt.Sprintf("is:issue repo:%v/%v in:title %v", *gh.Owner, *gh.Repository, title), ghMock.searchQuery)
		assert.Equal(t, 1, i)
		assert.Equal(t, body, b)
	})

	t.Run("success case - issue found, reopen", func(t *testing.T) {
		ctx := context.Background()
		number := 1
		state := "closed"
		title := "The Title"
		body := "the body of the issue"
		issue := github.Issue{Number: &number, State: &state, Title: &title, Body: &body}
		ghMock := ghServicesMock{searchResult: []*github.Issue{&issue}}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		i, b, err := gh.findExistingIssue(ctx, "The Title")

		assert.NoError(t, err)
		assert.Equal(t, 1, ghMock.editNumber)
		assert.Equal(t, "open", ghMock.editRequest.GetState())
		assert.Equal(t, 1, i)
		assert.Equal(t, body, b)
	})

	t.Run("success case - no issue found", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		i, body, err := gh.findExistingIssue(ctx, "The Title")

		assert.NoError(t, err)
		assert.Equal(t, 0, i)
		assert.Equal(t, "", body)
	})

	t.Run("error case - search failed", func(t *testing.T) {
		ctx := context.Background()
		ghMock := ghServicesMock{searchError: fmt.Errorf("search failed")}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		i, _, err := gh.findExistingIssue(ctx, "The Title")

		assert.EqualError(t, err, "error occurred when looking for existing issue: search failed")
		assert.Equal(t, 0, i)
	})

	t.Run("error case - reopen failed", func(t *testing.T) {
		ctx := context.Background()
		number := 1
		state := "closed"
		title := "The Title"
		issue := github.Issue{Number: &number, State: &state, Title: &title}
		ghMock := ghServicesMock{editError: fmt.Errorf("reopen failed"), searchResult: []*github.Issue{&issue}}
		gh := GitHub{
			Owner:         &owner,
			Repository:    &repository,
			IssueService:  &ghMock,
			SearchService: &ghMock,
		}

		_, _, err := gh.findExistingIssue(ctx, "The Title")
		assert.EqualError(t, err, "failed to re-open issue: reopen failed")
	})
}
