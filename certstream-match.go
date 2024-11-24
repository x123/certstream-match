/*
certstream-match matches certstream output against regex.
*/
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var stringarray []string

func main() {
	regexString := loadRegex()
	fmt.Printf("regex_string:%s\n", regexString)
	var re = regexp.MustCompile(regexString)
	echoMatchedStdin(re)
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
