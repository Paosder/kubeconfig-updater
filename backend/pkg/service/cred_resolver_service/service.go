package cred_resolver_service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure/auth"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/pubg/kubeconfig-updater/backend/controller/protos"
	"github.com/pubg/kubeconfig-updater/backend/pkg/persistence/cred_resolver_config_persist"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
)

type CredResolverService struct {
	store *cred_resolver_config_persist.CredResolverConfigStorage
}

func NewCredResolverService(store *cred_resolver_config_persist.CredResolverConfigStorage) *CredResolverService {
	return &CredResolverService{store: store}
}

func (s *CredResolverService) ListCredResolvers() []*protos.CredResolverConfig {
	return s.store.ListConfigs()
}

func (s *CredResolverService) GetCredResolver(credResolverId string) (*protos.CredResolverConfig, bool, error) {
	if credResolverId == "" {
		return nil, false, fmt.Errorf("credResolverId should not be empty")
	}

	cfg, exists := s.store.GetConfig(credResolverId)
	if !exists {
		return nil, false, nil
	}
	return cfg, true, nil
}

func (s *CredResolverService) SetCredResolver(cfg *protos.CredResolverConfig) error {
	if cfg == nil {
		return fmt.Errorf("credResolverConfig should not be null")
	}

	return s.store.SetAndSaveConfig(cfg)
}

func (s *CredResolverService) DeleteCredResolver(credResolverId string) error {
	return s.store.DeleteConfig(credResolverId)
}

const attribute_profile = "profile"

func (s *CredResolverService) GetAwsSdkConfig(ctx context.Context, credConf *protos.CredResolverConfig) (*aws.Config, string, error) {
	switch credConf.GetKind() {
	case protos.CredentialResolverKind_DEFAULT:
		cfg, err := config.LoadDefaultConfig(ctx)
		return &cfg, "", err
	case protos.CredentialResolverKind_ENV:
		cfg, err := config.LoadDefaultConfig(ctx)
		return &cfg, "", err
	case protos.CredentialResolverKind_IMDS:
		cfg, err := config.LoadDefaultConfig(ctx)
		return &cfg, "", err
	case protos.CredentialResolverKind_PROFILE:
		attributes := credConf.GetResolverAttributes()
		if attributes == nil {
			return nil, "", fmt.Errorf("attribute should not null")
		}
		profile, exists := attributes[attribute_profile]
		if !exists {
			return nil, "", fmt.Errorf("profile attribute should be exist")
		}
		cfg, err := config.LoadDefaultConfig(ctx, config.WithSharedConfigProfile(profile))
		return &cfg, profile, err
	default:
		return nil, "", fmt.Errorf("unknown kind value %s", credConf.GetKind())
	}
}

// Azure Cli는 멀티어카운트를 지원하지 않는다
func (s *CredResolverService) GetAzureSdkConfig(ctx context.Context, credConf *protos.CredResolverConfig) (autorest.Authorizer, error) {
	switch credConf.GetKind() {
	case protos.CredentialResolverKind_DEFAULT:
		return auth.NewAuthorizerFromCLI()
	case protos.CredentialResolverKind_ENV:
		return auth.NewAuthorizerFromEnvironment()
	case protos.CredentialResolverKind_IMDS:
		return auth.NewMSIConfig().Authorizer()
	case protos.CredentialResolverKind_PROFILE:
		return nil, fmt.Errorf("credentialType=PROFILE is not support for azure credResolver")
	default:
		return nil, fmt.Errorf("unknown kind value %s", credConf.GetKind())
	}
}

func (s *CredResolverService) GetTencentSdkConfig(credConf *protos.CredResolverConfig) (common.Provider, error) {
	switch credConf.GetKind() {
	case protos.CredentialResolverKind_DEFAULT:
		return common.DefaultProviderChain(), nil
	case protos.CredentialResolverKind_ENV:
		return common.DefaultEnvProvider(), nil
	case protos.CredentialResolverKind_IMDS:
		return common.DefaultCvmRoleProvider(), nil
	case protos.CredentialResolverKind_PROFILE:
		attributes := credConf.GetResolverAttributes()
		if attributes == nil {
			return nil, fmt.Errorf("attribute should not null")
		}
		profile, exists := attributes[attribute_profile]
		if !exists {
			return nil, fmt.Errorf("profile attribute should be exist")
		}
		return &TencentIntlProfileProvider{profileName: profile}, nil
	default:
		return nil, fmt.Errorf("unknown kind value %s", credConf.GetKind())
	}
}

// 중국 본토랑 intl이랑 credentials 포맷이 다름
type TencentIntlProfileProvider struct {
	profileName string
}

func (t *TencentIntlProfileProvider) GetCredential() (common.CredentialIface, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	credentialFilePath := filepath.Join(home, ".tccli", t.profileName+".credential")

	buf, err := os.ReadFile(credentialFilePath)
	if err != nil {
		return nil, err
	}

	rawCred := map[string]interface{}{}
	err = json.Unmarshal(buf, &rawCred)
	if err != nil {
		return nil, err
	}

	return &common.Credential{
		SecretId:  rawCred["secretId"].(string),
		SecretKey: rawCred["secretKey"].(string),
	}, nil
}
