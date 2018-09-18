package aws

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/aws/aws-sdk-go/service/kms/kmsiface"
	"github.com/chanzuckerberg/aws-tidy/pkg/policy"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// kms specific labels
const (
	labelKMSKeyDescription TypeEntityLabel = "key_description"
	labelKMSKeyState       TypeEntityLabel = "key_state"
)

type kmsKey struct {
	Entity
	keyID string
}

// Delete deletes this kms key
func (k *kmsKey) Delete() error {
	log.Warnf("Would delete KMS key %s", k.keyID)
	return nil
}

func (k *kmsKey) GetID() string {
	return fmt.Sprintf("kms: %s", k.keyID)
}

func newKMSKeyEntity(keyID string) *kmsKey {
	return &kmsKey{keyID: keyID}
}

// KMSClient is a kms client
type KMSClient struct {
	Svc kmsiface.KMSAPI
}

// NewKMS returns a KMS client
func NewKMS(s *session.Session) *KMSClient {
	return &KMSClient{kms.New(s)}
}

// Walk walks through all kms keys and applies the walkFun
func (k *KMSClient) Walk(p policy.Policy) error {
	input := &kms.ListKeysInput{}
	keyIDs := []string{}
	err := k.Svc.ListKeysPages(input, func(output *kms.ListKeysOutput, lastPage bool) bool {
		for _, kmsKeyListEntry := range output.Keys {
			if kmsKeyListEntry != nil && kmsKeyListEntry.KeyId != nil {
				keyIDs = append(keyIDs, *kmsKeyListEntry.KeyId)
			}
		}
		return true
	})
	if err != nil {
		return errors.Wrap(err, "Error listing KMS keys")
	}

	for _, keyID := range keyIDs {
		keyMetadata, err := k.DescribeKey(keyID)

		if err != nil {
			return errors.Wrapf(err, "error describing kms key %s", keyID)
		}
		entity := newKMSKeyEntity(keyID)
		entity.
			WithLabel(labelARN, keyMetadata.Arn).
			WithLabel(labelID, keyMetadata.KeyId).
			WithLabel(labelKMSKeyDescription, keyMetadata.Description).
			WithLabel(labelKMSKeyState, keyMetadata.KeyState).
			WithCreatedAt(keyMetadata.CreationDate)

		// Need to not fail on err here
		_, err = p.Eval(entity)
		if err != nil {
			return err
		}
	}
	return nil
}

// DescribeKey describes a key
func (k *KMSClient) DescribeKey(keyID string) (*kms.KeyMetadata, error) {
	input := &kms.DescribeKeyInput{}
	input.SetKeyId(keyID)
	output, err := k.Svc.DescribeKey(input)
	return output.KeyMetadata, errors.Wrapf(err, "Could not describe kms key %s", keyID)
}
