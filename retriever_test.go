package retriever

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

var tests = map[string]struct {
	Type       string
	Prefix     string
	Credential string
	Value      string
}{
	"paramWithPrefix": {
		Type:       "parameter",
		Prefix:     "/" + randomString(10),
		Credential: randomString(10),
		Value:      "foo",
	},
	"paramNoPrefix": {
		Type:       "parameter",
		Prefix:     "",
		Credential: randomString(10),
		Value:      "bar",
	},
	"secretWithPrefix": {
		Type:       "secret",
		Prefix:     "/" + randomString(10),
		Credential: randomString(10),
		Value:      "{\"foo\": \"bar\"}",
	},
	"secretNoPrefix": {
		Type:       "secret",
		Prefix:     "",
		Credential: randomString(10),
		Value:      "{\"foo\": \"bar\"}",
	},
}

func TestMain(m *testing.M) {
	ctx := context.TODO()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("ERROR: unable to load AWS configuration: %v", err.Error())
	}
	pc := ssm.NewFromConfig(cfg)
	sc := secretsmanager.NewFromConfig(cfg)

	defer func() {
		teardown(ctx, pc, sc)
	}()

	setup(ctx, pc, sc)
	m.Run()
}

func TestFetch(t *testing.T) {
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Logf("=== %v: Started", name)
			os.Setenv("RTVR_TYPE", test.Type)
			if len(test.Prefix) > 0 {
				os.Setenv("RTVR_PREFIX", test.Prefix)
			} else {
				os.Unsetenv("RTVR_PREFIX")
			}
			os.Setenv("RTVR_CREDENTIALS", test.Credential)
			res, err := Fetch()
			if err != nil {
				t.Errorf("=== %v: %v", name, err.Error())
			} else {
				if res[test.Credential] == test.Value {
					t.Logf("=== %v: %v", name, res)
				} else {
					t.Errorf("=== %v: %v", name, err.Error())
				}
			}
		})
	}
}

func setup(ctx context.Context, pc *ssm.Client, sc *secretsmanager.Client) {
	for name, test := range tests {
		fmt.Printf("setup %v: %v\n", name, test.Prefix+"/"+test.Credential)
		if test.Type == "parameter" {
			i := ssm.PutParameterInput{
				Description: aws.String("Retriever automated test value"),
				Name:        aws.String(test.Prefix + "/" + test.Credential),
				Overwrite:   true,
				Type:        types.ParameterTypeSecureString,
				Value:       aws.String(test.Value),
			}
			_, err := pc.PutParameter(ctx, &i)
			if err != nil {
				log.Fatalf("%v", err.Error())
			}
		} else {
			i := secretsmanager.CreateSecretInput{
				Description:  aws.String("Retriever automated test value"),
				Name:         aws.String(test.Prefix + "/" + test.Credential),
				SecretString: aws.String(test.Value),
			}
			_, err := sc.CreateSecret(ctx, &i)
			if err != nil {
				log.Fatalf("%v", err.Error())
			}

		}
	}
}

func teardown(ctx context.Context, pc *ssm.Client, sc *secretsmanager.Client) {
	for name, test := range tests {
		fmt.Printf("teardown %v: %v\n", name, test.Prefix+"/"+test.Credential)
		if test.Type == "parameter" {
			i := ssm.DeleteParameterInput{
				Name: aws.String(test.Prefix + "/" + test.Credential),
			}
			_, err := pc.DeleteParameter(ctx, &i)
			if err != nil {
				log.Fatalf("%v", err.Error())
			}
		} else {
			i := secretsmanager.DeleteSecretInput{
				SecretId: aws.String(test.Prefix + "/" + test.Credential),
			}
			_, err := sc.DeleteSecret(ctx, &i)
			if err != nil {
				log.Fatalf("%v", err.Error())
			}

		}
	}

}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyz0123456789")

	rand.Seed(time.Now().UnixNano())

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
