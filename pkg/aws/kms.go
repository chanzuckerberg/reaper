package aws

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/service/kms"
	cziAws "github.com/chanzuckerberg/go-misc/aws"
	"github.com/chanzuckerberg/reaper/pkg/policy"
	multierror "github.com/hashicorp/go-multierror"
	log "github.com/sirupsen/logrus"
)

// kms specific labels
const (
	labelKMSKeyDescription TypeEntityLabel = "key_description"
	labelKMSKeyState       TypeEntityLabel = "key_state"
)

type KmsKey struct {
	Entity
	keyID string
}

// Delete deletes this kms key
func (k *KmsKey) Delete() error {
	log.Warnf("Would delete KMS key %s", k.keyID)
	return nil
}

func (k *KmsKey) GetID() string {
	return fmt.Sprintf("kms: %s", k.keyID)
}
func (k *KmsKey) GetConsoleURL() string {
	s := "https://console.aws.amazon.com/iam/home?region=%s#/encryptionKeys/%s/%s"
	return fmt.Sprintf(s, k.Region, k.Region, k.ID)
}

func NewKMSKey(keyMetadata *kms.KeyMetadata, tags []*kms.Tag) *KmsKey {
	entity := &KmsKey{keyID: *keyMetadata.KeyId}
	entity.
		AddLabel(labelARN, keyMetadata.Arn).
		AddLabel(labelID, keyMetadata.KeyId).
		AddLabel(labelKMSKeyDescription, keyMetadata.Description).
		AddLabel(labelKMSKeyState, keyMetadata.KeyState).
		AddCreatedAt(keyMetadata.CreationDate)

	for _, tag := range tags {
		if tag == nil {
			continue
		}
		if tag.TagKey != nil && tag.TagValue != nil && *tag.TagKey == "Name" {
			entity.Name = *tag.TagValue
		}
		entity.AddTag(tag.TagKey, tag.TagValue)
	}

	return entity
}

// EvalKMSKey walks through all kms keys
func (c *Client) EvalKMSKey(accounts []*policy.Account, p policy.Policy, regions []string, f func(policy.Violation)) error {
	var errs error
	ctx := context.Background()
	err := c.WalkAccountsAndRegions(accounts, regions, func(client *cziAws.Client, account *policy.Account, region string) {
		input := &kms.ListKeysInput{}
		err := client.KMS.Svc.ListKeysPagesWithContext(ctx, input, func(output *kms.ListKeysOutput, done bool) bool {
			for _, key := range output.Keys {
				if key != nil && key.KeyId != nil {
					input := &kms.DescribeKeyInput{}
					input.SetKeyId(*key.KeyId)
					output, _ := client.KMS.Svc.DescribeKey(input)
					keyMetadata := output.KeyMetadata

					tagsInput := &kms.ListResourceTagsInput{KeyId: keyMetadata.KeyId}
					tagsOutput, _ := client.KMS.Svc.ListResourceTags(tagsInput)
					tags := tagsOutput.Tags

					k := NewKMSKey(keyMetadata, tags)
					if p.Match(k) {
						violation := policy.NewViolation(p, k, false, account)
						f(violation)
					}
				}
			}
			return true
		})
		errs = multierror.Append(errs, err)
	})
	errs = multierror.Append(errs, err)

	return errs
}
