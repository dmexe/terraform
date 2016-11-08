package github

import (
	"fmt"
	"github.com/google/go-github/github"
	"github.com/hashicorp/terraform/helper/schema"
	"hash/fnv"
	"regexp"
	"strconv"
)

func dataSourceGithubPullRequests() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceGithubPullRequestsRead,

		Schema: map[string]*schema.Schema{
			"repository": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"state": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "open",
				ForceNew: true,
			},
			"label_regexp": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp,
			},
			"title_regexp": &schema.Schema{
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validateRegexp,
			},

			"pulls": &schema.Schema{
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"number": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"title": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"issue_labels": &schema.Schema{
							Type:     schema.TypeList,
							Computed: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"user_login": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"head_label": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"head_sha": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"head_ref": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"head_repo_name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"base_label": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"base_sha": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"base_ref": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
						"base_repo_name": &schema.Schema{
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceGithubPullRequestsRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*Organization).client
	orgname := meta.(*Organization).name
	repo := d.Get("repository").(string)
	state := d.Get("state").(string)
	id := fnv.New32a()

	opts := &github.PullRequestListOptions{
		State:       state,
		ListOptions: github.ListOptions{PerPage: 30},
	}

	pulls := make([]map[string]interface{}, 0)

	for {
		found, resp, err := client.PullRequests.List(orgname, repo, opts)
		if err != nil {
			return err
		}

		for _, re := range found {
			labels, err := dataSourceGithubPullRequestsGetLabels(client, orgname, repo, *re.Number)
			if err != nil {
				return err
			}

			if dataSourceGithubPullRequestsFilterLabels(d, labels) == false {
				continue
			}

			if dataSourceGithubPullRequestsFilterTitle(d, *re.Title) == false {
				continue
			}

			number := strconv.Itoa(*re.Number)

			pull := map[string]interface{}{
				"number":         number,
				"state":          *re.State,
				"title":          *re.Title,
				"user_login":     *re.User.Login,
				"head_label":     *re.Head.Label,
				"head_ref":       *re.Head.Ref,
				"head_sha":       *re.Head.SHA,
				"head_repo_name": *re.Head.Repo.Name,
				"base_label":     *re.Base.Label,
				"base_ref":       *re.Base.Ref,
				"base_sha":       *re.Base.SHA,
				"base_repo_name": *re.Base.Repo.Name,
				"issue_labels":   labels,
			}

			pulls = append(pulls, pull)

			id.Write([]byte(number))
		}

		if resp.NextPage == 0 {
			break
		} else {
			opts.Page = resp.NextPage
		}
	}

	if err := d.Set("pulls", pulls); err != nil {
		return err
	}

	d.SetId(fmt.Sprint(id.Sum32()))

	return nil
}

func dataSourceGithubPullRequestsFilterLabels(d *schema.ResourceData, labels []string) bool {
	re, ok := d.GetOk("label_regexp")

	if ok == false {
		return true
	}

	r := regexp.MustCompile(re.(string))
	for _, label := range labels {
		if r.MatchString(label) {
			return true
		}
	}

	return false
}

func dataSourceGithubPullRequestsFilterTitle(d *schema.ResourceData, title string) bool {
	re, ok := d.GetOk("title_regexp")

	if ok == false {
		return true
	}

	r := regexp.MustCompile(re.(string))
	if r.MatchString(title) {
		return true
	}

	return false
}

func dataSourceGithubPullRequestsGetLabels(client *github.Client, orgname string, repo string, issue int) ([]string, error) {
	labels := make([]string, 0)
	opts := &github.ListOptions{
		PerPage: 30,
	}

	for {
		reply, resp, err := client.Issues.ListLabelsByIssue(orgname, repo, issue, opts)
		if err != nil {
			return nil, err
		}

		for _, label := range reply {
			labels = append(labels, *label.Name)
		}

		if resp.NextPage == 0 {
			break
		} else {
			opts.Page = resp.NextPage
		}
	}

	return labels, nil
}
