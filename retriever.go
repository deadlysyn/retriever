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

const DEFAULT_CONFIG = "retriever.yml"

type store map[string]string

var (
	Creds store = make(map[string]string)
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("RTVR")
	viper.AutomaticEnv()

	cfg := viper.GetString("config")
	if cfg != "" {
		viper.SetConfigFile(cfg)
	} else {
		viper.SetConfigFile(DEFAULT_CONFIG)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

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

func Fetch() (store, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("unable to load AWS configuration: %v", err)
		return nil, err
	}

	p := viper.GetString("prefix")
	t := strings.ToLower(viper.GetString("type"))
	if t == "parameter" {
		client := ssm.NewFromConfig(cfg)
		for _, v := range viper.GetStringSlice("credentials") {
			res, err := getParam(ctx, client, fmt.Sprintf("%s/%s", p, v))
			if err != nil {
				log.Fatalf("unable to retrieve %v/%v (%v)", p, v, err)
			}
			Creds[v] = aws.ToString(res.Parameter.Value)
		}
	} else if t == "secret" {
		client := secretsmanager.NewFromConfig(cfg)
		for _, v := range viper.GetStringSlice("credentials") {
			res, err := getSecret(ctx, client, fmt.Sprintf("%s/%s", p, v))
			if err != nil {
				log.Fatalf("unable to retrieve %v/%v (%v)", p, v, err)
			}
			Creds[v] = aws.ToString(res.SecretString)
		}
	} else {
		log.Fatalf("ERROR: unknown type \"%v\"", t)
	}

	return Creds, nil
}
