package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Article struct {
	Id      string `json:"Id"`
	Title   string `json:"Title"`
	Desc    string `json:"Desc"`
	Content string `json:"Content"`
}

var Articles []Article

// Id, err := ParseIdFromURL("/5/create")

func ParseIdFromURL(URL string) (id string, err error) {
	re := regexp.MustCompile(`/([0-9]+)`)
	Id := re.FindStringSubmatch(URL)
	if len(Id) != 2 {
		return id, fmt.Errorf("wrong ID")
	}
	return Id[1], nil
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: homePage")

	resp, err := http.Get("http://localhost:10000/articles")
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	err = json.Unmarshal(body, &Articles)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	for _, Article := range Articles {
		fmt.Fprintf(w, "<ul> <a href=\"/%s\">%s</a><br>", Article.Id, Article.Title)
		fmt.Fprintf(w, "<li> %s </li></ul><br>", Article.Desc)
	}
	fmt.Fprintf(w, "<a href=\"/create\">Create article</a>")
}

func returnArticles(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: returnAllArticles")

	Id, err := ParseIdFromURL(r.URL.Path)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(Id)
	resp, err := http.Get("http://localhost:10000/articles/" + Id)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	err = json.Unmarshal(body, &Articles)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	for _, Article := range Articles {
		fmt.Fprintf(w, "<a href=\"/\">Back to main page</a>")
		fmt.Fprintf(w, "<ul> %s <br>", Article.Title)
		fmt.Fprintf(w, "<li> %s </li>", Article.Desc)
		fmt.Fprintf(w, "<li> %s </li></ul><br>", Article.Content)
		fmt.Fprintf(w, "<a href=\"/%s/delete\">Delete article</a><br>", Article.Id)
		fmt.Fprintf(w, "<a href=\"/%s/update\">Edit article</a><br>", Article.Id)
	}
}

func newArticleGET(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Endpoint hit: newArticleGET")

	fmt.Fprintf(w, "<form method=\"post\" action=\"/create\"> <label> Title </label> <input type=\"text\" name=\"title\"/> <br>")
	fmt.Fprintf(w, "<label> Description </label> <input type=\"text\" name=\"description\"/> <br>")
	fmt.Fprintf(w, "<label> Content </label> <textarea name=\"content\" rows=\"5\" cols=\"80\"> </textarea> <br>")

	fmt.Fprintf(w, "<input type=\"submit\" value=\"Send\" /> </form>")
}

func newArticlePOST(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	Title := ""
	Desc := ""
	Content := ""

	for key, value := range r.Form {
		fmt.Printf("%s = %s\n", key, value)
		if key == "title" {
			Title = value[0]
		}
		if key == "description" {
			Desc = value[0]
		}
		if key == "content" {
			Content = value[0]
		}
	}
	postBody, _ := json.Marshal(map[string]string{
		"Title":   Title,
		"Desc":    Desc,
		"Content": Content,
	})
	body := bytes.NewBuffer(postBody)

	fmt.Println(Title, Desc, Content)

	resp, err := http.Post("http://localhost:10000/articles/create", "application/json", body)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s", err)
	}
	defer resp.Body.Close()
	apiResponseBody, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	err = json.Unmarshal(apiResponseBody, &Articles)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	w.Header().Set("Location", "http://localhost:9999/")
	w.WriteHeader(301)
}

func deleteArticle(w http.ResponseWriter, r *http.Request) {

	Id, err := ParseIdFromURL(r.URL.Path)
	if err != nil {
		fmt.Println(err)
		return
	}

	client := &http.Client{}

	req, err := http.NewRequest("DELETE", "http://localhost:10000/articles/"+Id, nil)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s", err)
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()
	_, err = io.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err)
	}

	w.Header().Set("Location", "http://localhost:9999/")
	w.WriteHeader(301)
}

func updateArticleGET(w http.ResponseWriter, r *http.Request) {

	Id, err := ParseIdFromURL(r.URL.Path)
	if err != nil {
		fmt.Println(err)
		return
	}

	resp, err := http.Get("http://localhost:10000/articles/" + Id)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	err = json.Unmarshal(body, &Articles)

	if err != nil {
		fmt.Printf("ERROR: %s", err)
	}

	if len(Articles) == 0 {
		fmt.Fprintf(w, "Article with ID %s doesn't exist", Id)
		return
	}

	fmt.Fprintf(w, "<form method=\"post\" action=\"/"+Id+"/update\"> <label> Title </label> <input type=\"text\" value=\""+Articles[0].Title+"\" name=\"title\"/> <br>")
	fmt.Fprintf(w, "<label> Description </label> <input type=\"text\"  value=\""+Articles[0].Desc+"\" name=\"description\"/> <br>")
	fmt.Fprintf(w, "<label> Content </label> <textarea name=\"content\" rows=\"5\" cols=\"80\">"+Articles[0].Content+"</textarea> <br>")

	fmt.Fprintf(w, "<input type=\"submit\" value=\"Send\" /> </form>")
}

func updateArticlePOST(w http.ResponseWriter, r *http.Request) {

	Id, err := ParseIdFromURL(r.URL.Path)
	if err != nil {
		fmt.Println(err)
		return
	}

	r.ParseForm()
	fmt.Println(r.Form)

	Title := ""
	Desc := ""
	Content := ""

	for key, value := range r.Form {
		fmt.Printf("%s = %s\n", key, value)
		if key == "title" {
			Title = value[0]
		}
		if key == "description" {
			Desc = value[0]
		}
		if key == "content" {
			Content = value[0]
		}
	}
	postBody, _ := json.Marshal(map[string]string{
		"Title":   Title,
		"Desc":    Desc,
		"Content": Content,
	})
	body := bytes.NewBuffer(postBody)

	fmt.Println(Title, Desc, Content)

	client := &http.Client{}

	req, err := http.NewRequest("PUT", "http://localhost:10000/articles/"+Id, body)
	if err != nil {
		fmt.Println(err)
		fmt.Fprintf(w, "%s", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	}

	defer resp.Body.Close()

	w.Header().Set("Location", "http://localhost:9999/")
	w.WriteHeader(301)
}

func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.Use(commonMiddleware)

	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/create", newArticleGET).Methods("GET")
	myRouter.HandleFunc("/create", newArticlePOST).Methods("POST")
	myRouter.HandleFunc("/{id}/update", updateArticleGET).Methods("GET")
	myRouter.HandleFunc("/{id}/update", updateArticlePOST).Methods("POST")
	myRouter.HandleFunc("/{id}/delete", deleteArticle)
	myRouter.HandleFunc("/{id}", returnArticles)
	log.Fatal(http.ListenAndServe(":9999", myRouter))
}

func commonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "text/html")
		next.ServeHTTP(w, r)
	})
}

func main() {
	fmt.Println("Blog started")

	handleRequests()
}
