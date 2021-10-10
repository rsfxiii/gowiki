// +build ignore

package gowiki

import (
    "errors"
    "html/template"
    "io/ioutil"
    "log"
    "regexp"
    "net/http"
)

var templates = template.Must(template.ParseFiles("edit.html", "view.html"))
// Determine valid, expected paths - use to avoid a user reading/writing elsewhere
var validPath = regexp.MustCompile("^/(edit|save|view)/([a-zA-Z0-9]+)$")

type Page struct {
    Title string
    Body []byte
}

func (p *Page) save() error {
    filename := p.Title + ".txt"
    return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage( title string) (*Page, error) {
    filename := title + ".txt"
    body, err := ioutil.ReadFile(filename)
    if err != nil {
        return nil, err
    }
    return &Page{Title: title, Body: body}, nil
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
    page, err := loadPage(title)
    if err != nil {
        page = &Page{Title: title}
    }
    renderTemplate(w, "edit", page)
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
    page, err := loadPage(title)
    if err != nil {
        http.Redirect(w, r, "/edit/"+title, http.StatusFound)
        return
    }
    renderTemplate(w, "view", page)
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
    err := templates.ExecuteTemplate(w, tmpl+".html", p)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
    body := r.FormValue("body")
    page := &Page{Title: title, Body: []byte(body)}
    err := page.save()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

func makeHandler(fn func (http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        pathMatch := validPath.FindStringSubmatch(r.URL.Path)
        if pathMatch == nil {
            http.NotFound(w, r)
            return
        }
        fn(w, r, pathMatch[2])
    }
}

func getTitle(w http.ResponseWriter, r *http.Request) (string, error) {
    pathMatch := validPath.FindStringSubmatch(r.URL.Path)
    if pathMatch == nil {
        http.NotFound(w, r)
        return "", errors.New("invalid Page Title")
    }
    return pathMatch[2], nil // The title is the second subexpression.
}

func main() {
    http.HandleFunc("/view/", makeHandler(viewHandler))
    http.HandleFunc("/edit/", makeHandler(editHandler))
    http.HandleFunc("/save/", makeHandler(saveHandler))
    log.Fatal(http.ListenAndServe(":8080", nil))
}
