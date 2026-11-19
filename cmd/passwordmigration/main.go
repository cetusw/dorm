package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"dorm/pkg/infrastructure/config"
	"dorm/pkg/infrastructure/mysql"
	"dorm/pkg/infrastructure/mysql/repository"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return errors.New("usage: dorm-password-migration dry-run|apply|check (login password lines on stdin)")
	}
	action := os.Args[1]
	if action != "dry-run" && action != "apply" && action != "check" {
		return errors.New("unknown migration action")
	}
	entries := map[string]string{}
	if action != "check" {
		reader := io.LimitReader(os.Stdin, 1048577)
		data, err := io.ReadAll(reader)
		if err != nil || len(data) > 1048576 {
			return errors.New("cannot read credential input or input too large")
		}
		scanner := bufio.NewScanner(strings.NewReader(string(data)))
		for scanner.Scan() {
			fields := strings.Fields(scanner.Text())
			if len(fields) == 0 {
				continue
			}
			if len(fields) != 2 || len(fields[0]) > 255 || len(fields[1]) > 4096 {
				return errors.New("invalid credential line; expected login and password without spaces")
			}
			login := strings.ToLower(strings.TrimSpace(fields[0]))
			if _, ok := entries[login]; ok {
				return errors.New("duplicate credential login")
			}
			entries[login] = fields[1]
		}
		if scanner.Err() != nil || len(entries) == 0 || len(entries) > 1000 {
			return errors.New("invalid or empty credential input")
		}
	}
	cfg, err := config.LoadDatabaseConfig()
	if err != nil {
		return errors.New("invalid database configuration")
	}
	db, err := mysql.NewConnection(cfg)
	if err != nil {
		return errors.New("cannot connect to credential database")
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if action == "check" {
		if err := repository.CheckPasswordReadiness(ctx, db); err != nil {
			return err
		}
		fmt.Println("Active credentials are Argon2id-ready")
		return nil
	}
	count, err := repository.MigratePasswords(ctx, db, entries, action == "dry-run")
	if err != nil {
		return err
	}
	fmt.Printf("Credential migration %s: %d updates\n", action, count)
	return nil
}
