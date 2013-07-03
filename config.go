// 16 september 2012
package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"flag"
	"bufio"
	"code.google.com/p/gopass"
	"net/url"
)

// TODO also update existing files
var configFlag = flag.Bool("config", false, "make new configuration file interactively and exit")

type Config struct {
	SiteName				string
	SiteBaseURL			string
	DBServer				string
	DBUsername			string
	DBPassword			string
	DBDatabase			string
	DBScanboxDatabase	string
	WikiBaseURL			string
	Consoles				Consoles
}

var config Config
var siteBaseURL url.URL		// for cloning for modifications
var wikiBaseURL url.URL

func loadConfig(file string) {
	f, err := os.Open(file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening configuration file %s: %v\n", file, err)
		os.Exit(1)
	}
	defer f.Close()
	jsondata, err := ioutil.ReadAll(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading configuration file: %v\n", err)
		os.Exit(1)
	}
	err = json.Unmarshal(jsondata, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing configuration file: %v\n", err)
		os.Exit(1)
	}

	notSpecified := func(what string) {
		fmt.Fprintf(os.Stderr, "error: %s is not specified in the configuration file\n", what)
		os.Exit(1)
	}
	invalid := func(what string, err error) {
		fmt.Fprintf(os.Stderr, "error: invalid %s specified in the configuration file: %v\n", what, err)
		os.Exit(1)
	}

	if config.SiteName == "" {
		notSpecified("site name")
	}
	if config.SiteBaseURL == "" {
		notSpecified("site base URL")
	}
	u, err := url.Parse(config.SiteBaseURL)
	if err != nil {
		invalid("wiki base URL", err)
	}
	siteBaseURL = *u
	if config.DBServer == "" {
		notSpecified("database unix socket pathname")
	}
	if config.DBUsername == "" {
		notSpecified("database server username")
	}
	if config.DBPassword == "" {
		notSpecified("database server password")
	}
	if config.DBDatabase == "" {
		notSpecified("database name")
	}
	if config.DBScanboxDatabase == "" {
		notSpecified("scanbox database name")
	}
	if config.WikiBaseURL == "" {
		notSpecified("wiki base URL")
	}
	u, err = url.Parse(config.WikiBaseURL)
	if err != nil {
		invalid("wiki base URL", err)
	}
	wikiBaseURL = *u

	// otherwise we're all good
}

func makeConfig(file string) {
	var stdin = bufio.NewReader(os.Stdin)
	var err error

	readline_how := func(what string, how func(string) (string, error)) (entry string) {
		prompt := "Enter " + what + "\n> "
		for {
			entry, err = how(prompt)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error reading line from standard input: %v\n", err)
				os.Exit(1)
			}
			if entry != "" {
				break
			}
			fmt.Println("Sorry, you must enter a value.")
		}
		return
	}
	readline := func(what string) string {
		return readline_how(what, func(prompt string) (string, error) {
			fmt.Print(prompt)
			entry, err := stdin.ReadString('\n')
			if err == nil {
				entry = entry[:len(entry) - 1]		// strip \n
			}
			return entry, err
		})
	}
	readpassword := func(what string) string {
		return readline_how(what, gopass.GetPass)
	}

	config.SiteName = readline(`name of the scan catalogue website (for example, "Sega Retro Scan Information"`)
	config.SiteBaseURL = readline(`base URL of the scan catalogue website (for example, http://andlabs.sonicretro.org/scans/"`)
	config.DBServer = readline(`database unix socket pathname, form [host]:[path] (for example, /path/to/socket)`)
	config.DBUsername = readline(`database server username`)
	config.DBPassword = readpassword(`database server password (will not be echoed)`)
	config.DBDatabase = readline(`database to use; this is the name you chose when you set up MediaWiki (for example, wiki_db)`)	// TODO need better example
	config.DBScanboxDatabase = readline(`database to use for scanboxes: this is the name you chose when you set up KwikiData (for example, wikidata)`)		// TODO need better example
	config.WikiBaseURL = readline(`base URL of wiki pages (for game page links; for example, http://segaretro.org/)`)

	jsondata, err := json.MarshalIndent(config, "", "\t")
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating configuration file data: %v\n", err)
		os.Exit(1)
	}
	f, err := os.OpenFile(		// use OpenFile so we can give it permissions rw------- (0600)
		file, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating configuration file %s: %v\n", file, err)
		os.Exit(1)
	}
	n, err := f.Write(jsondata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error writing configuration fo file: %v\n", err)
		os.Exit(1)
	}
	if n != len(jsondata) {
		fmt.Fprintf(os.Stderr, "somehow did not write the entire configuration file (%d of %d bytes), yet no error was returned\n", n, len(jsondata))
		os.Exit(1)
	}
	f.Write([]byte("\n"))		// end file on blank line
	f.Close()

	// TODO write sample category parameters, especially since I have a simple comment syntax

	// TODO adjust to talk about additional parameters if we move them out
	fmt.Printf(`The configuration file %s has been created successfully.
Relaunch %s, passing just the filename as a parameter.
You may wish to write-protect the file first
(for instance, on Unix systems, with chmod -w %s).
`,
		file, os.Args[0], file)
}
