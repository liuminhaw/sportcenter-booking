package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
)

func getTime() (time.Time, error) {
	var err error

	var yearToTick int
	if val, ok := os.LookupEnv("YEAR_TO_TICK"); ok {
		yearToTick, err = strconv.Atoi(val)
		if err != nil {
			return time.Time{}, errors.New("failed to convert string to integer")
		}
	} else {
		return time.Time{}, errors.New("missing environment variable: YEAR_TO_TICK")
	}

	var monthToTick int
	if val, ok := os.LookupEnv("MONTH_TO_TICK"); ok {
		monthToTick, err = strconv.Atoi(val)
		if err != nil {
			return time.Time{}, errors.New("failed to convert string to integer")
		}
	} else {
		return time.Time{}, errors.New("missing environment variable: MONTH_TO_TICK")
	}

	var dateToTick int
	if val, ok := os.LookupEnv("DATE_TO_TICK"); ok {
		dateToTick, err = strconv.Atoi(val)
		if err != nil {
			return time.Time{}, errors.New("failed to convert string to integer")
		}
	} else {
		return time.Time{}, errors.New("missing environment variable: DATE_TO_TICK")
	}

	var hourToTick int
	if val, ok := os.LookupEnv("HOUR_TO_TICK"); ok {
		hourToTick, err = strconv.Atoi(val)
		if err != nil {
			return time.Time{}, errors.New("failed to convert string to integer")
		}
	} else {
		return time.Time{}, errors.New("missing environment variable: HOUR_TO_TICK")
	}

	var minuteToTick int
	if val, ok := os.LookupEnv("MINUTE_TO_TICK"); ok {
		minuteToTick, err = strconv.Atoi(val)
		if err != nil {
			return time.Time{}, errors.New("failed to convert string to integer")
		}
	} else {
		return time.Time{}, errors.New("missing environment variable: MINUTE_TO_TICK")
	}

	var secondToTick int
	if val, ok := os.LookupEnv("SECOND_TO_TICK"); ok {
		secondToTick, err = strconv.Atoi(val)
		if err != nil {
			return time.Time{}, errors.New("failed to convert string to integer")
		}
	} else {
		return time.Time{}, errors.New("missing environment variable: MINUTE_TO_TICK")
	}

	loc, err := time.LoadLocation("Asia/Taipei")
	if err != nil {
		return time.Time{}, errors.New("timezone Asia/Taipei not found")
	}

	execution := time.Date(yearToTick, time.Month(monthToTick), dateToTick, hourToTick, minuteToTick, secondToTick, 0, loc)

	return execution, nil
}

func getURL() (string, error) {
	var qCourt string
	if val, ok := os.LookupEnv("Q_COURT"); ok {
		qCourt = val
	} else {
		return "", errors.New("missing environment variable: Q_COURT")
	}

	var qTime string
	if val, ok := os.LookupEnv("Q_TIME"); ok {
		qTime = val
	} else {
		return "", errors.New("missing environment variable: Q_TIME")
	}

	var qDate string
	if val, ok := os.LookupEnv("Q_DATE"); ok {
		qDate = val
	} else {
		return "", errors.New("missing environment variable: Q_DATE")
	}

	return fmt.Sprintf("https://scr.cyc.org.tw/tp03.aspx?module=net_booking&files=booking_place&StepFlag=25&QPid=%s&QTime=%s&PT=1&D=%s", qCourt, qTime, qDate), nil
}

func getCookie() (string, error) {
	var cookie string
	if val, ok := os.LookupEnv("COOKIE"); ok {
		cookie = val
	} else {
		return "", errors.New("missing environment variable: COOKIE")
	}

	return fmt.Sprintf("ASP.NET_SessionId=%s", cookie), nil
}

func submit(t time.Time, url string, cookie string) {
	client := &http.Client{}

	fmt.Printf("time until execution: %v\n", time.Until(t))
	time.Sleep(time.Until(t))

	fmt.Println(time.Now(), "- just ticked")
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Cookie", cookie)
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer resp.Body.Close()

	fmt.Printf("Response status: %v\n", resp.Status)
	fmt.Printf("Response header: %v\n", resp.Header)
}

func Tick() error {

	execution, err := getTime()
	if err != nil {
		return err
	}

	url, err := getURL()
	if err != nil {
		return err
	}

	cookie, err := getCookie()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func(t time.Time, url string, cookie string) {
		defer wg.Done()
		submit(t, url, cookie)
	}(execution, url, cookie)

	go func(t time.Time, url string, cookie string) {
		defer wg.Done()
		submit(t, url, cookie)
	}(execution.Add(500*time.Millisecond), url, cookie)

	go func(t time.Time, url string, cookie string) {
		defer wg.Done()
		submit(t, url, cookie)
	}(execution.Add(1*time.Second), url, cookie)

	go func(t time.Time, url string, cookie string) {
		defer wg.Done()
		submit(t, url, cookie)
	}(execution.Add(1*time.Second+500*time.Millisecond), url, cookie)

	wg.Wait()

	return nil
}

func main() {
	lambda.Start(Tick)
}
