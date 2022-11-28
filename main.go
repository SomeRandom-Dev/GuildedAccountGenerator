package main

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"fmt"
	uuid "github.com/satori/go.uuid"
	"io"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var totalCreated, totalSent, createdThisSecond, sentThisSecond, fails = 0, 0, 0, 0, 0
var proxyList = make([]string, 0)

var proxyType = "http" // Type of proxies (using proxies with the proxy type included e.g http://1.1.1.1:80 will override this setting)
var serverInvite = "" // Guilded server invite
var serverID = "" // Guilded server ID

var messageSpam = true
var channelID = "" // Guilded channel ID
var messageContent = "" // Content of message

func messageSpammer(client *http.Client, UserAgent string) {
	body := bytes.Buffer{}
	uuid_ := uuid.NewV4()
	uuidStr := fmt.Sprintf("%s", uuid_)
	body.WriteString(`{"messageId": "` + uuidStr + `", "content": {"object": "value", "document": {"object": "document", "data": {}, "nodes": [{"object": "block", "type": "paragraph", "data": {}, "nodes": [{"object": "text", "leaves": [{"object": "leaf", "text": "` + messageContent + `", "marks": []}]}]}]}}, "repliesToIds": [], "confirmed": false, "isSilent": false, "isPrivate": false}`)
	req, _ := http.NewRequest("POST", "https://www.guilded.gg/api/channels/"+channelID+"/messages", &body)
	req.Header.Add("User-Agent", UserAgent)
	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fails += 1
		return
	}
	//goland:noinspection GoDeferInLoop
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	totalSent += 1
	sentThisSecond += 1
}

func registerAccounts() {
	for true {
		jar, _ := cookiejar.New(nil)
		client := &http.Client{Jar: jar}

		vn := strconv.Itoa(rand.Intn(105-80) + 80)
		UserAgent := "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:" + vn + ".0) Gecko/" + strconv.Itoa(rand.Intn(999999-100000)+100000) + " Firefox/" + vn + ".0"

		proxy := proxyList[rand.Intn(len(proxyList))]
		if !strings.Contains(proxy, "://") {
			proxy = proxyType + "://" + proxy
		}
		proxy_, err := url.Parse(proxy)
		if err != nil {
			// fmt.Println(err)
			fails += 1
			continue
		}

		tr := &http.Transport{
			Proxy:           http.ProxyURL(proxy_),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}

		client.Transport = tr
		client.Timeout = 5 * time.Second

		email := strconv.Itoa(rand.Intn(9999999-1000000)+1000000) + "@" + strconv.Itoa(rand.Intn(9999999-1000000)+1000000) + "h.us"
		password := "myGoodPW!$" + strconv.Itoa(rand.Intn(9999999-100000)+100000)
		body := bytes.Buffer{}
		body.WriteString(`{"email":"` + email + `","fullName":"` + strconv.Itoa(rand.Int()) + `","name":"` + strconv.Itoa(rand.Intn(9999999-1000000)+1000000) + `cats","password":"` + password + `"}`)

		req, _ := http.NewRequest("POST", "https://www.guilded.gg/api/users?type=email", &body)

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("User-Agent", UserAgent)

		resp, err := client.Do(req)
		if err != nil {
			fails += 1
			continue
		}

		//goland:noinspection GoDeferInLoop
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		time.Sleep(100 * time.Millisecond)

		body = bytes.Buffer{}
		body.WriteString("")
		req, _ = http.NewRequest("GET", "https://www.guilded.gg/api/me?isLogin=false&v2=true", &body)
		req.Header.Add("User-Agent", UserAgent)
		resp, err = client.Do(req)
		if err != nil {
			fails += 1
			// fmt.Println(err)
			continue
		}

		//goland:noinspection GoDeferInLoop
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)

		time.Sleep(150 * time.Millisecond)

		body = bytes.Buffer{}
		body.WriteString(`{"type":"consume"}`)
		req, _ = http.NewRequest("PUT", "https://www.guilded.gg/api/invites/"+serverInvite+"?teamId="+serverID+"&includeLandingChannel=true", &body)
		req.Header.Add("User-Agent", UserAgent)
		req.Header.Add("Content-Type", "application/json")
		resp, err = client.Do(req)
		if err != nil {
			fails += 1
			// fmt.Println(err)
			continue
		}
		// body__, _ := io.ReadAll(resp.Body)
		// fmt.Println(string(body__))
		body_, _ := io.ReadAll(resp.Body)
		bodyStr := string(body_)
		//goland:noinspection GoDeferInLoop
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		if strings.Contains(bodyStr, "have been banned") || strings.Contains(bodyStr, "be signed in") || strings.Contains(bodyStr, "too often") || strings.Contains(bodyStr, "ManyRequests") || !strings.Contains(bodyStr, "landingChannel") {
			fails += 1
			// fmt.Println(bodyStr)
			continue
		} else {
			// fmt.Println(bodyStr)
			fmt.Println("Created! Email: " + email + " | Password: " + password)
			totalCreated += 1
			createdThisSecond += 1

			//goland:noinspection GoBoolExpressions
			if messageSpam {
				go messageSpammer(client, UserAgent)
			}
		}
	}
}

func main() {
	proxyListPath := "./proxies.txt" // Path to your proxies!

	readFile, err := os.Open(proxyListPath)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fileScanner := bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		proxyList = append(proxyList, fileScanner.Text())
	}

	err = readFile.Close()

	if err != nil {
		fmt.Println(err)
	}

	threadCount := len(proxyList) - int(0.1*float64(len(proxyList)))
	// threadCount = 1474 // Uncomment for a custom amount of threads, recommended to use default unless you use a rotating proxy.

	for i := 0; i < threadCount; i++ {
		go registerAccounts()
	}

	titleContent := ""
	for true {
		titleContent = "Accounts: " + strconv.Itoa(totalCreated) + " | J/S: " + strconv.Itoa(createdThisSecond) + " | Messages: " + strconv.Itoa(totalSent) + " | M/S: " + strconv.Itoa(sentThisSecond) + " | Fails: " + strconv.Itoa(fails)
		//goland:noinspection GoBoolExpressions
		if runtime.GOOS == "windows" {
			_ = exec.Command("cmd", "/C", "title", titleContent).Start()
		} else {
			fmt.Println("\\033]0;" + titleContent + "\a")
		}
		createdThisSecond, sentThisSecond = 0, 0
		time.Sleep(1 * time.Second)
	}
}
