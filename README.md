# sportcenter-reserve
Online reservation for sport center's facilities

## Lambda
### Setup
#### Runtime settings 
- Handler: `sportcenter-booking`

#### Configuration
Timeout: Set `timeout` according to EventBridge trigger time, `timeout` should be longer than the duration of function triggered and function submit.

#### Environment variables
| Key | Value |
| --- | --- |
| COOKIE | Login session cookie |
| YEAR_TO_TICK | Submit year |
| MONTH_TO_TICK | Submit month (1 ~ 12) |
| DATE_TO_TICK | Submit date of month |
| HOUR_TO_TICK | Submit time hour |
| MINUTE_TO_TICK | Submit time minute |
| SECOND_TO_TICK | Submit time second |
| Q_DATE | Reservation date (YYYY/mm/dd) |
| Q_TIME | Reservation time session (06 ~ 21) |
| Q_COURT | Reserve target court (Checkout court mapping table) |

## EventBridge
Setup schedule rules to trigger `sportcenter-booking` lambda session

## Deployment
1. Compile executable
	```
	GOOS=linux go build
	```
1. Create deployment package
	```
	zip function.zip sportcenter-booking
	```
1. Upload zip file (`function.zip`) to Lambda function

## Court mapping
### Taipei Da-an sport center
| Court | ID |
| --- | --- |
| 10 | 1094 |
| 9 | 1093 |
| 8 | 1092 |
| 7 | 1091 |
| 6 | 1090 |
| 5 | 1089 |
| 4 | 1088 |
| 3 | 1087 |
| 2 | 1086 |
| 1 | 1085 |