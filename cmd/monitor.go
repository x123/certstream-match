/*
Copyright © 2024 x123 <x123@users.noreply.github.com>
*/
package cmd

import (
	"bufio"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

var (
	stringarray []string
	// addr        = flag.String("addr", "127.0.0.1:8181", "http service address")

	regexFile     string
	URL           string
	TLSSkipVerify bool

	// monitorCmd represents the monitor command
	monitorCmd = &cobra.Command{
		Use:   "monitor",
		Args:  cobra.MatchAll(cobra.OnlyValidArgs),
		Short: "monitor a certstream websocket, optionally matching against regex",
		Long:  `monitor a certstream websocket, optionally matching against regex`,
		Run:   monitorGeneral,
	}
)

/*
response from certstream-server /domains-only endpoint looks like
{"data":["example.com","www.example.com"],"message_type":"dns_entries"}
*/
type response1 struct {
	Data        []string `json:"data"`
	MessageType string   `json:"message_type"`
}

func loadRegex(path string) string {
	f, err := os.Open(path)
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

func monitorGeneral(cmd *cobra.Command, args []string) {
	var re *regexp.Regexp
	if regexFile != "" {
		regexString := loadRegex("./regexs.txt")
		fmt.Printf("regex_string:%s\n", regexString)
		re = regexp.MustCompile(regexString)
	}

	// flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "wss", Host: URL, Path: "domains-only"}
	log.Printf("connecting to %s", u.String())

	dialer := *websocket.DefaultDialer
	if TLSSkipVerify {
		dialer.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
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
			for _, value := range res.Data {
				if regexFile != "" {
					match := re.MatchString(value)
					if !strings.HasPrefix(value, "*") && match {
						fmt.Println(value)
					}
				} else {
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

func init() {
	rootCmd.AddCommand(monitorCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	monitorCmd.PersistentFlags().StringVarP(
		&regexFile, "regexFile", "f", "", "File containing regex",
	)
	monitorCmd.PersistentFlags().StringVarP(
		&URL, "URL", "u", "", "URL for certstream server websocket",
	)
	monitorCmd.MarkPersistentFlagRequired("URL")
	monitorCmd.PersistentFlags().BoolVarP(
		&TLSSkipVerify, "TLSSkipVerify", "k", false, "Skip TLS certificate verification",
	)
	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// monitorCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
