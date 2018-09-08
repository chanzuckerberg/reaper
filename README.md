AWS tidy can be used to enforce policies on AWS resource tagging, etc

Some example policies:

```yaml
policies:
  # Give me all taggable resources that have a managedBy tag
  - resource_selector: "" # select amongst aws services
    tag_selector: "managedBy" # select amongst aws resources tags
    label_selector: "" # select amongst resource-specific tags (generated internally)

  # Give me all taggable resources that do not have a managedBy tag
  - resource_selector: "" # select amongst aws services
    tag_selector: "!managedBy" # select amongst aws resources tags
    label_selector: "" # select amongst resource-specific tags (generated internally)

  # Give me all s3 buckets that do not have a managedBy tag
  - resource_selector: "name=s3" # select amongst aws services
    tag_selector: "!managedBy" # select amongst aws resources tags
    label_selector: "" # select amongst resource-specific tags (generated internally)

  # Give me all s3 buckets and rds instances that do not have a managedBy tag
  - resource_selector: "name in (s3,rds)" # select amongst aws services
    tag_selector: "!managedBy" # select amongst aws resources tags
    label_selector: "" # select amongst resource-specific tags (generated internally)

  # Give me all taggable resources that don't have managedBy nor env nor owner nor service tags
  - resource_selector: "" # select amongst aws services
    tag_selector: "!managedBy,!env,!owner,!service" # select amongst aws resources tags
    label_selector: "" # select amongst resource-specific tags (generated internally)

  # Give me all public s3 buckets
  - resource_selector: "name=s3" # select amongst aws services
    tag_selector: "" # select amongst aws resources tags
    label_selector: "s3_is_public" # select amongst resource-specific tags (generated internally)

  # Give me all private s3 buckets
  - resource_selector: "name=s3" # select amongst aws services
    tag_selector: "" # select amongst aws resources tags
    label_selector: "!s3_is_public" # select amongst resource-specific tags (generated internally to this project)

```
