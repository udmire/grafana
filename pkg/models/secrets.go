package models

import (
	"errors"
	"time"
)

var (
	ErrSecretNotFound         = errors.New("secret not found")
	ErrSecretEntityUidExists  = errors.New("secret with the same entity_uid and org_id already exists")
	ErrSecretIdentifierNotSet = errors.New("unique identifier and org id are needed to be able to get or delete a secret")
)

type Secret struct {
	Id        int64  `json:"id"`
	OrgId     int64  `json:"orgId"`
	EntityUid string `json:"entityUid"`

	SecureJsonData map[string][]byte `json:"secureJsonData"`

	Created time.Time `json:"created"`
	Updated time.Time `json:"updated"`
}

// ----------------------
// COMMANDS

// Also acts as api DTO
type AddSecretCommand struct {
	EntityUid      string            `json:"entityUid" binding:"Required"`
	SecureJsonData map[string]string `json:"secureJsonData" binding:"Required"`

	OrgId                   int64             `json:"-"`
	EncryptedSecureJsonData map[string][]byte `json:"-"`

	Result *Secret `json:"-"`
}

// Also acts as api DTO
type UpdateSecretCommand struct {
	EntityUid      string            `json:"entityUid" binding:"Required"`
	SecureJsonData map[string]string `json:"secureJsonData" binding:"Required"`

	OrgId                   int64             `json:"-"`
	Id                      int64             `json:"-"`
	EncryptedSecureJsonData map[string][]byte `json:"-"`

	Result *Secret `json:"-"`
}

// DeleteSecretCommand will delete a Secret based on OrgID as well as the UID (preferred), ID, or Name.
// At least one of the UID, ID, or Name properties must be set in addition to OrgID.
type DeleteSecretCommand struct {
	ID        int64
	OrgID     int64
	EntityUID string

	DeletedSecretsCount int64
}

// ---------------------
// QUERIES

type GetSecretsQuery struct {
	OrgId       int64
	SecretLimit int
	User        *SignedInUser
	Result      []*Secret
}

// GetSecretQuery will get a Secret based on OrgID as well as the UID (preferred), ID, or Name.
// At least one of the UID, ID, or Name properties must be set in addition to OrgID.
type GetSecretQuery struct {
	Id int64

	OrgId     int64
	EntityUid string

	Result *Secret
}

// ---------------------
//  Permissions
// ---------------------

type SecretsPermissionFilterQuery struct {
	User    *SignedInUser
	Secrets []*Secret
	Result  []*Secret
}
