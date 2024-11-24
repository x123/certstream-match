/*
certstream-match matches certstream output against regex.
*/
package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	echo()
}

func echo() error {
	// This function just echos any stdin.
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		fmt.Println(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}
