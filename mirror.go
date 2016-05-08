package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"

	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/codeskyblue/go-sh"
	"github.com/phayes/hookserve/hookserve"
)

var (
	version = "head" // set by command-line on CI release builds
	app     = kingpin.New("git-mirror", "Push to gitlab on github hook")

	cacheDir   = app.Flag("cachedir", "Directory in which to mirror (bare) repositories").Default("/tmp/git-mirror").String()
	gitlabHost = app.Flag("gitlabhost", "Host where to push your repositories").Default("git.itch.ovh").String()
	webPort    = app.Flag("port", "Port to listen on").Default("6298").Int()
	webPath    = app.Flag("path", "Path to respond to (other paths will yield a 404)").Default("/.git-mirror").String()
	secret     = app.Flag("secret", "GitHub secret").String()
)

func main() {
	app.HelpFlag.Short('h')
	app.Version(version)
	app.VersionFlag.Short('V')
	kingpin.MustParse(app.Parse(os.Args[1:]))

	server := hookserve.NewServer()
	server.IgnoreTags = false
	server.Port = *webPort
	server.Path = *webPath
	server.Secret = *secret
	server.GoListenAndServe()

	err := os.MkdirAll(*cacheDir, os.FileMode(0755))
	if err != nil {
		log.Fatal(err.Error())
	}

	log.Print("Now listening for GitHub hooks on port ", server.Port)

	for event := range server.Events {
		fmt.Println(event.Owner + " " + event.Repo + " " + event.Branch + " " + event.Commit)
		fullName := fmt.Sprintf("%s/%s", event.Owner, event.Repo)
		cloneDir := path.Join(*cacheDir, filepath.FromSlash(fullName))
		gitlabRemote := fmt.Sprintf("git@%s:%s.git", *gitlabHost, fullName)
		githubRemote := fmt.Sprintf("git@github.com:%s.git", fullName)

		cloneSess := sh.NewSession().SetDir(*cacheDir)
		cloneSess.ShowCMD = true
		cloneSess.Command("git", "clone", "--mirror", githubRemote, cloneDir)
		err := cloneSess.Run()
		if err != nil {
			// totally expected if has been mirroring before
			fmt.Printf("clone error: %s\n", err.Error())
		}

		fetchSess := sh.NewSession().SetDir(cloneDir)
		fetchSess.ShowCMD = true
		fetchSess.Command("git", "fetch", "--prune")
		err = fetchSess.Run()
		if err != nil {
			fmt.Printf("fetch error: %s\n", err.Error())
		}

		pushSess := sh.NewSession().SetDir(cloneDir)
		pushSess.ShowCMD = true
		pushSess.Command("git", "push", "--mirror", gitlabRemote)
		err = pushSess.Run()
		if err != nil {
			fmt.Printf("push error: %s\n", err.Error())
		}
	}
}
