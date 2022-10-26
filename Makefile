build-ReserveRegistry:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o reservation-registry reservationRegistry/main.go
	mv reservation-registry $(ARTIFACTS_DIR)

build-FetchRegistry:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o fetch-registry fetchRegistry/main.go
	mv fetch-registry $(ARTIFACTS_DIR)

build-DailyExecCheck:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o daily-check dailyExecCheck/main.go
	mv daily-check $(ARTIFACTS_DIR)