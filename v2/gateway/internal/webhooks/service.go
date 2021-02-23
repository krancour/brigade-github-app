package webhooks

import (
	"context"
	"fmt"
	"net/http"

	"github.com/brigadecore/brigade/sdk/v2/core"
	"github.com/google/go-github/v32/github"
	gin "gopkg.in/gin-gonic/gin.v1"
)

type ServiceConfig struct {
	AllowedAuthors []string
	EmittedEvents  []string
}

type Service interface {
	Handle(ctx context.Context, srcEvent interface{}, payload []byte) error
}

type service struct {
	eventsClient core.EventsClient
	config       ServiceConfig
}

func NewService(
	eventsClient core.EventsClient,
	config ServiceConfig,
) Service {
	return &service{
		eventsClient: eventsClient,
		config:       config,
	}
}

func (s *service) Handle(
	ctx context.Context,
	srcEvent interface{},
	payload []byte,
) error {
	brigadeEvent := core.Event{
		Source:  "github.com/brigadecore/brigade-github-gateway/v2",
		Payload: string(payload),
	}

	switch e := srcEvent.(type) {

	case *github.CheckRunEvent:
		brigadeEvent.Type = fmt.Sprintf("check_run:%s", e.GetAction())
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.CheckRun.CheckSuite.GetHeadSHA(),
			Ref:    e.CheckRun.CheckSuite.GetHeadBranch(),
		}

	case *github.CheckSuiteEvent:
		brigadeEvent.Type = fmt.Sprintf("check_suite:%s", e.GetAction())
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.CheckSuite.GetHeadSHA(),
			Ref:    e.CheckSuite.GetHeadBranch(),
		}

	case *github.CommitCommentEvent:
		brigadeEvent.Type = fmt.Sprintf("commit_comment:%s", e.GetAction())
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.Comment.GetCommitID(),
		}

	case *github.CreateEvent:
		brigadeEvent.Type = "create"
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		// TODO: There are three ref_type values: tag, branch, and repo. Do we
		// want to be opinionated about how we handle these?
		brigadeEvent.Git = &core.GitDetails{
			Ref: e.GetRef(),
		}

	case *github.DeploymentEvent:
		brigadeEvent.Type = "deployment"
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.Deployment.GetSHA(),
			Ref:    e.Deployment.GetRef(),
		}

	case *github.DeploymentStatusEvent:
		brigadeEvent.Type = "deployment_status"
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.Deployment.GetSHA(),
			Ref:    e.Deployment.GetRef(),
		}

	case *github.PingEvent:
		brigadeEvent.Type = "ping"

	case *github.PullRequestEvent:
		if !s.isAllowedPullRequest(e) {
			// TODO: Need to devise a method of returning errors of a non-technical
			// nature
			c.JSON(http.StatusOK, gin.H{"status": "build skipped"})
			return nil
		}
		brigadeEvent.Type = fmt.Sprintf("pull_request:%s", e.GetAction())
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.ShortTitle, brigadeEvent.LongTitle =
			getTitlesFromPR(e.PullRequest)
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.PullRequest.Head.GetSHA(),
			Ref:    fmt.Sprintf("refs/pull/%d/head", e.PullRequest.GetNumber()),
		}

	case *github.PullRequestReviewEvent:
		brigadeEvent.Type = fmt.Sprintf("pull_request_review:%s", e.GetAction())
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.ShortTitle, brigadeEvent.LongTitle =
			getTitlesFromPR(e.PullRequest)
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.PullRequest.Head.GetSHA(),
			Ref:    fmt.Sprintf("refs/pull/%d/head", e.PullRequest.GetNumber()),
		}

	case *github.PullRequestReviewCommentEvent:
		brigadeEvent.Type =
			fmt.Sprintf("pull_request_review_comment:%s", e.GetAction())
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.ShortTitle, brigadeEvent.LongTitle =
			getTitlesFromPR(e.PullRequest)
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.PullRequest.Head.GetSHA(),
			Ref:    fmt.Sprintf("refs/pull/%d/head", e.PullRequest.GetNumber()),
		}

	case *github.PushEvent:
		brigadeEvent.Type = "push"
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		// If this is a branch deletion, skip the build.
		if e.GetDeleted() {
			// TODO: Need to devise a method of returning errors of a non-technical
			// nature
			c.JSON(http.StatusOK, gin.H{"status": "build skipped on branch deletion"})
			return nil
		}
		brigadeEvent.ShortTitle, brigadeEvent.LongTitle = getTitlesFromPushEvent(e)
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.HeadCommit.GetID(),
			Ref:    e.GetRef(),
		}

	case *github.ReleaseEvent:
		brigadeEvent.Type = fmt.Sprintf("release:%s", e.GetAction())
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.Git = &core.GitDetails{
			Ref: e.Release.GetTagName(),
		}

	case *github.StatusEvent:
		brigadeEvent.Type = "status"
		brigadeEvent.Labels = core.Labels{
			"repo": e.Repo.GetFullName(),
		}
		brigadeEvent.Git = &core.GitDetails{
			Commit: e.Commit.GetSHA(),
		}

	default:
		// TODO: Need to devise a method of returning errors of a non-technical
		// nature
		c.JSON(http.StatusNotImplemented, gin.H{"status": "unsupported event type"})
		return nil
	}

	if s.shouldEmit(brigadeEvent.Type) {
		_, err := s.eventsClient.Create(ctx, brigadeEvent)
		return err
	}

	return nil
}
