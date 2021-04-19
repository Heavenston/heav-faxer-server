package api

import (
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

type FaxFile struct {
	id        string
	name      string
	createdAt time.Time
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
	s.HandleFunc("/upload", s.upload()).Methods("POST")

	os.RemoveAll("files")
	os.MkdirAll("files", 0755)

	return s
}

func (s *Server) upload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		multipartReader, err := r.MultipartReader()
		if err != nil {
			r.Response.StatusCode = 404
			r.Write(nil)
			return
		}
		added_files := make([]*FaxFile, 0)

		for {
			part, err := multipartReader.NextPart()
			if err != nil || part == nil {
				if err != nil && err.Error() != "EOF" {
					println("And error occurred: ", err)
				}
				break
			}
			println("FILE!")
			println(part.FileName())

			file_id := "hello"
			file, err := os.Create("files/" + file_id)
			if err != nil {
				panic(err)
			}

			for {
				bytes := make([]byte, 500)
				n, err := part.Read(bytes)
				if n == 0 || err != nil {
					break
				}
				file.Write(bytes)
			}

			fax_file := &FaxFile{
				id:        file_id,
				name:      part.FileName(),
				createdAt: time.Now(),
			}
			s.files[fax_file.id] = fax_file

			added_files = append(added_files, fax_file)
		}

		w.Header().Add("Content-Type", "application/json")

		if len(added_files) == 0 {
			w.WriteHeader(400)
			w.Write([]byte("{\"type\":\"error\",\"message\":\"Invalid request\"}"))
		}

		w.Write([]byte("Hello"))
	}
}
