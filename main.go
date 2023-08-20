package main

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"sync"
	"syscall"
	"time"

	"io/ioutil"
	"strings"

	"github.com/go-ping/ping"
	"github.com/nsf/termbox-go"
)

type Result struct {
	IP           string
	Name         string // New field for name
	ResponseTime time.Duration
	Error        error
}

var results []Result // Use an array to store results
var mu sync.Mutex

func drawUI() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	_, termHeight := termbox.Size()

	y := 1
	for _, result := range results {
		if y > termHeight-2 {
			break
		}
		x := 1
		fgColor := termbox.ColorWhite
		if result.ResponseTime == 0 {
			fgColor = termbox.ColorRed // Set text color to red for "0s"
		}
		ipAndName := result.IP
		if result.Name != "" {
			ipAndName += " (" + result.Name + ")"
		}
		drawString(x, y, fgColor, termbox.ColorDefault, ipAndName) // Display IP and name
		x += 50
		drawString(x, y, fgColor, termbox.ColorDefault, result.ResponseTime.String())
		y++
	}

	termbox.Flush()
}

func drawString(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x++
	}
}

func readIPsFromFile(filename string) ([]Result, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(string(content), "\n")
	var results []Result
	for _, line := range lines {
		fields := strings.Split(line, ",")
		if len(fields) >= 1 {
			ip := strings.TrimSpace(fields[0])
			var name string
			if len(fields) >= 2 {
				name = strings.TrimSpace(fields[1])
			}
			results = append(results, Result{IP: ip, Name: name})
		}
	}
	return results, nil
}

func pingIP(ip string, wg *sync.WaitGroup, interval time.Duration, done <-chan struct{}, idx int) {
	defer wg.Done()

	for {
		select {
		case <-done:
			return
		default:
			pinger, err := ping.NewPinger(ip)
			if err != nil {
				fmt.Printf("Error creating pinger for %s: %v\n", ip, err)
				return
			}

			pinger.Count = 1
			pinger.Timeout = interval

			err = pinger.Run()
			responseTime := time.Duration(0)
			if err != nil {
				//fmt.Printf("Error pinging %s: %v\n", ip, err)
				responseTime = 0
			} else {
				responseTime = pinger.Statistics().AvgRtt
			}

			mu.Lock()
			results[idx].ResponseTime = responseTime // Update only the response time
			mu.Unlock()

			time.Sleep(interval)
		}
	}
}

func main() {
	if err := termbox.Init(); err != nil {
		fmt.Println("Error initializing termbox:", err)
		return
	}
	defer termbox.Close()

	if len(os.Args) < 3 {
		fmt.Println("Usage: app <interval> <ips_file>")
		return
	}

	interval, _ := time.ParseDuration(os.Args[1])
	ipsFile := os.Args[2]

	ipResults, err := readIPsFromFile(ipsFile)
	if err != nil {
		fmt.Println("Error reading IPs from file:", err)
		return
	}

	// Sort the list of IP results by IP address
	sort.Slice(ipResults, func(i, j int) bool {
		return ipResults[i].IP < ipResults[j].IP
	})

	done := make(chan struct{})
	results = ipResults // Use the read IP results
	var wg sync.WaitGroup
	for idx, ipResult := range ipResults {
		wg.Add(1)
		go pingIP(ipResult.IP, &wg, interval, done, idx)
	}

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigchan
		fmt.Printf("Received signal %s. Exiting...\n", sig)
		close(done)
		wg.Wait()
		os.Exit(0)
	}()

	updateTicker := time.NewTicker(interval)
	defer updateTicker.Stop()

	eventQueue := make(chan termbox.Event)
	go func() {
		for {
			eventQueue <- termbox.PollEvent()
		}
	}()

	for {
		select {
		case <-updateTicker.C:
			drawUI()
		case ev := <-eventQueue:
			switch ev.Type {
			case termbox.EventKey:
				if ev.Ch == 'q' || ev.Key == termbox.KeyCtrlC {
					close(done)
					wg.Wait()
					return
				}
			}
		}
	}
}
