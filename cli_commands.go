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

//Error codes from database
const duplicateErrorCode = "23505"

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
    if len(cmd.args) != 1 {
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
    if len(cmd.args) != 1 {
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
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pqerror.Code(duplicateErrorCode) {
            return fmt.Errorf("User with %s username already exists. Error: %v", UserName, err)
        }
        return err
    }
    fmt.Printf("User %s successfully created.\n", user.Name)
    s.cfg.SetUser(user.Name)

    return nil
}

func handlerReset(s *state, cmd command) error {
    err := s.db.ResetUsers(context.Background())
    if err != nil {
        return fmt.Errorf("Error while reseting table \"users\": %v", err)
    }

    fmt.Printf("Database successfully reset.\n")
    return nil
}

func handlerListUsers(s *state, cmd command) error {
    userList, err := s.db.GetUsers(context.Background())
    if err != nil {
        return err
    }

    for _, username := range userList {
        if s.cfg.Current_user_name == username {
            fmt.Printf("* %s (current)\n", username)
        } else {
            fmt.Printf("* %s\n", username)
        }
    }

    return nil
}

func handlerAggegator(s *state, cmd command) error {
    URL := "https://www.wagslane.dev/index.xml"
    XMLData, err := fetchFeed(context.Background(), URL)
    if err != nil {
        return err
    }

    decodeEscapedHTML(XMLData)

    fmt.Print(XMLData)

    return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
    if len(cmd.args) != 2 {
        return fmt.Errorf("addfeed expects name and url argument:  aggfeed <name> <url>")
    }

    Name := cmd.args[0]
    URL := cmd.args[1]

    FeedArgs := database.CreateFeedParams {
        ID: int32(uuid.New().ID()),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Name: Name,
        Url: URL,
        UserID: user.ID,
    }

    _, err := s.db.CreateFeed(context.Background(), FeedArgs)
    if err != nil {
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pqerror.Code(duplicateErrorCode) {
            return fmt.Errorf("Feed %s with %s URL already exists. Error: %v", Name, URL, err)
        }
        return fmt.Errorf("Error while creating new feed entry: %v", err)
    }

    fmt.Printf("Feed %s, %s successfully created.\n", Name, URL)
    return middlewareLoggedIn(handlerFollow)(s, command{name: "follow", args: []string{URL}})
}

func handlerFeeds(s *state, cmd command) error {
    if len(cmd.args) != 0 {
        return fmt.Errorf("This command doesn't take any arguments")
    }

    FeedList, err := s.db.GetFeeds(context.Background())
    if err != nil {
        return fmt.Errorf("Error while geting data on feeds from database: %v", err)
    }

    for _, feed := range FeedList {
        fmt.Printf("%s, %s, %s\n", feed.FeedName, feed.Url, feed.User)
    }

    return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
    if len(cmd.args) != 1 {
        return fmt.Errorf("commands explects url argument: follow <url>")
    }

    url := cmd.args[0]

    feed, err := s.db.GetFeed(context.Background(), url)
    if err != nil {
        return fmt.Errorf("Feed with %s url doesn't exist. Error: %v", url, err)
    }

    feedFollow := database.CreateFeedFollowParams{
        ID: int32(uuid.New().ID()),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID: user.ID,
        FeedID: feed.ID,
    }

    if _, err = s.db.CreateFeedFollow(context.Background(), feedFollow); err != nil {
        if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pqerror.Code(duplicateErrorCode) {
            return fmt.Errorf("User %s is already following %s", user.Name, feed.Name)
        }
        return fmt.Errorf("Error while creating new feed entry: %v", err)
    }

    fmt.Printf("User %s successfuly followed feed %s", user.Name, feed.Name)
    return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
    feeds, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
    if err != nil {
        return fmt.Errorf("Error while quering feed followers. Error: %v", err)
    }

    for _, feed := range feeds {
        fmt.Printf("%s %s\n", feed.FeedName, feed.UserName)
    }

    return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
    if len(cmd.args) != 1 {
        return fmt.Errorf("unfollow expects an url: unfollow <url>")
    }

    url := cmd.args[0]

    feed, err := s.db.GetFeed(context.Background(), url)
    if err != nil {
        return err
    }

    args := database.RemoveFeedFollowParams{FeedID: feed.ID, UserID: user.ID}
    err = s.db.RemoveFeedFollow(context.Background(), args) 
    if err != nil {
        return err
    }

    fmt.Printf("Feed %s unfollowed.\n", feed.Name)
    return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
    return func(s *state, cmd command) error {
        user, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
        if err != nil {
            return fmt.Errorf("Error: ")
        }

        return handler(s, cmd, user)
    }
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
