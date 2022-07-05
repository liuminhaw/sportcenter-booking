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

# curl and sent data
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
