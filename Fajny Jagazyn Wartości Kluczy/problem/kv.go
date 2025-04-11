package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

func checkPath(path string) error {
	if strings.Contains(path, ".") {
		return fmt.Errorf("🛑 nielegalne (hacking)")
	}

	if strings.Contains(path, "flag") {
		return fmt.Errorf("🛑 nielegalne (just to be sure)")
	}

	return nil
}

func main() {
	time.AfterFunc(180*time.Second, func() {
		os.Exit(0)
	})

	session, ok := os.LookupEnv("SESSION")
	if !ok {
		panic("SESSION env not set")
	}

	dataDir := "/tmp/kv." + session
	err := os.Mkdir(dataDir, 0o777)
	if err != nil {
		panic(err)
	}
	err = os.Chdir(dataDir)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/get", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if err = checkPath(name); err != nil {
			http.Error(w, "checkPath :(", http.StatusInternalServerError)
			return
		}

		file, err := os.Open(name)
		if err != nil {
			http.Error(w, "Open :(", http.StatusInternalServerError)
			return
		}

		data, err := io.ReadAll(io.LimitReader(file, 1024))
		if err != nil {
			http.Error(w, "ReadAll :(", http.StatusInternalServerError)
			return
		}

		w.Write(data)
	})

	http.HandleFunc("/set", func(w http.ResponseWriter, r *http.Request) {
		name := r.URL.Query().Get("name")
		if err = checkPath(name); err != nil {
			http.Error(w, "checkPath :(", http.StatusInternalServerError)
			return
		}

		err := os.WriteFile(name, []byte(r.URL.Query().Get("value"))[:1024], 0o777)
		if err != nil {
			http.Error(w, "WriteFile :(", http.StatusInternalServerError)
			return
		}
	})

	unixListener, err := net.Listen("unix", dataDir+"/kv.socket")
	if err != nil {
		panic(err)
	}
	http.Serve(unixListener, nil)
}
