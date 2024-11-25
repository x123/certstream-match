/*
certstream-match matches certstream output against regex.
*/
package main

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

var stringarray []string
var addr = flag.String("addr", "127.0.0.1:8181", "http service address")

// {"data":["ph-faq.ru","www.ph-faq.ru"],"message_type":"dns_entries"}
type response1 struct {
	Data        []string `json:"data"`
	MessageType string   `json:"message_type"`
}

func main() {
	regexString := loadRegex()
	fmt.Printf("regex_string:%s\n", regexString)
	var re = regexp.MustCompile(regexString)
	// echoMatchedStdin(re)

	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: *addr, Path: "domains-only"}
	log.Printf("connecting to %s", u.String())

	dialer := *websocket.DefaultDialer
	dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			res := response1{}
			json.Unmarshal(message, &res)
			//fmt.Printf("recv: %s\n", message)
			// fmt.Printf("json recv: %s\n", res.Data)
			for _, value := range res.Data {
				match := re.MatchString(value)
				if match {
					fmt.Println(value)
				}
			}
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func loadRegex() string {
	f, err := os.Open("./regexs.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	var array []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		match, _ := regexp.MatchString("^#", scanner.Text())
		if !match {
			array = append(array, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	var sb strings.Builder
	sb.WriteString("(")
	sb.WriteString(strings.Join(array, "|"))
	sb.WriteString(")")
	return sb.String()
}

func echoMatchedStdin(re *regexp.Regexp) error {
	// This function just echos any stdin that matches regexp re
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		match := re.MatchString(scanner.Text())
		if match {
			fmt.Println(scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
