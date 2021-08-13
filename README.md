# retriever

![Golden Retriever Puppies](https://raw.githubusercontent.com/deadlysyn/retriever/main/assets/retriever.png "retrievers")

Easily fetch secrets from AWS
[Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)
or [Secrets Manager](https://aws.amazon.com/secrets-manager).

After a week where I wrote two CLIs and a small web service that required copying
and pasting the same secret-fetching boilerplate, I decided to abstract the
boring details.

# Usage

Configure using a YAML or environment thanks to viper. Some examples...

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

Environment variables are prefixed with `RTVR_` and have the same keys as YAML.

```console
❯ RTVR_TYPE="secret" RTVR_PREFIX="/foo" RTVR_CREDENTIALS="BAR" aws-vault exec dev -- go run main.go
INFO: retriever not config found; using environment
RESULT: map[BAR:{"key1":"value1","key2":"value2"}]%

❯ RTVR_TYPE="parameter" RTVR_CREDENTIALS="foo/BAR QUX" aws-vault exec dev -- go run main.go
INFO: retriever not config found; using environment
RESULT: map[QUX:top secret foo/BAR:baz]%
```

To dispel any magic, test code is just:

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

Of course the idea is not to print things, but to use returned values.
`Fetch()` returns a map of credentials with keys equal to secret names.
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
package main

import (
    // ...

    "github.com/deadlysyn/retriever"
)

var (
    creds = make(map[string]string)
)

func init() {
    creds, err := retriever.Fetch()
    if err != nil {
        log.Fatalf("deal with it: %v", err)
    }
}

func getClient(ctx context.Context) *http.Client {
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
- Values other than strings (?)
- Ideas? Open an issue or PR.

# Thanks

On the shoulders of giants.

- [AWS SDK for Go V2](https://aws.github.io/aws-sdk-go-v2/docs/getting-started) :rocket: :heart_eyes:
- [spf13/viper](https://github.com/spf13/viper) :sunglasses: :sparkles:
- [flashback coding music](https://open.spotify.com/playlist/3Y8Dpo4TuNX0QHDDum45Gg?si=3f616312eabd4024) :notes: :metal:
