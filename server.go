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
	"time"
)

var htmlPath, _ = filepath.Abs("html")

// defListenPort - default port for the service to listen on
const defListenPort = 8081

var fMap = template.FuncMap{
	"formatAsDate": func (t time.Time) string {

	/* Layouts must use the reference time Mon Jan 2 15:04:05 MST 2006 to show
	the pattern with which to format/parse a given time/string.

	 PHP-level stupidity here.
	 */
	return t.Format("2006-01-02")
	},
}

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
	muxer := mux.NewRouter()

	// Generic handler for users visiting the homepage
	muxer.HandleFunc("/", HandleHome).Methods("GET")

	log.Println("Starting a service listening on port", defListenPort)
	//Set up a listening HTTP server on specified port
	log.Panic(http.ListenAndServe(fmt.Sprintf(":%d", defListenPort), muxer))
}

func HandleHome (w http.ResponseWriter, r *http.Request) {

	type PullReqs struct{
		RepoName string
		Request *github.PullRequest
	}

	type NamedPullRepos struct {
		PullRepos []PullReqs
	}

	var pulls []PullReqs

	repoOpts := &github.RepositoryListByOrgOptions{Type:"private"}
	repos, _, err := GitClient.Repositories.ListByOrg("wrapp", repoOpts)
	if err != nil {
		log.Panic(err)
	}

	for r := range repos {
		//We don't care about forks
		if *repos[r].Fork != true {
			pullOpts := &github.PullRequestListOptions{}
			pullList, _, err := GitClient.PullRequests.List("wrapp", *repos[r].Name, pullOpts)
			if err != nil {
				log.Panic(err)
			}
			for p := range pullList {
				pulls = append(pulls, PullReqs{RepoName: *repos[r].Name, Request: pullList[p]})
			}

			var repoNames []string
			repoNames = append(repoNames, *repos[r].Name)
			log.Print(repoNames)
		}
	}
	if err != nil {
		log.Panic(err)
	}


	t := template.Must(template.New("index.html").Funcs(fMap).ParseFiles(path.Join(htmlPath, "index.html")))
	p := NamedPullRepos{PullRepos: pulls}
	t.Execute(w, p)
}