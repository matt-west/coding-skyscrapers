package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"text/template"
	"os"
	"strconv"
	"time"
)

type Config struct {
	URL         string
	Title       string
	Description string
	Lang        string
	Editor      string
	Webmaster   string
}

type Page struct {
	Slug        string
	Title       string
	Keywords    string
	Description string
}

type Post struct {
	Title       string
	Slug        string
	Date        string
	MachineDate string
	Keywords    string
	Description string
	Content     string
	Tags        []string
}

type Tag struct {
	Title string
	Posts map[string]*Post
}

type Sidebar struct {
	Recent []Post
	Tags   map[string]*Tag
	Pages  map[string]*Page
}

type RSS struct {
	Config *Config
	Posts  []Post
}

type Sitemap struct {
	Config *Config
	Pages  []Page
	Posts  []Post
}

const assetPath = len("/")
const pagePath = len("/page/")
const tagPath = len("/tag/")
const postPath = len("/")

const maxPosts = 10 // Number posts to display on homepage

// Config
var config = new(Config)

// Pages
var pages = make(map[string]*Page)
var pagesJSON []Page
var pageTemplates = make(map[string]*template.Template)

// Posts
var posts = make(map[string]*Post)
var postsJSON []Post // Need this so that there is an ordered list of posts

// Templates
var layoutTemplates *template.Template
var errorTemplates *template.Template
var rssTemplate *template.Template
var sitemapTemplate *template.Template
var sidebarAssets *Sidebar

// Tags
var tags = make(map[string]*Tag)

// Static Assets i.e. Favicons or Humans.txt
var staticAssets = []string{"humans.txt", "favicon.ico"}

// Init Function to Load Template Files and JSON Dict to Cache
func init() {

	log.Println("Loading Config")
	loadConfig()

	log.Println("Loading Templates")
	loadTemplates()

	log.Println("Loading Pages")
	loadPages()

	log.Println("Loading Posts")
	loadPosts()

	log.Println("Loading Tags")
	loadTags()

	n := 5

	if len(postsJSON) < 5 {
		n = len(postsJSON)
	}

	sidebarAssets = &Sidebar{postsJSON[0:n], tags, pages}
}

// Load the Config File (config/app.json)
func loadConfig() {
	configRaw, _ := ioutil.ReadFile("config/app.json")
	err := json.Unmarshal(configRaw, config)

	if err != nil {
		panic("Could not parse config file!")
	}
}

// Load The Tags Map
func loadTags() {
	for i := 0; i < len(postsJSON); i++ {

		for t := 0; t < len(postsJSON[i].Tags); t++ {

			_, ok := tags[postsJSON[i].Tags[t]]
			if ok {
				tags[postsJSON[i].Tags[t]].Posts[postsJSON[i].Title] = &postsJSON[i]
			} else {
				tagPosts := make(map[string]*Post)
				tagPosts[postsJSON[i].Title] = &postsJSON[i]
				tags[postsJSON[i].Tags[t]] = &Tag{postsJSON[i].Tags[t], tagPosts}
			}

		}
	}
}

// Load Pages Dict and Templates
func loadPages() {
	pagesRaw, _ := ioutil.ReadFile("data/pages.json")
	err := json.Unmarshal(pagesRaw, &pagesJSON)
	if err != nil {
		panic("Could not parse Pages JSON!")
	}

	for i := 0; i < len(pagesJSON); i++ {
		pages[pagesJSON[i].Slug] = &pagesJSON[i]
	}

	for _, tmpl := range pages {
		t := template.Must(template.ParseFiles("./pages/" + tmpl.Slug + ".html"))
		pageTemplates[tmpl.Slug] = t
	}
}

// Load Posts Dict and Templates
func loadPosts() {
	postsRaw, _ := ioutil.ReadFile("data/posts.json")
	err := json.Unmarshal(postsRaw, &postsJSON)
	if err != nil {
		panic("Could not parse Posts JSON!")
	}

	for i := 0; i < len(postsJSON); i++ {
		slug := postsJSON[i].Slug
		posts[slug] = &postsJSON[i]

		// Read the post content file.
		b, _ := ioutil.ReadFile("./posts/" + slug + ".html")
		posts[slug].Content = string(b)
	}
}

// Load Layout and Error Templates
func loadTemplates() {
	layoutTemplates = template.Must(template.ParseFiles("./templates/layouts.html"))
	errorTemplates = template.Must(template.ParseFiles("./templates/errors/404.html", "./templates/errors/505.html"))
	rssTemplate = template.Must(template.ParseFiles("./templates/rss.xml"))
	sitemapTemplate = template.Must(template.ParseFiles("./templates/sitemap.xml"))
}

