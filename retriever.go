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

func configure() (*viper.Viper, error) {
	rtvrViper := viper.New()

	rtvrViper.SetConfigName("retriever")
	rtvrViper.SetConfigType("yaml")
	rtvrViper.AddConfigPath(".")

	rtvrViper.SetEnvPrefix("RTVR")
	rtvrViper.AutomaticEnv()
	cfg := rtvrViper.GetString("conf")
	if cfg != "" {
		rtvrViper.SetConfigFile(cfg)
	}

	if err := rtvrViper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Println("retriever config not found; using environment")
		} else {
			return nil, err
		}
	}

	return rtvrViper, nil
}

// Fetch retrieves secrets specified via configuration or environment
// from Parameter Store or Secrets Manager and returns a map with
// secret names as keys.
func Fetch() (map[string]string, error) {
	creds := make(map[string]string)
	ctx := context.TODO()

	v, err := configure()
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("ERROR: unable to load AWS configuration: %v", err.Error())
	}

	p := v.GetString("prefix")
	t := strings.ToLower(v.GetString("type"))
	if t == "parameter" {
		client := ssm.NewFromConfig(cfg)
		for _, v := range v.GetStringSlice("credentials") {
			res, err := getParam(ctx, client, fmt.Sprintf("%s/%s", p, v))
			if err != nil {
				return nil, fmt.Errorf("ERROR: unable to retrieve %v/%v (%v)", p, v, err.Error())
			}
			creds[v] = aws.ToString(res.Parameter.Value)
		}
	} else if t == "secret" {
		client := secretsmanager.NewFromConfig(cfg)
		for _, v := range v.GetStringSlice("credentials") {
			res, err := getSecret(ctx, client, fmt.Sprintf("%s/%s", p, v))
			if err != nil {
				return nil, fmt.Errorf("ERROR: unable to retrieve %v/%v (%v)", p, v, err.Error())
			}
			creds[v] = aws.ToString(res.SecretString)
		}
	} else {
		fmt.Printf("DEBUG: %+v", v.ConfigFileUsed())
		return nil, fmt.Errorf("ERROR: unknown secret type \"%v\"", t)
	}

	return creds, nil
}
