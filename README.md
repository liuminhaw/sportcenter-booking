# sportcenter-booking
Online reservation for sport center's facilities

## AWS sam
Building application
```
sam build
``` 
Local testing API
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