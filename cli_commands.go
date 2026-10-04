package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/TeamDiamond524/Gator/internal/config"
	"github.com/TeamDiamond524/Gator/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

// Error codes from database
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

func handlerAggregator(s *state, cmd command) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Command expects duration between fetchs: agg <timer>")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("Error with parsing time: %v", err)
	}

	fmt.Printf("Collecting feeds every %s\n", timeBetweenRequests.String())

	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		scrapefeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("addfeed expects name and url argument:  aggfeed <name> <url>")
	}

	Name := cmd.args[0]
	URL := cmd.args[1]

	FeedArgs := database.CreateFeedParams{
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      Name,
		Url:       URL,
		UserID:    user.ID,
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
		ID:        int32(uuid.New().ID()),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
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

func handlerBrowse(s *state, cmd command) error {
    if len(cmd.args) != 0 && len(cmd.args) != 2 {
        return fmt.Errorf("invalid parameters. Command definition: browse {--limit <number>}")
    }

    var limit int
    if len(cmd.args) == 2 {
        if cmd.args[0] != "--limit"{
            return fmt.Errorf("%s optional parameter dows not exist. Command definition: browse {--limit <number>}", cmd.args[0])
        }

        var err error
        limit, err = strconv.Atoi(cmd.args[1])
        if err != nil {
            limit = 2
        }
    } else {
        limit = 2
    }

    user, err := s.db.GetUser(context.Background(), s.cfg.Current_user_name)
    if err != nil {
        return err
    }

    args := database.GetPostsForUserParams{
        UserID: user.ID,
        Limit: int32(limit),
    }
    posts, err := s.db.GetPostsForUser(context.Background(), args)
    if err != nil {
        return err
    }

    fmt.Println("Posts: ")
    for _, post := range posts {
        fmt.Printf(" -Title: %s, URL: %s\n", post.Title, post.Url)
    }

    return nil
}

func scrapefeeds(s *state) error {
	feedToFetch, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("Error while getting next feed to fetch: %v", err)
	}

	feedArgs := database.MarkFeedFetchedParams{
		ID:            feedToFetch.ID,
		UpdatedAt:     time.Now(),
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
	}

	if err = s.db.MarkFeedFetched(context.Background(), feedArgs); err != nil {
		return fmt.Errorf("Error while marking feed as fetched: %v", err)
	}

	XMLFeed, err := fetchFeed(context.Background(), feedToFetch.Url)
	if err != nil {
		return fmt.Errorf("Error while fetching feed: %v", err)
	}

	decodeEscapedHTML(XMLFeed)

	for _, item := range XMLFeed.Channel.Item {
        PubDate := sql.NullTime{Valid: false}
		parsedPubDate, err := time.Parse(time.Layout, item.PubDate)
        if err != nil {
            PubDate = sql.NullTime{Time: parsedPubDate, Valid: true}
        }

		post := database.CreatePostParams{
			ID:          int32(uuid.New().ID()),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: PubDate,
            FeedID: feedArgs.ID,
		}

        _, err = s.db.CreatePost(context.Background(), post)
        if err != nil {
            if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pqerror.Code(duplicateErrorCode) {
                break
            }
            return err
        }
	}

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
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}

	return true
}

func isAlphaNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) || !unicode.IsNumber(r) {
			return false
		}
	}

	return true
}
