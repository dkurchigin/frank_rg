package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type ViewData struct{
	Title string
	Error string
	Files []FileElement
}

type FileElement struct {
	Name string
	IsFile bool
}

func getFilesAndDirs(path string) []FileElement {
	var fileList []FileElement

	lst, err := ioutil.ReadDir(path)
	if err != nil {
		panic(err)
	}
	for _, val := range lst {
		if val.IsDir() {
			fileList = append(fileList, FileElement{val.Name(), false})
		} else {
			fileList = append(fileList, FileElement{val.Name(), true})
		}
	}
	return fileList
}

func createDir(path string) error {
	err := os.Mkdir(path, 0777)
	if err != nil {
		return err
	}
	return nil
}

func saveFile(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	w.Header().Set("Content-Disposition", "attachment; filename=" + params["path"])
	w.Header().Set("Content-Type", "application/octet-stream")
	http.ServeFile(w, r, params["path"])
}

func main() {
	currentPath, _ := filepath.Abs(".")
	myError := ""

	r := mux.NewRouter()
	r.HandleFunc("/", func (w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("templates/index.html")
		files := getFilesAndDirs(currentPath)
		tmpl.Execute(w, ViewData{currentPath, myError, files})
		myError = ""
	})

	r.HandleFunc("/create/dir", func (w http.ResponseWriter, r *http.Request) {
		createErr := createDir(filepath.Join(currentPath, "EmptyDir"))
		if createErr != nil {
			myError = createErr.Error()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.HandleFunc("/move/{path}", func (w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		if params["path"] == "to_parent" {
			currentPath = filepath.Dir(currentPath)
		} else {
			currentPath = filepath.Join(currentPath, params["path"])
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.HandleFunc("/delete/{path}", func (w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		deleteErr := os.Remove(filepath.Join(currentPath, params["path"]))
		if deleteErr != nil {
			myError = deleteErr.Error()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	r.HandleFunc("/download/{path}", saveFile)
	r.HandleFunc("/new", func (w http.ResponseWriter, r *http.Request) {
		tmpl, _ := template.ParseFiles("templates/create.html")
		tmpl.Execute(w, ViewData{"Create file", myError, nil})
	})

	r.HandleFunc("/upload", func (w http.ResponseWriter, r *http.Request) {
		file, header, getError := r.FormFile("myFile")
		if getError != nil {
			myError = getError.Error()
		}
		f, createError := os.OpenFile(header.Filename, os.O_WRONLY|os.O_CREATE, 0666)
		io.Copy(f, file)
		if createError != nil {
			myError = createError.Error()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
		defer f.Close()
		defer file.Close()
	})

	r.HandleFunc("/rename/{path}", func (w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		tmpl, _ := template.ParseFiles("templates/rename.html")
		var fileList []FileElement
		fileList = append(fileList, FileElement{params["path"], false})
		tmpl.Execute(w, ViewData{"Rename file", myError, fileList})
	})

	r.HandleFunc("/make_rename", func (w http.ResponseWriter, r *http.Request) {
		renameErr := os.Rename(r.FormValue("old_name"), r.FormValue("new_name"))
		if renameErr != nil {
			myError = renameErr.Error()
		}
		http.Redirect(w, r, "/", http.StatusSeeOther)
	})

	fmt.Println("Server is listening...")
	log.Fatal(http.ListenAndServe(":8000", r))
}