package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
        "os"
)

const Password = "4BU3L4_M4R14_LU154"

type StoredFile struct {
	Name string
	Data []byte
}

var (
	mu   sync.Mutex
	file *StoredFile
)

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	parts := strings.SplitN(string(body), "\n", 3)
	if len(parts) != 3 {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	password := strings.TrimSpace(parts[0])
	filename := strings.TrimSpace(parts[1])
	b64 := strings.TrimSpace(parts[2])

	if password != Password {
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		http.Error(w, "Invalid Base64", http.StatusBadRequest)
		return
	}

	mu.Lock()
	file = &StoredFile{
		Name: filename,
		Data: data,
	}
	mu.Unlock()

	fmt.Fprintln(w, "OK")
}

func download(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("password") != Password {
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if file == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	fmt.Fprintf(w, "%s\n", file.Name)
	fmt.Fprint(w, base64.StdEncoding.EncodeToString(file.Data))
}

func deleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Query().Get("password") != Password {
		http.Error(w, "Wrong password", http.StatusUnauthorized)
		return
	}

	mu.Lock()
	file = nil
	mu.Unlock()

	fmt.Fprintln(w, "Deleted")
}

func ping(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func main() {
	http.HandleFunc("/upload", upload)
	http.HandleFunc("/download", download)
	http.HandleFunc("/delete", deleteFile)
	http.HandleFunc("/ping", ping)

	fmt.Println("Listening on :10000")

	port := os.Getenv("PORT")
        if port == "" {
            port = "10000"
        }

        err := http.ListenAndServe(":"+port, nil)
        if err != nil {
            panic(err)
        }
}
