package retriever

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/spf13/viper"
)

const DEFAULT_CONFIG = "config.yml"

type store map[string]string

var (
	Creds store = make(map[string]string)
)

func init() {
	viper.SetConfigType("yaml")
	viper.SetEnvPrefix("RTVR")
	viper.AutomaticEnv()

	cfg := viper.GetString("config")

	fmt.Printf("DEBUG: %+v\n", cfg)

	if cfg != "" {
		viper.SetConfigFile(cfg)
	} else {
		viper.SetConfigFile(DEFAULT_CONFIG)
	}

	// c := viper.Get("config")
	// if c != nil {
	// 	viper.SetConfigFile(c.(string))
	// }

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("ERROR: %v", err)
	}
}

func getParam(c *ssm.Client, p string) (*ssm.GetParameterOutput, error) {
	i := ssm.GetParameterInput{
		Name:           aws.String(p),
		WithDecryption: true,
	}

	out, err := c.GetParameter(context.TODO(), &i)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func getSecret() {
	fmt.Println("not implemented")
}

func Load() (store, error) {
	ctx := context.TODO()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		fmt.Printf("unable to load AWS configuration: %v", err)
		return nil, err
	}

	t := strings.ToLower(viper.GetString("type"))
	if t == "parameter" {
		ssmClient := ssm.NewFromConfig(cfg)

		for _, v := range viper.GetStringSlice("credentials") {
			result, err := getParam(ssmClient, fmt.Sprintf("%s/%s", viper.GetString("ssm.prefix"), v))
			if err != nil {
				fmt.Printf("unable to retrieve %v (%v)", p, err)
			}
			Creds[v] = aws.ToString(result.Parameter.Value)
		}
	} else if t == "secret" {
		getSecret()
	} else {
		log.Fatalf("ERROR: unknown type \"%v\"", t)
	}

	return Creds, nil
}
