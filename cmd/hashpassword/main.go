package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	useruc "dorm/pkg/core/usecase/user"
)

func main() {
	value, err := io.ReadAll(io.LimitReader(os.Stdin, 4097))
	if err != nil || len(value) > 4096 || strings.TrimSpace(string(value)) == "" {
		fmt.Fprintln(os.Stderr, "cannot read password")
		os.Exit(1)
	}
	hash, err := useruc.HashPassword(strings.TrimSpace(string(value)))
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot hash password")
		os.Exit(1)
	}
	fmt.Println(hash)
}
