package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/jeromedoucet/dahu-git/types"
	"github.com/jeromedoucet/dahu-tests/container"
	"github.com/jeromedoucet/dahu-tests/ssh"
)

const DockerApiVersion = "1.37"

var tmpDir string

func TestMain(m *testing.M) {
	resetTmpDir()
	gogsId := container.StartGogs(DockerApiVersion)
	retCode := m.Run()
	container.StopContainer(gogsId, DockerApiVersion)
	os.RemoveAll(tmpDir)
	os.Exit(retCode)
}

func resetTmpDir() {
	if tmpDir != "" {
		os.RemoveAll(tmpDir)
	}
	var err error
	tmpDir, err = ioutil.TempDir("", "dahu-git-test")
	if err != nil {
		log.Fatal(err)
	}
}

func TestCloneProtectedKey(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	sshAuth := types.SshAuth{Url: "ssh://git@localhost:10022/tester/test-repo.git", Key: ssh.PrivateProtected, KeyPassword: "tester"}
	request := types.CloneRequest{SshAuth: sshAuth, NoCheckout: true, UseSsh: true, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expect 200 return code when testing a private git repository with protected ssh private key "+
			"Got %d", resp.StatusCode)
	}
	_, err = os.Stat(filepath.Join(tmpDir, ".git"))
	if err != nil {
		t.Error("Expect the repo to have been cloned in the dedicated directory, but there is not .git directory in it !")
	}
}

func TestCloneUnProtectedKey(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	sshAuth := types.SshAuth{Url: "ssh://git@localhost:10022/tester/test-repo.git", Key: ssh.PrivateUnprotected}
	request := types.CloneRequest{SshAuth: sshAuth, NoCheckout: true, UseSsh: true, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expect 200 return code when testing a private git repository with unprotected ssh private key "+
			"Got %d", resp.StatusCode)
	}
	_, err = os.Stat(filepath.Join(tmpDir, ".git"))
	if err != nil {
		t.Error("Expect the repo to have been cloned in the dedicated directory, but there is not .git directory in it !")
	}
}

func TestCloneHttpPrivate(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	httpAuth := types.HttpAuth{Url: "http://localhost:10080/tester/test-repo.git", User: "tester", Password: "test"}
	request := types.CloneRequest{HttpAuth: httpAuth, NoCheckout: true, UseHttp: true, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expect 200 return code when testing a private git repository with http "+
			"Got %d", resp.StatusCode)
	}
	_, err = os.Stat(filepath.Join(tmpDir, ".git"))
	if err != nil {
		t.Error("Expect the repo to have been cloned in the dedicated directory, but there is not .git directory in it !")
	}
}

func TestCloneHttpPublic(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	httpAuth := types.HttpAuth{Url: "http://localhost:10080/tester/test-repo-pub.git"}
	request := types.CloneRequest{HttpAuth: httpAuth, NoCheckout: true, UseHttp: true, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expect 200 return code when testing a public git repository with http "+
			"Got %d", resp.StatusCode)
	}
	_, err = os.Stat(filepath.Join(tmpDir, ".git"))
	if err != nil {
		t.Error("Expect the repo to have been cloned in the dedicated directory, but there is not .git directory in it !")
	}
}

func TestCloneHttpPrivateBadCredentials(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	httpAuth := types.HttpAuth{Url: "http://localhost:10080/tester/test-repo.git", User: "tester", Password: "toto"}
	request := types.CloneRequest{HttpAuth: httpAuth, NoCheckout: true, UseHttp: true, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 403 {
		t.Errorf("Expect 403 return code when testing a private git repository with http and bad credentials "+
			"Got %d", resp.StatusCode)
	}
}

func TestCloneUnknownKey(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	sshAuth := types.SshAuth{Url: "ssh://git@localhost:10022/tester/test-repo.git", Key: ssh.PrivateBad}
	request := types.CloneRequest{SshAuth: sshAuth, NoCheckout: true, UseSsh: true, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 403 {
		t.Errorf("Expect 403 return code when testing a private git repository with unknown key "+
			"Got %d", resp.StatusCode)
	}
}

func TestCloneUnknownRepo(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	httpAuth := types.HttpAuth{Url: "http://localhost:10080/tester/wrong-repo-pub.git"}
	request := types.CloneRequest{HttpAuth: httpAuth, NoCheckout: true, UseHttp: true, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expect 404 return code when testing an unknown git repository "+
			"Got %d", resp.StatusCode)
	}
}

func TestCloneUnknownBranch(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	httpAuth := types.HttpAuth{Url: "http://localhost:10080/tester/test-repo-pub.git"}
	request := types.CloneRequest{HttpAuth: httpAuth, NoCheckout: false, UseHttp: true, Branch: "feature/unknown"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 404 {
		t.Errorf("Expect 404 return code when testing an unknown git branch "+
			"Got %d", resp.StatusCode)
	}
}

func TestCloneNoAuthScheme(t *testing.T) {
	// given
	// server setup
	defer resetTmpDir()
	handler := new(cloneHandler)
	handler.directory = tmpDir
	s := httptest.NewServer(handler)
	defer s.Close()

	// request setup
	request := types.CloneRequest{NoCheckout: false, Branch: "master"}
	body, _ := json.Marshal(request)
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s", s.URL), bytes.NewBuffer(body))
	cli := &http.Client{}

	// when
	resp, err := cli.Do(req)

	// then
	if err != nil {
		t.Errorf("Expect to have no error, but got %s", err.Error())
	}
	if resp.StatusCode != 400 {
		t.Errorf("Expect 400 return code when testing a git clone without auth scheme "+
			"Got %d", resp.StatusCode)
	}
}
