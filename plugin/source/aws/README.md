# AWS Secrets Manager Secret Source (`aws://`)

The **AWS Secrets Manager** secret source retrieves secrets directly from [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/).

## Status

**Plugin**: This source is **opt-in**. It is not enabled by default and requires explicit configuration using `WithAWS()`.

## Dependencies

This plugin requires the official AWS SDK for Go v2:
- `github.com/aws/aws-sdk-go-v2/service/secretsmanager`

## Usage

To use the AWS Secrets Manager source, use the `aws://` scheme followed by either the secret name or the full secret ARN. 

### Syntax

```text
aws://<SECRET_NAME>
aws:///<SECRET_ARN>
```

**⚠️ Important NOTE regarding ARNs:**
Because an AWS ARN contains colons (`arn:aws:secretsmanager:...`), a standard URI parser will attempt to interpret the text after the first colon as a port number, resulting in an error. 
To bypass this, you **must use three slashes** (`aws:///arn:...`) when addressing a secret by its ARN. This tells the parser that the URI has an empty host and the ARN is simply the path.

### Examples

Retrieve a secret using its short name:

```text
aws://my-database-credentials
```

Retrieve a secret using its full ARN:

```text
aws:///arn:aws:secretsmanager:us-east-1:123456789012:secret:my-database-credentials-1234
```

Using modifiers (e.g., extracting JSON path) safely ignores trailing slashes in the path:

```text
aws://my-json-secret/?jp=$.password
```

## Configuration

To use this source, you must initialize `spelunk` with an AWS Secrets Manager client:

```go
import (
    "context"
    "github.com/aws/aws-sdk-go-v2/config"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
    "github.com/detro/spelunk"
    spelunkaws "github.com/detro/spelunk/plugin/source/aws"
)

func main() {
    ctx := context.Background()

    // 1. Load AWS configuration (picks up from env, ~/.aws/credentials, or IAM roles)
    cfg, _ := config.LoadDefaultConfig(ctx)

    // 2. Create the Secrets Manager client
    awsClient := secretsmanager.NewFromConfig(cfg)

    // 3. Initialize Spelunker with the AWS plugin
    s := spelunk.NewSpelunker(
        spelunkaws.WithAWS(awsClient),
    )

    // 4. Dig up secrets
    secret, _ := s.DigUp(ctx, coord)
}
```

## Behavior

1. **Parsing & Cleaning**: 
   - Uses the location (hostname + path) as the Secret ID. 
   - If a leading slash is present (common when using ARNs with `aws:///`), it is trimmed.
   - Any trailing slash (e.g., when the URI contains query parameters like `/?jp=$.password`) is stripped automatically.
2. **Validation**: The cleaned Secret ID is strictly validated against official AWS rules before any network call is made:
   - **Names**: Must be 1-512 characters containing only alphanumeric characters and `/_+=.@-`.
   - **ARNs**: Must match the standard Secrets Manager ARN format. The validation logic naturally supports alternative AWS partitions (e.g., `aws-cn`, `aws-us-gov`, `aws-iso`).
   - **Restriction**: As per [AWS documentation](https://docs.aws.amazon.com/secretsmanager/latest/apireference/API_CreateSecret.html), a secret name **must not** end with a hyphen followed by six alphanumeric characters (to avoid confusion with ARNs).
3. **Retrieval**: Uses `awsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{SecretId: ...})` to fetch the secret.
4. **Extraction**: Returns either the `SecretString` or `SecretBinary` depending on how the secret is stored in AWS.
5. **Errors**:
    - Returns `ErrSecretSourceAWSInvalidLocation` if the location does not match either the valid Name or ARN format.
    - Returns `ErrSecretSourceAWSInvalidNameSuffix` if a secret name violates the "no hyphen + 6 characters suffix" rule.
    - Returns `ErrCouldNotFetchSecret` if the API call fails due to permissions or network issues.
    - Returns `ErrSecretNotFound` if the secret does not exist or has no payload.

## Testing

Integration tests for this plugin are powered by [Testcontainers](https://golang.testcontainers.org/) using the [localstack/localstack](https://hub.docker.com/r/localstack/localstack) image. They are automatically skipped in short test mode (`go test -short` or `task test.short`).

## Use Cases

- Dynamically fetching database credentials, API keys, or certificates managed by AWS Secrets Manager across AWS environments (EKS, ECS, EC2, Lambda).
