package secretstore

import (
	"context"
	"sync"
	"time"

	"github.com/grafana/grafana/pkg/bus"
	"github.com/grafana/grafana/pkg/models"
	"github.com/grafana/grafana/pkg/services/secrets"
	"github.com/grafana/grafana/pkg/services/sqlstore"
)

type Service struct {
	Bus            bus.Bus
	SQLStore       *sqlstore.SQLStore
	SecretsService secrets.Service

	scDecryptionCache secureJSONDecryptionCache
}

type secureJSONDecryptionCache struct {
	cache map[int64]cachedDecryptedJSON
	sync.Mutex
}

type cachedDecryptedJSON struct {
	updated time.Time
	json    map[string]string
}

func ProvideService(bus bus.Bus, store *sqlstore.SQLStore, secretsService secrets.Service) *Service {
	s := &Service{
		Bus:            bus,
		SQLStore:       store,
		SecretsService: secretsService,
		scDecryptionCache: secureJSONDecryptionCache{
			cache: make(map[int64]cachedDecryptedJSON),
		},
	}

	s.Bus.AddHandler(s.GetSecrets)
	s.Bus.AddHandler(s.GetSecret)
	s.Bus.AddHandler(s.AddSecret)
	s.Bus.AddHandler(s.DeleteSecret)
	s.Bus.AddHandler(s.UpdateSecret)

	return s
}

type SecretRetriever interface {
	GetSecret(ctx context.Context, query *models.GetSecretQuery) error
}

func (s *Service) GetSecret(ctx context.Context, query *models.GetSecretQuery) error {
	return s.SQLStore.GetSecret(ctx, query)
}

func (s *Service) GetSecrets(ctx context.Context, query *models.GetSecretsQuery) error {
	return s.SQLStore.GetSecrets(ctx, query)
}

func (s *Service) AddSecret(ctx context.Context, cmd *models.AddSecretCommand) error {
	var err error
	cmd.EncryptedSecureJsonData, err = s.SecretsService.EncryptJsonData(ctx, cmd.SecureJsonData, secrets.WithoutScope())
	if err != nil {
		return err
	}

	return s.SQLStore.AddSecret(ctx, cmd)
}

func (s *Service) DeleteSecret(ctx context.Context, cmd *models.DeleteSecretCommand) error {
	return s.SQLStore.DeleteSecret(ctx, cmd)
}

func (s *Service) UpdateSecret(ctx context.Context, cmd *models.UpdateSecretCommand) error {
	var err error
	cmd.EncryptedSecureJsonData, err = s.SecretsService.EncryptJsonData(ctx, cmd.SecureJsonData, secrets.WithoutScope())
	if err != nil {
		return err
	}

	return s.SQLStore.UpdateSecret(ctx, cmd)
}

func (s *Service) DecryptedValues(sc *models.Secret) map[string]string {
	s.scDecryptionCache.Lock()
	defer s.scDecryptionCache.Unlock()

	if item, present := s.scDecryptionCache.cache[sc.Id]; present && sc.Updated.Equal(item.updated) {
		return item.json
	}

	json, err := s.SecretsService.DecryptJsonData(context.Background(), sc.SecureJsonData)
	if err != nil {
		return map[string]string{}
	}

	s.scDecryptionCache.cache[sc.Id] = cachedDecryptedJSON{
		updated: sc.Updated,
		json:    json,
	}

	return json
}
