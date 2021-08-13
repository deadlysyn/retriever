// Package retriever provides primitives for interacting with
// AWS Parameter Store and Secrets Manager.
package retriever

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
)

func init() {
	log.SetFlags(log.Lshortfile)

	viper.SetConfigName("retriever")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")

	viper.SetEnvPrefix("RTVR")
	viper.AutomaticEnv()

	cfg := viper.GetString("conf")
	if cfg != "" {
		viper.SetConfigFile(cfg)
	}

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("INFO: retriever config not found; using environment")
		} else {
			log.Fatalf("ERROR: %v", err.Error())
		}
	}
}

// getParam retrieves the specified secret from Parameter Store.
func getParam(ctx context.Context, c *ssm.Client, p string) (*ssm.GetParameterOutput, error) {
	i := ssm.GetParameterInput{
		Name:           aws.String(p),
		WithDecryption: true,
	}

	out, err := c.GetParameter(ctx, &i)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// getSecret retrieves the specified secret from Secrets Manager.
func getSecret(ctx context.Context, c *secretsmanager.Client, p string) (*secretsmanager.GetSecretValueOutput, error) {
	i := secretsmanager.GetSecretValueInput{
		SecretId: aws.String(p),
	}

	out, err := c.GetSecretValue(ctx, &i)
	if err != nil {
		return nil, err
	}

	return out, nil
}

// Fetch retrieves secrets specified via configuration or environment
// from Parameter Store or Secrets Manager. It returns map[string]string
// with secret names as keys.
func Fetch() (map[string]string, error) {
	creds := make(map[string]string)
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("ERROR: unable to load AWS configuration: %v", err.Error())
	}

	p := viper.GetString("prefix")
	t := strings.ToLower(viper.GetString("type"))
	if t == "parameter" {
		client := ssm.NewFromConfig(cfg)
		for _, v := range viper.GetStringSlice("credentials") {
			res, err := getParam(ctx, client, fmt.Sprintf("%s/%s", p, v))
			if err != nil {
				return nil, fmt.Errorf("ERROR: unable to retrieve %v/%v (%v)", p, v, err.Error())
			}
			creds[v] = aws.ToString(res.Parameter.Value)
		}
	} else if t == "secret" {
		client := secretsmanager.NewFromConfig(cfg)
		for _, v := range viper.GetStringSlice("credentials") {
			res, err := getSecret(ctx, client, fmt.Sprintf("%s/%s", p, v))
			if err != nil {
				return nil, fmt.Errorf("ERROR: unable to retrieve %v/%v (%v)", p, v, err.Error())
			}
			creds[v] = aws.ToString(res.SecretString)
		}
	} else {
		return nil, fmt.Errorf("ERROR: unknown secret type \"%v\"", t)
	}

	return creds, nil
}