// Page Handler Constructs and Serves Pages
func pageHandler(w http.ResponseWriter, r *http.Request) {

	// Get the page slug, use 'index' if no slug is present
	slug := r.URL.Path[pagePath:]
	if slug == "" {
		indexHandler(w, r)
		return
	}

	// Check that the page exists and return 404 if it doesn't
	_, ok := pages[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		errorTemplates.ExecuteTemplate(w, "404", nil)
		return
	}

	// Find the page
	p := pages[slug]

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Page Template
	err := pageTemplates[slug].Execute(w, nil)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorTemplates.ExecuteTemplate(w, "505", nil)
		return
	}

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", nil)
}

// Post Handler
func postHandler(w http.ResponseWriter, r *http.Request) {
	// Check to see if the request is after a static asset
	for _, asset := range staticAssets {
		if asset == r.URL.Path[1:] {
			http.ServeFile(w, r, asset)
			return
		}
	}

	// Get the post slug, use 'index' if no slug is present
	slug := r.URL.Path[postPath:]
	if slug == "" {
		indexHandler(w, r)
		return
	}

	// Check that the post exists and return 404 if it doesn't
	_, ok := posts[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		errorTemplates.ExecuteTemplate(w, "404", nil)
		return
	}

	// Find the post
	p := posts[slug]

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Post
	err := layoutTemplates.ExecuteTemplate(w, "Post", p)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		errorTemplates.ExecuteTemplate(w, "505", nil)
		return
	}

	// Comments
	layoutTemplates.ExecuteTemplate(w, "Comments", nil)

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", nil)
}

// Asset Handler Serves CSS, JS and Images
func assetHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, r.URL.Path[assetPath:])
}

func archiveHandler(w http.ResponseWriter, r *http.Request) {
	p := Page{"archive", "Archive", "", ""}

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Archives
	layoutTemplates.ExecuteTemplate(w, "Archive", postsJSON)

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", p)
}

func tagHandler(w http.ResponseWriter, r *http.Request) {
	slug := r.URL.Path[tagPath:]
	if slug == "" {
		indexHandler(w, r)
		return
	}

	// Check that the tag exists and return 404 if it doesn't
	_, ok := tags[slug]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		errorTemplates.ExecuteTemplate(w, "404", nil)
		return
	}

	p := &Page{"/tag/" + slug, "Posts Tagged #" + slug, "", ""}

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	for _, tmpl := range tags[slug].Posts {
		// Post
		layoutTemplates.ExecuteTemplate(w, "Post", posts[tmpl.Slug])
	}

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", p)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	p := pages["index"]

	// Header
	layoutTemplates.ExecuteTemplate(w, "Header", p)

	// Sidebar
	layoutTemplates.ExecuteTemplate(w, "Sidebar", sidebarAssets)

	// Show Recent Posts
	for i, tmpl := range postsJSON {
		if i >= maxPosts {
			break
		}
		
		// Post
		layoutTemplates.ExecuteTemplate(w, "Post", posts[tmpl.Slug])
	}

	// Footer
	layoutTemplates.ExecuteTemplate(w, "Footer", p)
}

func rssHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/atom+xml; charset=utf-8")
	rss := RSS{config, postsJSON}
	rssTemplate.Execute(w, rss)
}

func sitemapHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/xml")
	sitemap := Sitemap{config, pagesJSON, postsJSON}
	sitemapTemplate.Execute(w, sitemap)
}

// Starts Server and Routes Requests
func main() {

	if len(os.Args) < 2 {
		log.Println("Available Commands: start stop restart")
		os.Exit(1)
	}

	if os.Args[1] == "start" {
		startServer()
	} else if os.Args[1] == "stop" {
		stopServer()
		os.Exit(0)
	} else if os.Args[1] == "restart" {
		restartServer()
	} else {
		log.Println("Available Commands: start stop restart")
		os.Exit(1)
	}

}

func startServer() {
	log.Println("Starting: " + config.Title)

	http.HandleFunc("/archive", archiveHandler)
	http.HandleFunc("/page/", pageHandler)
	http.HandleFunc("/tag/", tagHandler)
	http.HandleFunc("/assets/", assetHandler)
	http.HandleFunc("/rss", rssHandler)
	http.HandleFunc("/sitemap", sitemapHandler)
	http.HandleFunc("/", postHandler)

	log.Println("PID: " + strconv.Itoa(os.Getpid()))
	ioutil.WriteFile("tmp/go-blog.pid", []byte(strconv.Itoa(os.Getpid())), 0600)

	err := http.ListenAndServe(":9981", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func stopServer() {
	b, _ := ioutil.ReadFile("tmp/go-blog.pid")
	pid, _ := strconv.Atoi(string(b))

	p, _ := os.FindProcess(pid)
	log.Println("Stopping process " + string(b))
	err := p.Kill()

	if err != nil {
		log.Println("Could not stop process")
	}
}

func restartServer() {
	stopServer()
	time.Sleep(1000 * time.Millisecond) // Wait 1 second
	startServer()
}
