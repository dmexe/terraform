
---
layout: "github"
page_title: "GitHub: pull_requests"
sidebar_current: "docs-github-data-pull-requests"
description: |-
  Provides details about a specific pull requests
---

# github\_pull_requests

Provides a Data Source to access to Github pull requests.

## Example Usage

```
data "github_pull_requests" "staging" {
  repository   = "example"
  label_regexp = "staging"
}
```

## Argument Reference

The following arguments are supported:

* `repository` - (Required) The name of the repository.

* `state` - (Optional) A state of pull request, either `open`, `closed`, or `all` to filter by state. Default: `open`.

* `label_regexp` - (Optional)  A regex string to apply to the pull request labels.

* `title_regexp` - (Optional)  A regex string to apply to the pull request title.

## Attributes Reference

The following additional attributes are exported:

* `pulls` - A list of macthed pull requests.

Each `pulls` supports the following:

* `number` - A number of a pull request.

* `state` - A state of a pull request, either `open`, `closed`.

* `title` - A title of a pull request.

* `issue_labels` - A list of labels attached to an associated issue.

* `user_login` - A login of a user who has created pull request

* `head_label`, `head_sha`, `head_ref`, `head_repo_name` - fields from head branch in a pull request.

* `base_label`, `base_sha`, `base_ref`, `base_repo_name` - fields from base branch in a pull request.
