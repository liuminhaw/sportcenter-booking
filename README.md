# sportcenter-booking
Online reservation for sport center's facilities

## AWS sam
Building application
```
sam build
``` 
### Local testing API
```bash
# start local api service
sam local start-api

# start local api service with custom environment variables and aws credential profile name
sam local start-api -n environments-dev.json --profile role-profile-name
```

### Reserve Registry Request
```
curl -X POST localhost:3000/api/registry -d '
    {
        "username": "redone", 
        "password": "password", 
        "reserveDate": "2022-07-04T23:23:23Z", 
        "reserveCourt": "1", 
        "reserveTime": "21"
    }
'
```

### Fetch Registry Request
```
curl -X GET localhost:3000/api/registry?id=xxxxxxx
```

### Environment file
Environment json file format for local testing
```json
{
    "ReserveRegistry": {
        "S3Bucket": "s3-bucket-name",
        "secretKey": "s3 byte hex value (32byte hex)"
    },
    "FetchRegistry": {
        "S3Bucket": "s3-bucket-name",
        "secretKey": "s3 byte hex value (32byte hex)"
    }
}
```

## AWS ECS

### Login Cookie Feature
Create `.env` file to store required variables 
- _AWS_REGION
- _ENDPOINT_AWS_PROFILE
- _APP_AWS_ROLE
- _ENC_KEY
- _S3_BUCKET
- _DAAN_LOGIN

#### Testing with container and role
Execute docker compose for local testing
```bash
docker compose up -d
```
View outputs
```bash
docker compose logs [SERVICE]
```

#### Testing program execution
Build
```bash
go build -o loginCookie.out
```
Test
```bash
env $(cat .env | xargs) ./loginCookie.out -profile="aws profile name"
```