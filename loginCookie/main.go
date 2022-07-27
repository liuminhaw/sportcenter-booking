package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/chromedp/chromedp"
)

// create chrome instance in windowed mode, run, and close after 5 seconds
func main() {
	headless := flag.Bool("headless", true, "Set chrome browser headless mode")
	flag.Parse()

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", *headless),
	)

	ctx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel = chromedp.NewContext(ctx)
	defer cancel()

	var title string
	if err := chromedp.Run(ctx,
		chromedp.Navigate("https://golang.org/pkg/time/"),
		chromedp.Title(&title),
	); err != nil {
		log.Fatal(err)
	}
	time.Sleep(time.Second * 5)

	fmt.Printf("Title: %s\n", title)
}
