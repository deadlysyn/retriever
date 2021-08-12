# retriever

![Golden Retriever Puppies](https://raw.githubusercontent.com/deadlysyn/retriever/main/assets/retriever.png "retrievers")

Easily fetch secrets from AWS [Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html) or [Secrets Manager](https://aws.amazon.com/secrets-manager).

# Why

After a week where I wrote two CLIs and a small web service that required copying
and pasting the same secret-fetching boilerplate, I decided to abstract the
boring details.

# How

You can configure using a YAML file or environment thanks to viper.
Let's see some examples...

Parameter Store with many secrets under one prefix:

```yaml
type: parameter
prefix: /foo
credentials:
  - BAR
  - BAZ_QUX
```

Parameter Store with different secret prefixes (note: no leading `/`):

```yaml
type: parameter
credentials:
  - foo/BAR
  - baz/QUX
```

Same rules for Secrets Manager, just change `type`:

```yaml
type: secret
prefix: /foo
credentials:
  - BAR
  - BAZ_QUX
```

Environment variable names are prefixed with `RTVR_` and have the same keys as the YAML config.

```console
❯ RTVR_TYPE="secret" RTVR_PREFIX="/foo" RTVR_CREDENTIALS="BAR"  aws-vault exec dev --region us-east-2 -- go run main.go
no retriever config found; using environment
RESULT: map[BAR:{"key1":"value1","key2":"value2"}]%

❯ RTVR_TYPE="parameter" RTVR_CREDENTIALS="foo/BAR QUX" aws-vault exec dev --region us-east-2 -- go run main.go
no retriever config found; using environment
RESULT: map[QUX:top secret foo/BAR:baz]%
```

To dispell any magic, that test code is just:

```go
package main

import (
        "fmt"
        "log"

        "github.com/deadlysyn/retriever"
)

func main() {
        c, err := retriever.Fetch()
        if err != nil {
                log.Fatalf("CALLER: %v", err)
        }
        fmt.Printf("RESULT: %+v", c)
}
```

Of course the idea is not to print things out, but to use the returned values.
`Fetch()` returns a map of credentials with keys equal to the secret names.
Let's see that in action:

```yaml
type: parameter
prefix: /app/dev
credentials:
  - CONSUMER_SECRET
  - CONSUMER_KEY
  - OAUTH_TOKEN
  - OAUTH_TOKEN_SECRET
```

```go
// ...

var (
        creds = make(map[string]string)
)

func jiraGetClient(ctx context.Context) *http.Client {
        jiraURL := viper.GetString("jira.url")

        keyDERBlock, _ := pem.Decode([]byte(creds["CONSUMER_SECRET"]))
        privateKey, _ := x509.ParsePKCS1PrivateKey(keyDERBlock.Bytes)

        config := oauth1.Config{
                ConsumerKey: creds["CONSUMER_KEY"],
                // etc...
        }

        tok := &oauth1.Token{
                Token:       creds["OAUTH_TOKEN"],
                TokenSecret: creds["OAUTH_TOKEN_SECRET"],
        }

        return config.Client(ctx, tok)
}

// ...
```

# TODO

- Secrets Manager versioning
- Customer managed KMS keys
- Values other than strings
- Ideas? Open an issue or PR.

# Thanks

On the shoulders of giants.

- [AWS SDK for Go V2](https://aws.github.io/aws-sdk-go-v2/docs/getting-started) :rocket:
- [spf13/viper](https://github.com/spf13/viper) :mindblown:
