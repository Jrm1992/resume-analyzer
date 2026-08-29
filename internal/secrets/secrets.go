package secrets

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

const SecretName = "resume-analyzer/config"

func Load(ctx context.Context, secretName string) (map[string]string, error) {
	endpoint := os.Getenv("AWS_ENDPOINT_URL")
	if endpoint == "" {
		return nil, nil
	}

	region := os.Getenv("AWS_DEFAULT_REGION")
	if region == "" {
		region = "us-east-1"
	}

	akid := os.Getenv("AWS_ACCESS_KEY_ID")
	secretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(akid, secretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("secrets: aws config: %w", err)
	}
	client := secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = &endpoint
	})

	out, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: &secretName,
	})
	if err != nil {
		return nil, fmt.Errorf("secrets: get %q: %w", secretName, err)
	}

	m, err := ParseEnvMap(*out.SecretString)
	if err != nil {
		return nil, fmt.Errorf("secrets: parse %q: %w", secretName, err)
	}

	slog.Info("secrets: loaded from Secrets Manager", "name", secretName, "keys", len(m))
	return m, nil
}

// ParseEnvMap unmarshals a secrets JSON document into env-style string values.
// Scalars (numbers, booleans) are coerced to their string form; null values
// are skipped; nested objects and arrays are rejected since they are not
// representable as env vars.
func ParseEnvMap(raw string) (map[string]string, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		switch val := v.(type) {
		case string:
			out[k] = val
		case float64, bool:
			b, err := json.Marshal(val)
			if err != nil {
				return nil, fmt.Errorf("key %q: %w", k, err)
			}
			out[k] = string(b)
		case nil:
			continue
		default:
			return nil, fmt.Errorf("key %q: value must be a string, number, or boolean", k)
		}
	}
	return out, nil
}

func SetIfAbsent(m map[string]string) {
	for k, v := range m {
		if os.Getenv(k) == "" {
			_ = os.Setenv(k, v)
		}
	}
}
