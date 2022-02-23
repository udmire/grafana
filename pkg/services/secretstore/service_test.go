package secretstore

import (
	"context"
	"testing"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/secrets/database"
	secretsManager "github.com/grafana/grafana/pkg/services/secrets/manager"
	"github.com/grafana/grafana/pkg/services/sqlstore"
	"github.com/grafana/grafana/pkg/setting"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	sqlStore := sqlstore.InitTestDB(t)

	origSecret := setting.SecretKey
	setting.SecretKey = "datasources_service_test"
	t.Cleanup(func() {
		setting.SecretKey = origSecret
	})

	secretsService := secretsManager.SetupTestService(t, database.ProvideSecretsStore(sqlStore))
	s := ProvideService(bus.New(), sqlStore, secretsService)

	var ds *models.Secret

	t.Run("create datasource should encrypt the secure json data", func(t *testing.T) {
		ctx := context.Background()

		sjd := map[string]string{"password": "12345"}
		cmd := models.AddSecretCommand{SecureJsonData: sjd}

		err := s.AddSecret(ctx, &cmd)
		require.NoError(t, err)

		ds = cmd.Result
		decrypted, err := s.SecretsService.DecryptJsonData(ctx, ds.SecureJsonData)
		require.NoError(t, err)
		require.Equal(t, sjd, decrypted)
	})

	t.Run("update datasource should encrypt the secure json data", func(t *testing.T) {
		ctx := context.Background()
		sjd := map[string]string{"password": "678910"}
		cmd := models.UpdateSecretCommand{Id: ds.Id, OrgId: ds.OrgId, SecureJsonData: sjd}
		err := s.UpdateSecret(ctx, &cmd)
		require.NoError(t, err)

		decrypted, err := s.SecretsService.DecryptJsonData(ctx, cmd.Result.SecureJsonData)
		require.NoError(t, err)
		require.Equal(t, sjd, decrypted)
	})
}
