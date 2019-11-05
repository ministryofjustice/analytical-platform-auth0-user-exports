# Auth0 Bulk User Exports

##### Query Auth0's API and export data in csv format to either local file system or an S3 bucket

### Prerequisites

- OIDC Client ID
- OIDC Client Secret 

You'll need to have an auth0 [application](https://auth0.com/docs/applications) in order to retrieve these values

**Note**

This app requests temporary credentials every time it executes.  You'll need to ensure the resulting bearer tokens have the following scopes

- `create:client_grants`
- `read:client_grants`
- `read:connections`
- `read:users`

### Usage

```bash
go run main.go
```
##### Or

```bash
GOOS=linux go build -o auth0-bulk-user-export
```

```bash
./auth0-bulk-user-export
```

##### Or

```bash
docker image build -t analytical-platform-auth0-user-exports .
```

```bash
docker container run -it --rm -v /tmp:/tmp --env CLIENT_ID=41Tmoz00wBN1... --env CLIENT_SECRET=StdNYnUaPuv9iMEfKsiLEZ0GUTe... --name analytical-platform-auth0-user-exports analytical-platform-auth0-user-exports
```

#### Write data to local file: `~/Dowloads/userdata.csv`

```bash
export CLIENT_ID="41Tmoz00wBN1..."
export CLIENT_SECRET="StdNYnUaPuv9iMEfKsiLEZ0GUTe...."
export FILE_PATH="~/Dowloads/userdata.csv"
```

#### Write to S3

```bash
export CLIENT_ID="41Tmoz00wBN1..."
export CLIENT_SECRET="StdNYnUaPuv9iMEfKsiLEZ0GUTe...."
export ENV=aws
export BUCKET=my-auth0-bucket
```

__Using [SSM](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-parameter-store.html)__

```bash
export SSM_PATH=/airflow/my_airflow_job_secrets/
export ENV=aws
export BUCKET=my-auth0-bucket

```

### Configuration

| Env Variable  | Default  | Description                                |
|---------------|----------|--------------------------------------------|
| `CLIENT_ID` | (**Required**) | The Client ID of the auth0 application used for this app |
| `CLIENT_SECRET` | (**Required**) | The Client Secret of the auth0 application used for this app |
| `SSM_PATH`    | (**Required**) | **Required If** `CLIENT_ID` and `CLIENT_SECRET` are unset | 
| `API_URL`     | https://alpha-analytics-moj.eu.auth0.com | Auth0 management API endpoint |
| `CONNECTION_NAME` | `github` | Config param for auth0.  Which connection to target when querying the API (https://auth0.com/docs/identityproviders) |
| `ENV` | (**Write Data Locally**) | **Do not set** to write locally or set to `aws` to write data to `S3` |
| `ROLE_ARN` | `arn:aws:iam::593291632749:role/airflow_auth0-user-exports` | The role this app uses to write to `S3` |
| `FILE_PATH` | `/tmp/userdata.csv` | File path when writing locally. Only works when `ENV` is **not** set |
| `BUCKET` | `auth0-userdata` | The `S3` bucket to write to when `ENV=aws` is set.  The resulting key will be suffixed with the date i.e `userdata-22-09-2019` |

##### Test

To run tests, `cd` to the root of this project

```bash
go test ./...
```
