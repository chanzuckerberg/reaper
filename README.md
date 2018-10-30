# Reaper
Reaper is a tool for cleaning up cloud resources. You might want to clean them up based on policy non-compliance, idleness or after running tests.

## Example

Below is an example config file.

```yaml
# version is a required field, the only valid version is 1
version: 1

# accounts lists the AWS accounts you would like to scan
accounts:
    # name is used to provide useful output (more useful than an id)
  - name: hub
    # id is the AWS account id. We need this to assume a role in the account.
    id: 123456789
    #  role is the name of the role to assume
    role: reaper
    # owner is the default owner for resources in this account.
    owner: infra@example.com
    #  you can list multiple accounts and we will scan all of them
  - name: other
    id: 987654321
    role: reaper
    owner: infra@example.com

# identity_map maps from email -> slack identities for cases where we can't figure it out
# this is most useful for email lists and slack channels. For users we can look up the
# slack user based on their email address.
identity_map:
  - email: infra@example.com
    slack: infra-ops

# aws_regions lists the regions we want to scan
aws_regions:
  - us-east-1
  - us-west-1
  - us-west-2

# policies is a list of polices we'd like to enforce
policies:
  - name: owner-ec2
    # resource_selector specifies which resources this policy applies to
    resource_selector: "name in (ec2_instance)"
    # tag_selector selects resources based on tags
    # in this case it is saying 'resources without an owner tag'
    tag_selector: "!owner"
    # label_selector selects resources based on other attributes of the resource
    # these are resource specific (and not well documented)
    label_selector: ""

    # notifications lists the notifcations you want to send for resources that match the policy
    notifications:
      # recipient can either be an email address (in which case we will look up the slack identity),
      # or $owner, in which we will calculate the owner for this resource
      - recipient: $owner
        #  message_template specifies how to message the recipient
        message_template: >
          *WARNING*– EC2 Instance <{{.Resource.GetConsoleURL}}|{{.ResourceID}}> in account
          `{{.AccountName}}` does not have an owner tag. See our <https://example.com/cloud-policy|Usage Policy> for more information.
```