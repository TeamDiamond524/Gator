package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/TeamDiamond524/Gator/internal/config"
	"github.com/TeamDiamond524/Gator/internal/database"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type state struct {
    cfg *config.Config
    db  *database.Queries
}

type command struct {
    name string
    args []string
}

type commands struct {
    CommandList map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
    if s == nil {
        return fmt.Errorf("state not provided or is nil")
    }

    callback, exists := c.CommandList[cmd.name]
    if !exists {
        return fmt.Errorf("that command does not exist")
    }

    return callback(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
    if name == "" { 
        fmt.Printf("command name cannot be empty")
        return
    }

    if !isAlpha(name) {
        fmt.Printf("invalid command name: %s", name)
    }

    c.CommandList[strings.ToLower(name)] = f
}

func handlerLogin(s *state, cmd command) error {
    if len(cmd.args) == 0 || cmd.args[0] == "" {
        return fmt.Errorf("invalid arguments, username is required")
    }

    UserName := cmd.args[0]
    if isAlphaNumeric(UserName) {
        return fmt.Errorf("username can't contain special characters")
    }

    user, err := s.db.GetUser(context.Background(), UserName)
    if err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return fmt.Errorf("User with %s username does not exist.", UserName)
        }
        return err
    }

    err = s.cfg.SetUser(user.Name)
    if err != nil {
        return err
    }

    fmt.Println("Logged in successfuly!")
    return nil
}

func handlerRegister(s *state, cmd command) error {
    if len(cmd.args) == 0 || cmd.args[0] == "" {
        return fmt.Errorf("invalid arguments, username is required")
    }

    UserName := cmd.args[0]
    if isAlphaNumeric(UserName) {
        return fmt.Errorf("username can't contain special characters")
    }

    args := database.CreateUserParams{
        ID:        int32(uuid.New().ID()),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name:      UserName,
    }
    user, err := s.db.CreateUser(context.Background(), args)
    if err != nil {
        duplicateErrorCode := "23505"
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pqerror.Code(duplicateErrorCode) {
            return fmt.Errorf("User with %s username already exists. Error: %v", UserName, err)
        }
        return err
    }
    fmt.Printf("User %s successfully created.\n", user.Name)
    s.cfg.SetUser(user.Name)

    return nil
}

func handleReset(s *state, cmd command) error {
    err := s.db.ResetUsers(context.Background())
    if err != nil {
        return fmt.Errorf("Error while reseting table \"users\": %v", err)
    }

    return nil
}

func isAlpha(s string) bool {
    if s == "" { return false }
    for _, r := range s {
        if !unicode.IsLetter(r) { return false }
    }

    return true
}

func isAlphaNumeric(s string) bool {
    if s == "" { return false }
    for _, r := range s {
        if !unicode.IsLetter(r) || !unicode.IsNumber(r) { return false }
    }

    return true
}
