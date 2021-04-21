package api

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

var fileLifetime time.Duration = 120000000000
var fileIdLetters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateFileID() string {
	output := make([]byte, 7)
	for i := range output {
		output[i] = fileIdLetters[rand.Intn(len(fileIdLetters))]
	}
	return string(output)
}

type FaxFile struct {
	id        string
	name      string
	createdAt time.Time

	uploadedBy string
	downloads  int
}

type Server struct {
	*mux.Router

	files map[string]*FaxFile
}

func NewServer() *Server {
	s := &Server{
		Router: mux.NewRouter(),
		files:  make(map[string]*FaxFile),
	}
	s.HandleFunc("/upload", s.upload()).Methods("POST", "OPTIONS")
	s.HandleFunc("/file/{id}", s.download()).Methods("GET", "OPTIONS")
	s.Use(mux.CORSMethodMiddleware(s.Router))

	os.RemoveAll("files")
	os.MkdirAll("files", 0755)

	return s
}

func (s *Server) upload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			return
		}
		println("Receiving upload request from", r.RemoteAddr)

		multipartReader, err := r.MultipartReader()
		if err != nil {
			println("Error while parsing request", err.Error())
			w.WriteHeader(400)
			w.Write(nil)
			return
		}
		added_files := make(map[string]*FaxFile)

		for {
			part, err := multipartReader.NextPart()
			if err != nil || part == nil {
				if err != nil && err.Error() != "EOF" {
					println("And error occurred: ", err)
				}
				break
			}
			file_id := generateFileID()
			println("Receiving file", file_id, "from", r.RemoteAddr)

			file, err := os.Create("files/" + file_id)
			if err != nil {
				panic(err)
			}

			for {
				bytes := make([]byte, 1000)
				n, err := part.Read(bytes)
				if n != 0 {
					file.Write(bytes[:n])
				}
				if err == nil {
					continue
				}
				if err.Error() == "EOF" {
					break
				} else {
					println("Could not read file: " + err.Error())
					return
				}
			}

			file.Close()

			fax_file := &FaxFile{
				id:        file_id,
				name:      part.FileName(),
				createdAt: time.Now(),

				downloads:  0,
				uploadedBy: r.RemoteAddr,
			}
			s.files[fax_file.id] = fax_file

			time.AfterFunc(fileLifetime, func() {
				delete(s.files, file_id)
			})
			time.AfterFunc(fileLifetime+60000000000, func() {
				os.RemoveAll("files/" + file_id)
			})

			added_files[part.FormName()] = fax_file
		}

		w.Header().Set("Content-Type", "application/json")

		if len(added_files) == 0 {
			w.WriteHeader(400)
			w.Write([]byte("{\"type\":\"error\",\"message\":\"Invalid request\"}"))
		}

		file_name_map := make(map[string]string, len(added_files))
		for key, file := range added_files {
			file_name_map[key] = file.id
		}
		file_name_map_string, _ := json.Marshal(file_name_map)

		w.Write([]byte("{\"type\":\"success\", \"files\":" + string(file_name_map_string) + "}"))
	}
}

func (s *Server) download() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			return
		}

		vars := mux.Vars(r)
		fax_file, ok := s.files[vars["id"]]
		if !ok {
			w.WriteHeader(404)
			w.Write(nil)
			return
		}
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+fax_file.name+"\"")

		file, err := os.Open("files/" + fax_file.id)
		if err != nil {
			println("Could not open a file: " + err.Error())
			return
		}

		fax_file.downloads += 1

		stats, _ := file.Stat()
		w.Header().Set("Content-Length", fmt.Sprint(stats.Size()))
		w.WriteHeader(200)

		for {
			bytes := make([]byte, 1000)
			n, err := file.Read(bytes)
			if n != 0 {
				w.Write(bytes[:n])
			}
			if err == nil {
				continue
			}
			if err.Error() == "EOF" {
				break
			} else {
				println("Could not read file: " + err.Error())
				return
			}
		}
	}
}
