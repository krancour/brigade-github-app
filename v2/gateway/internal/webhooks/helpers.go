package webhooks

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/google/go-github/v32/github"
)

// TODO: Rename this file. "helpers.go" smells bad.

var (
	branchRefRegex = regexp.MustCompile("refs/heads/(.+)")
	tagRefRegex    = regexp.MustCompile("refs/tags/(.+)")
)

func (s *service) isAllowedPullRequest(e *github.PullRequestEvent) bool {
	isFork := e.PullRequest.Head.Repo.GetFork()
	// This applies the author association to forked PRs.
	// PRs sent against origin will be accepted without a check.
	// See https://developer.github.com/v4/reference/enum/commentauthorassociation/
	if assoc := e.PullRequest.GetAuthorAssociation(); isFork && !s.isAllowedAuthor(assoc) {
		log.Printf("skipping pull request for disallowed author %s", assoc)
		return false
	}
	switch e.GetAction() {
	case "opened", "synchronize", "reopened", "labeled", "unlabeled", "closed":
		return true
	}
	log.Println("unsupported pull_request action:", e.GetAction())
	return false
}

func (s *service) isAllowedAuthor(author string) bool {
	for _, a := range s.config.AllowedAuthors {
		if a == author {
			return true
		}
	}
	return false
}

func (s *service) shouldEmit(eventType string) bool {
	unqualifiedEventType := strings.Split(eventType, ":")[0]
	for _, emitableEvent := range s.config.EmittedEvents {
		if eventType == emitableEvent || unqualifiedEventType == emitableEvent ||
			emitableEvent == "*" {
			return true
		}
	}
	return false
}

func getTitlesFromPushEvent(pe *github.PushEvent) (string, string) {
	var shortTitle, longTitle string
	if pe != nil && pe.Ref != nil {
		if refSubmatches :=
			branchRefRegex.FindStringSubmatch(*pe.Ref); len(refSubmatches) == 2 {
			shortTitle = fmt.Sprintf("branch: %s", refSubmatches[1])
			longTitle = shortTitle
		} else if refSubmatches :=
			tagRefRegex.FindStringSubmatch(*pe.Ref); len(refSubmatches) == 2 {
			shortTitle = fmt.Sprintf("tag: %s", refSubmatches[1])
			longTitle = shortTitle
		}
	}
	return shortTitle, longTitle
}

// func getTitlesFromIssue(issue *github.Issue) (string, string) {
// 	var shortTitle, longTitle string
// 	if issue != nil && issue.Number != nil {
// 		shortTitle = fmt.Sprintf("Issue #%d", *issue.Number)
// 		if issue.Title != nil {
// 			longTitle = fmt.Sprintf("%s: %s", shortTitle, *issue.Title)
// 		}
// 	}
// 	return shortTitle, longTitle
// }

func getTitlesFromPR(pr *github.PullRequest) (string, string) {
	var shortTitle, longTitle string
	if pr != nil && pr.Number != nil {
		shortTitle = fmt.Sprintf("PR #%d", *pr.Number)
		if pr.Title != nil {
			longTitle = fmt.Sprintf("%s: %s", shortTitle, *pr.Title)
		}
	}
	return shortTitle, longTitle
}
