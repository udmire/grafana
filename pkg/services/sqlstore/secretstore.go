package sqlstore

import (
	"context"
	"strings"
	"time"

	"github.com/grafana/grafana/pkg/events"
	"github.com/grafana/grafana/pkg/models"
	"xorm.io/xorm"
)

// GetSecret adds a secret to the query model by querying by org_id as well as
// either uid (preferred), id, or name and is added to the bus.
func (ss *SQLStore) GetSecret(ctx context.Context, query *models.GetSecretQuery) error {
	// metrics.MDBSecretQueryByID.Inc()

	return ss.WithDbSession(ctx, func(sess *DBSession) error {
		if query.OrgId == 0 || (query.Id == 0 && len(query.EntityUid) == 0) {
			return models.ErrSecretIdentifierNotSet
		}

		secret := &models.Secret{EntityUid: query.EntityUid, OrgId: query.OrgId, Id: query.Id}
		has, err := sess.Get(secret)

		if err != nil {
			sqlog.Error("Failed getting secret", "err", err, "uid", query.EntityUid, "id", query.Id, "orgId", query.OrgId)
			return err
		} else if !has {
			return models.ErrSecretNotFound
		}

		query.Result = secret

		return nil
	})
}

func (ss *SQLStore) GetSecrets(ctx context.Context, query *models.GetSecretsQuery) error {
	var sess *xorm.Session
	return ss.WithDbSession(ctx, func(dbSess *DBSession) error {
		if query.SecretLimit <= 0 {
			sess = dbSess.Where("org_id=?", query.OrgId).Asc("id")
		} else {
			sess = dbSess.Limit(query.SecretLimit, 0).Where("org_id=?", query.OrgId).Asc("id")
		}

		query.Result = make([]*models.Secret, 0)
		return sess.Find(&query.Result)
	})
}

// DeleteSecret removes a secret by org_id as well as either uid (preferred), id, or name
// and is added to the bus.
func (ss *SQLStore) DeleteSecret(ctx context.Context, cmd *models.DeleteSecretCommand) error {
	params := make([]interface{}, 0)

	makeQuery := func(sql string, p ...interface{}) {
		params = append(params, sql)
		params = append(params, p...)
	}

	switch {
	case cmd.OrgID == 0:
		return models.ErrSecretIdentifierNotSet
	case cmd.ID != 0:
		makeQuery("DELETE FROM secrets WHERE id=? and org_id=?", cmd.ID, cmd.OrgID)
	case cmd.EntityUID != "":
		makeQuery("DELETE FROM secrets WHERE entity_uid=? and org_id=?", cmd.EntityUID, cmd.OrgID)
	default:
		return models.ErrSecretIdentifierNotSet
	}

	return ss.WithTransactionalDbSession(ctx, func(sess *DBSession) error {
		result, err := sess.Exec(params...)
		cmd.DeletedSecretsCount, _ = result.RowsAffected()

		sess.publishAfterCommit(&events.SecretDeleted{
			Timestamp: time.Now(),
			EntityUID: cmd.EntityUID,
			ID:        cmd.ID,
			OrgID:     cmd.OrgID,
		})

		return err
	})
}

func (ss *SQLStore) AddSecret(ctx context.Context, cmd *models.AddSecretCommand) error {
	return ss.WithTransactionalDbSession(ctx, func(sess *DBSession) error {
		existing := models.Secret{OrgId: cmd.OrgId, EntityUid: cmd.EntityUid}
		has, _ := sess.Get(&existing)

		if has {
			return models.ErrSecretEntityUidExists
		}

		s := &models.Secret{
			OrgId:          cmd.OrgId,
			EntityUid:      cmd.EntityUid,
			SecureJsonData: cmd.EncryptedSecureJsonData,
			Created:        time.Now(),
			Updated:        time.Now(),
		}

		if _, err := sess.Insert(s); err != nil {
			if dialect.IsUniqueConstraintViolation(err) && strings.Contains(strings.ToLower(dialect.ErrorMessage(err)), "entity_uid") {
				return models.ErrSecretEntityUidExists
			}
			return err
		}

		cmd.Result = s

		sess.publishAfterCommit(&events.SecretCreated{
			Timestamp: time.Now(),
			EntityUID: cmd.EntityUid,
			ID:        s.Id,
			OrgID:     cmd.OrgId,
		})
		return nil
	})
}

func (ss *SQLStore) UpdateSecret(ctx context.Context, cmd *models.UpdateSecretCommand) error {
	return ss.WithTransactionalDbSession(ctx, func(sess *DBSession) error {
		s := &models.Secret{
			Id:             cmd.Id,
			OrgId:          cmd.OrgId,
			EntityUid:      cmd.EntityUid,
			SecureJsonData: cmd.EncryptedSecureJsonData,
			Updated:        time.Now(),
		}

		var updateSession *xorm.Session
		affected, err := updateSession.Update(s)
		if err != nil {
			return err
		}

		if affected == 0 {
			return models.ErrSecretNotFound
		}

		cmd.Result = s
		return err
	})
}
