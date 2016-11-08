package github

import (
	"fmt"
	"testing"

	"github.com/google/go-github/github"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceGithubPullRequests_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                  func() { testAccPreCheck(t) },
		Providers:                 testAccProviders,
		PreventPostDestroyRefresh: true,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccDataSourceGithubPullRequestsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDateSourceGithubPullRequestsCreatePullRequest("foo"),
				),
			},
			resource.TestStep{
				Config: testAccDataSourceGithubPullRequestsDataConfigMatched,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.github_pull_requests.foo", "id", "873244444"),
					resource.TestCheckResourceAttr("data.github_pull_requests.foo", "pulls.#", "1"),
					resource.TestCheckResourceAttr("data.github_pull_requests.foo", "pulls.0.number", "1"),
					resource.TestCheckResourceAttr("data.github_pull_requests.foo", "pulls.0.title", "Test"),
				),
			},
			resource.TestStep{
				Config: testAccDataSourceGithubPullRequestsDataConfigNotMatchedLabel,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.github_pull_requests.foo", "pulls.#", "0"),
				),
			},
			resource.TestStep{
				Config: testAccDataSourceGithubPullRequestsDataConfigNotMatchedTitle,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.github_pull_requests.foo", "pulls.#", "0"),
				),
			},
		},
	})
}

func testAccDateSourceGithubPullRequestsCreatePullRequest(repo string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*Organization).client
		orgName := testAccProvider.Meta().(*Organization).name
		title := "Test"
		baseRef := "refs/heads/master"
		headRef := "refs/heads/feature"
		pullBase := "master"
		pullHead := "feature"

		// get base branch
		baseGitRef, _, err := conn.Git.GetRef(orgName, repo, baseRef)
		if err != nil {
			return err
		}

		// create new branch
		_, _, err = conn.Git.CreateRef(orgName, repo, &github.Reference{
			Ref: &headRef,
			Object: &github.GitObject{
				SHA: baseGitRef.Object.SHA,
			},
		})
		if err != nil {
			return err
		}

		newFileOption := &github.RepositoryContentFileOptions{
			Branch:  &pullHead,
			Message: &title,
			Content: []byte(title),
			Committer: &github.CommitAuthor{
				Name:  &title,
				Email: &title,
			},
		}

		// modify head branch
		_, _, err = conn.Repositories.CreateFile(orgName, repo, "file.txt", newFileOption)
		if err != nil {
			return err
		}

		// create pull request
		inputHead := fmt.Sprintf("%s:%s", orgName, pullHead)
		input := &github.NewPullRequest{
			Title: &title,
			Base:  &pullBase,
			Head:  &inputHead,
		}

		pull, _, err := conn.PullRequests.Create(orgName, repo, input)
		if err != nil {
			return err
		}

		// add labels to pull request
		labels := []string{"label-a", "label-b"}
		_, _, err = conn.Issues.AddLabelsToIssue(orgName, repo, *pull.Number, labels)
		if err != nil {
			return err
		}

		return nil
	}
}

const testAccDataSourceGithubPullRequestsConfig = `
resource "github_repository" "foo" {
  name = "foo"
  description = "Terraform acceptance tests"
  homepage_url = "http://example.com/"

  # So that acceptance tests can be run in a github organization
  # with no billing
  private = false

  has_issues = true
  has_wiki = true
  has_downloads = true

	auto_init = true
}
`

const testAccDataSourceGithubPullRequestsDataConfigMatched = testAccDataSourceGithubPullRequestsConfig + "\n" + `
data "github_pull_requests" "foo" {
	repository = "foo"
	label_regexp = "\\Alabel\\-(a|b)\\z"
	title_regexp = "\\ATest\\z"
}
`
const testAccDataSourceGithubPullRequestsDataConfigNotMatchedLabel = testAccDataSourceGithubPullRequestsConfig + "\n" + `
data "github_pull_requests" "foo" {
	repository = "foo"
	label_regexp = "notFound"
}
`
const testAccDataSourceGithubPullRequestsDataConfigNotMatchedTitle = testAccDataSourceGithubPullRequestsConfig + "\n" + `
data "github_pull_requests" "foo" {
	repository = "foo"
	title_regexp = "notFound"
}
`
