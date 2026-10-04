# GATOR

This is guided project from [boot.dev](https://www.boot.dev/lessons/14b7179b-ced3-4141-9fa5-e67dbc3e5242).
The goap of this project is to teach myself how to use postgres database and cli commands.

## Prerequisites
This program expects you to already have [PostgresSQL](https://www.postgresql.org/download/) and [goose](https://github.com/pressly/goose) migration tool.

## Installation

Firstly, we have to install the package. You can do that in terminal like so: 
```bash
go install github.com/TeamDiamond524/Gator```

## Setup
After the installation we need to set up database. Using PostgreSQL create database named `gator`.
```psql
CREATE DATABASE gator;
```

We will aslo set up the config file `~/.gatorconfig.json`:
```json
{
    "db_url":"postgres://<user>:<password>@localhost:5432/gator?sslmode=disable"
}
```

After we will use database migration tool [`goose`](https://github.com/pressly/goose).
First clone the repository:
```git
git clone github.com/TeamDiamond524/Gator
```

Then navigate to `sql/schema`. In this directory we will run the `goose` migration tool using `db_url` from `~/.gatorconfig.json`:
```bash
goose postgres <db_url> up
```

## CLI Commands
You access the program via `Gator` command.

List of commands:
- `Gator login <user>` -- logins already existing user from database
- `Gator register <user>` -- registers new user to the program
- `Gator reset` -- deletes all data from database (I do not recommend using it)
- `Gator users` -- displays all registered users
- `Gator agg <timer>` -- sets up the loop at which the program fetchs the data from the internet
- `Gator addfeed <title> <url>` -- add the feed source to the database and it automaticly follows
- `Gator feeds` -- displays all feeds and user that created it from database
- `Gator follow <url>` -- follows the given feed url for the current user
- `Gator following` -- displays all following feed for the current user
- `Gator unfollow <user>` -- unfollows the given feed ulr for the current user
- `Gator browse {--limit <number>}` -- displays posts from feeds the current user follows. If limit is not given, it defaults to 2

## Usage
When using program you need to run
```bash
Gator agg <timer>
```
in seperate terminal.

## Example
```bash
#Seperate terminal
Gator agg 5m

#Main terminal

#We register user td524
Gator register td524

#We register another user user2
Gator register user2

#We login back as user tc524
Gator login td524

#We add some feed to our database
Gator addfeed TechCrunch https://techcrunch.com/feed/
Gator addfeed "Hacker News" https://news.ycombinator.com/rss

#User td524 unfollows from Hacker News
Gator unfollow https://news.ycombinator.com/rss

#Displays posts from following feed. In this case from TechCrunch
Gator browse --limit 5
```
