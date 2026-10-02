package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/TeamDiamond524/Gator/internal/config"
	"github.com/TeamDiamond524/Gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
    var State state
    var err error
    State.cfg, err = config.Read()
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    db, err := sql.Open("postgres", State.cfg.DB_url)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }

    State.db = database.New(db)

    Commands := commands{
        CommandList: make(map[string]func(*state, command) error),
    }

    Commands.register("login",     handlerLogin)
    Commands.register("register",  handlerRegister)
    Commands.register("reset",     handlerReset)
    Commands.register("users",     handlerListUsers)
    Commands.register("agg",       handlerAggegator)
    Commands.register("addfeed",   middlewareLoggedIn(handlerAddFeed))
    Commands.register("feeds",     handlerFeeds)
    Commands.register("follow",    middlewareLoggedIn(handlerFollow))
    Commands.register("following", middlewareLoggedIn(handlerFollowing))

    args := os.Args
    if len(args) < 2 {
        fmt.Println("not enough arguments passed")
        os.Exit(1)
    }

    CommandName := args[1]
    CommandArgs := make([]string, 0)
    if len(args) > 2 {
        CommandArgs = append(CommandArgs, args[2:]...)
    }

    Command := command{ name: CommandName, args: CommandArgs,}

    err = Commands.run(&State, Command)
    if err != nil {
        fmt.Println(err)
        os.Exit(1)
    }
}
