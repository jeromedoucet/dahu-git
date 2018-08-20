package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/jeromedoucet/dahu-git/client"
	"github.com/jeromedoucet/dahu-git/types"
)

type cloneHandler struct {
	directory string // the directory where the code must be cloned
}

func (h cloneHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var cloneReq types.CloneRequest
	var err client.GitError
	d := json.NewDecoder(r.Body)
	d.Decode(&cloneReq)

	cloneContext := types.CloneContext{
		Directory:  h.directory,
		NoCheckout: cloneReq.NoCheckout,
		Branch:     cloneReq.Branch,
		Progress:   os.Stdout,
	}

	if cloneReq.UseSsh {
		err = client.CloneWithSsh(cloneContext, cloneReq.SshAuth)
	} else if cloneReq.UseHttp {
		err = client.CloneWithHttp(cloneContext, cloneReq.HttpAuth)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if err == nil {
		w.WriteHeader(http.StatusOK)
	} else {
		if err.ErrorType() == client.RepositoryNotFound {
			w.WriteHeader(http.StatusNotFound)
		} else if err.ErrorType() == client.BadCredentials {
			w.WriteHeader(http.StatusForbidden)
		} else if err.ErrorType() == client.SshKeyReadingError {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusBadRequest)
		}
	}
}

func main() {

	port := flag.String("port", "80", "the port dahu-git will listen on")
	directory := flag.String("directory", "/data", "the place where dahu-git will clone the repository")
	flag.Parse()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	handler := new(cloneHandler)
	handler.directory = *directory

	http.Handle("/", handler)

	go func() {
		log.Printf("INFO >> Listening on http://0.0.0.0:%s", *port)
		if err := http.ListenAndServe(fmt.Sprintf(":%s", *port), nil); err != nil {
			log.Fatal(err)
		}
	}()

	// wait for kill signal
	<-stop
}
