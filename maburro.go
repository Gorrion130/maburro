package main

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Persona struct {
	Nombre   string
	cookie   string
	Password string
}

type Datos struct {
	mutex sync.Mutex
	datos []Persona
}

func (datoss *Datos) writeDbToFile() {
	db, _ := os.OpenFile("holi.db", os.O_WRONLY, 0777)
	dbParsed, _ := json.Marshal(datoss.datos)
	db.Write(dbParsed)
	db.Close()
	//fmt.Println(string(dbParsed))
}

func (datoss *Datos) readDbFromFile() {
	dbFile, _ := os.OpenFile("holi.db", os.O_RDONLY, 0777)
	len, _ := io.Copy(io.Discard, dbFile)
	dbData := make([]byte, len)
	dbFile.ReadAt(dbData, 0)
	//fmt.Println(string(dbData))
	db := make([]Persona, 1)
	json.Unmarshal(dbData, &db)
	dbFile.Close()
	datoss.datos = append(datoss.datos, db...)
	//fmt.Println(db)
}

func (datoss *Datos) generarHtml(w http.ResponseWriter, r *http.Request, plantillaString string, param ...string) {
	html1, _ := os.OpenFile(plantillaString, os.O_RDONLY, 0777)
	i, _ := io.Copy(io.Discard, html1)
	plantilla := make([]byte, i)
	html1.ReadAt(plantilla, 0)
	html1.Close()
	plantillaParsed := string(plantilla)
	//fmt.Println(plantillaParsed)
	htmlParsed := strings.ReplaceAll(plantillaParsed, "$name", param[0])
	fmt.Fprintln(w, htmlParsed)
}

func (datoss *Datos) register(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		for _, x := range r.Cookies() {
			if x.Name == "maburro" {
				for _, y := range datoss.datos {
					if y.cookie == x.Value && y.cookie != "" {
						http.Redirect(w, r, "/my", 303)
					}
				}
			}
		}
		datoss.generarHtml(w, r, "plantilla_register.html", "")
	case "POST":
		hash := base64.StdEncoding.EncodeToString([]byte(time.Now().String()))
		usr := r.FormValue("test")
		pass := r.FormValue("pass")
		passParsed := fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
		userExists := false

		for _, x := range datoss.datos {
			if x.Nombre == usr {
				userExists = true
				break
			}
		}
		if !userExists {
			dato := Persona{
				Nombre:   usr,
				cookie:   string(hash),
				Password: passParsed,
			}
			cookie := &http.Cookie{
				Name:   "maburro",
				Value:  string(hash),
				MaxAge: 400,
			}
			http.SetCookie(w, cookie)
			datoss.datos = append(datoss.datos, dato)
			http.Redirect(w, r, "/my", 303)
		}
		datoss.writeDbToFile()
	}

}

func (datoss *Datos) img(w http.ResponseWriter, r *http.Request) {
	img1, _ := os.OpenFile("mine.jpg", os.O_RDONLY, 0777)
	io.Copy(w, img1)
}

func (datoss *Datos) login(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		for _, x := range r.Cookies() {
			if x.Name == "maburro" {
				for _, y := range datoss.datos {
					if y.cookie == x.Value && y.cookie != "" {
						http.Redirect(w, r, "/my", 303)
					}
				}
			}
		}
		datoss.generarHtml(w, r, "plantilla.html", "")
	case "POST":
		user := r.FormValue("test")
		pass := r.FormValue("pass")
		hash := base64.StdEncoding.EncodeToString([]byte(time.Now().String()))

		for z, x := range datoss.datos {

			cookie := &http.Cookie{
				Name:   "maburro",
				Value:  hash,
				MaxAge: 1400,
			}

			if x.Nombre == user && fmt.Sprintf("%x", sha256.Sum256([]byte(pass))) == x.Password {
				datoss.datos[z].cookie = hash
				http.SetCookie(w, cookie)
				http.Redirect(w, r, "/my", 303)
			}
		}
		datoss.generarHtml(w, r, "plantilla.html", "")
	}
}

func (datoss *Datos) logged(w http.ResponseWriter, r *http.Request) {
	logged := false
	switch r.Method {
	case "GET":
		for _, x := range r.Cookies() {
			if x.Name == "maburro" {
				for _, y := range datoss.datos {
					if y.cookie == x.Value && y.cookie != "" {
						datoss.generarHtml(w, r, "logged.html", y.Nombre)
						logged = true
					}
				}
			}
		}
		if !logged {
			http.Redirect(w, r, "/", 303)
		}
	case "POST":
		for _, x := range r.Cookies() {
			if x.Name == "maburro" {
				for z, y := range datoss.datos {
					if y.cookie == x.Value {
						datoss.datos[z].cookie = ""
						http.Redirect(w, r, "/", 303)
						logged = true
					}
				}
			}
		}
		if !logged {
			http.Redirect(w, r, "/", 303)
		}
	}
}

func main() {
	fmt.Println("loading...")
	datoss := Datos{datos: make([]Persona, 0)}
	datoss.readDbFromFile()

	http.HandleFunc("/register", datoss.register)
	http.HandleFunc("/", datoss.login)
	http.HandleFunc("/my", datoss.logged)
	http.HandleFunc("/img1", datoss.img)

	if runtime.GOOS == "windows" {
		exec.Command("explorer", "http://127.0.0.1:8080").Run()
	} else {
		exec.Command("xdg-open", "http://127.0.0.1:8080").Run()
	}

	http.ListenAndServe(":8080", nil)
}
