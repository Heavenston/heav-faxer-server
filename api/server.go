package api

import (
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"cloud.google.com/go/storage"
	"github.com/gorilla/mux"
)

var fileIdLetters = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func generateFileID() string {
	output := make([]byte, 7)
	for i := range output {
		output[i] = fileIdLetters[rand.Intn(len(fileIdLetters))]
	}
	return string(output)
}

type Server struct {
	*mux.Router

	googleAccessId   string
	googlePrivateKey []byte
	googleBucket     string
}

func NewServer(googleAccessId string, googlePrivateKey string, googleBucket string) (*Server, error) {
	googlePrivateKeyBytes, err := ioutil.ReadFile(googlePrivateKey)
	if err != nil {
		println("Could not read google private key: ", err.Error())
		return nil, err
	}

	s := &Server{
		Router: mux.NewRouter(),

		googleAccessId:   googleAccessId,
		googlePrivateKey: googlePrivateKeyBytes,
		googleBucket:     googleBucket,
	}
	s.HandleFunc("/getUploadUrl", s.getUploadURL()).Methods("POST", "OPTIONS")
	s.Use(mux.CORSMethodMiddleware(s.Router))

	return s, nil
}

func (s *Server) getUploadURL() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == "OPTIONS" {
			return
		}

		filename := r.URL.Query().Get("filename")
		if len(filename) < 2 || len(filename) > 30 {
			w.WriteHeader(400)
			w.Write(nil)
			return
		}

		fileId := generateFileID()

		signedUrl, err := storage.SignedURL(s.googleBucket, fileId, &storage.SignedURLOptions{
			GoogleAccessID: s.googleAccessId,
			PrivateKey:     s.googlePrivateKey,
			Expires:        time.Now().Add(time.Minute * 15),
			Method:         "PUT",
			Headers: []string{
				"content-disposition:attachment; filename=\"" + filename + "\"",
			},
			ContentType: "application/octet-stream",
			Scheme:      storage.SigningSchemeV4,
		})
		if err != nil {
			println("Could not create a post policy:", err.Error())
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		w.Write([]byte("{\"type\": \"success\", \"url\": \"" + signedUrl + "\", \"file_id\": \"" + fileId + "\"}"))
	}
}
