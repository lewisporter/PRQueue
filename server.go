package main

import (
	"log"
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"html/template"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

var htmlPath, _ = filepath.Abs("html")

// defListenPort - default port for the service to listen on
const defListenPort = 8081

var GitClient *github.Client

func init() {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: "df59a008a3bf575d3c3fb13a16c5a6e6becd4401"},
	)
	tc := oauth2.NewClient(oauth2.NoContext, ts)

	GitClient = github.NewClient(tc)
}

func main() {

	log.Println("Setting Routes...")
	// Create a new URL multiplexer (mux) to handle routing
	// While this could have been done with the standard http library, it wouldn't allow for regex to be used,
	// so we shall use Mozilla's mux package
	muxer := mux.NewRouter()

	// Generic handler for users visiting the homepage
	muxer.HandleFunc("/", HandleHome).Methods("GET")

	log.Println("Starting a service listening on port", defListenPort)
	//Set up a listening HTTP server on specified port
	log.Panic(http.ListenAndServe(fmt.Sprintf(":%d", defListenPort), muxer))
}

func HandleHome (w http.ResponseWriter, r *http.Request) {
	repoOpts := &github.RepositoryListByOrgOptions{}
	repos, _, err := GitClient.Repositories.ListByOrg("wrapp", repoOpts)
	if err != nil {
		log.Panic(err)
	}

	var pulls = make(map[string][]*github.PullRequest)

	for r := range repos {
		pullOpts := &github.PullRequestListOptions{}
		pull, _, err := GitClient.PullRequests.List("wrapp", *repos[r].Name, pullOpts)
		if err != nil {
			log.Panic(err)
		}
		pulls[*repos[r].Name] = pull
	}
	if err != nil {
		log.Panic(err)
	}

	log.Print(pulls)

	t, err := template.ParseFiles(path.Join(htmlPath, "index.html"))
	if err != nil {
		// If we can't parse the template, something is dangerously broken
		log.Panic(err)
	}
	p := map[string]interface{}{"pullRequests": pulls}
	t.Execute(w, p)
}