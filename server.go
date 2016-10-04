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
	"sort"
)

var htmlPath, _ = filepath.Abs("html")

// defListenPort - default port for the service to listen on
const defListenPort = 8081

var GitClient *github.Client

type RenderedPull struct {
	RepoName *string
	UserName *string
	Age int
	Title *string
}

type RenderedPulls []RenderedPull


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

func (slice RenderedPulls) Len() int {
	return len(slice)
}

func (slice RenderedPulls) Less(i, j int) bool {
	return slice[i].Age < slice[j].Age;
}

func (slice RenderedPulls) Swap(i, j int) {
	slice[i], slice[j] = slice[j], slice[i]
}

func HandleHome (w http.ResponseWriter, r *http.Request) {

	var RenderedPullReqs RenderedPulls

	var repoOpts = &github.RepositoryListByOrgOptions{Type:"private"}
	var pullOpts = &github.PullRequestListOptions{}

	t := template.Must(template.New("index.html").ParseFiles(path.Join(htmlPath, "index.html")))

	repos, _, err := GitClient.Repositories.ListByOrg("wrapp", repoOpts)
	if err != nil {
		log.Panic(err)
	}

	for r := range repos {
		//We don't care about forks
		if *repos[r].Fork != true {
			pullList, _, err := GitClient.PullRequests.List("wrapp", *repos[r].Name, pullOpts)
			if err != nil {
				log.Panic(err)
			}

			for p := range pullList {
				RenderedPullReqs = append(RenderedPullReqs, RenderedPull{
					RepoName: repos[r].Name,
					UserName: pullList[p].User.Login,
					Age: int(time.Now().Sub(*pullList[p].CreatedAt).Hours()),
					Title: pullList[p].Title,

				})}
		}
	}
	if err != nil {
		log.Panic(err)
	}

	sort.Sort(RenderedPullReqs)

	p := RenderedPullReqs
	t.Execute(w, p)
}