package main

import (
	"log"
	"os"

	"github.com/codegangsta/cli"
	"github.com/corybuecker/reconfig"
	"github.com/corybuecker/steam-stats-fetcher/database"
	"github.com/corybuecker/steam-stats-fetcher/fetcher"
	"github.com/corybuecker/steam-stats-fetcher/giantbomb"
	"github.com/corybuecker/steam-stats-fetcher/jobs"
	"github.com/corybuecker/steam-stats-fetcher/steam"
	"github.com/corybuecker/steam-stats-fetcher/storage"
	"github.com/dancannon/gorethink"
)

func getDatabase(databaseHost string) (database.Interface, error) {
	session, err := gorethink.Connect(gorethink.ConnectOpts{Address: databaseHost})
	if err != nil {
		return nil, err
	}
	var rethinkDB database.RethinkDB
	rethinkDB = database.RethinkDB{Session: session}
	return &rethinkDB, nil
}

func main() {
	databaseHost := "localhost"

	app := cli.NewApp()
	app.Name = "steam-stats-fetcher"

	var rethinkDB database.Interface
	var err error

	if rethinkDB, err = getDatabase(databaseHost); err != nil {
		log.Fatal(err)
	}
	var steamFetcher = steam.Fetcher{}
	var giantBombFetcher = giantbomb.Fetcher{}

	reconfig.Get(rethinkDB.(*database.RethinkDB).Session, "steam", &steamFetcher)
	reconfig.Get(rethinkDB.(*database.RethinkDB).Session, "steam-stats", &giantBombFetcher)
	var job = &jobs.Job{Fetcher: &fetcher.JSONFetcher{}, Database: rethinkDB}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host",
			Value:       "localhost",
			Usage:       "connection host for RethinkDB",
			Destination: &databaseHost,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:  "setup",
			Usage: "create needed database and tables",
			Action: func(c *cli.Context) {
				storage.Setup(rethinkDB, "videogames", []string{"ownedgames", "giantbomb"})
			},
		},

		{
			Name:  "steam",
			Usage: "update all owned games from steam",
			Action: func(c *cli.Context) {
				job.OwnedGamesFetch(&steamFetcher)
			},
		},

		{
			Name:  "search",
			Usage: "search for the name of all owned games in GiantBomb",
			Action: func(c *cli.Context) {
				job.OwnedGamesSearch(&steamFetcher, &giantBombFetcher)
			},
		},

		{
			Name:  "fetch",
			Usage: "fetch all known games from GiantBomb",
			Action: func(c *cli.Context) {
				job.OwnedGamesFetchByID(&steamFetcher, &giantBombFetcher)
			},
		},
	}
	app.Run(os.Args)
}
