package client

import (
	"fmt"
	"log"
	"strings"

	"github.com/jeromedoucet/dahu-git/types"
	ssh2 "golang.org/x/crypto/ssh"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
)

type GitErrorType int

const (
	BadCredentials GitErrorType = 1 + iota
	RepositoryNotFound
	SshKeyReadingError
	OtherError
)

type GitError interface {
	Error() string
	ErrorType() GitErrorType
}

type simpleGitError struct {
	msg     string
	errType GitErrorType
}

func (err simpleGitError) Error() string {
	return err.msg
}

func (err simpleGitError) ErrorType() GitErrorType {
	return err.errType
}

func newGitError(msg string, errType GitErrorType) GitError {
	return simpleGitError{msg: msg, errType: errType}
}

func CloneWithSsh(ctx types.CloneContext, auth types.SshAuth) GitError {
	gitAuth, sshError := ssh.NewPublicKeys("git", []byte(auth.Key), auth.KeyPassword)
	if sshError != nil {
		return newGitError(sshError.Error(), SshKeyReadingError)
	}
	// see https://github.com/src-d/go-git/issues/454
	// for the moment, we don't want to add host verification
	// as a feature. If an host is not listed in the ~/.ssh/known_host
	// an error is returned by go-git lib.
	gitAuth.HostKeyCallback = ssh2.InsecureIgnoreHostKey()
	return doClone(auth.Url, ctx, gitAuth)
}

func CloneWithHttp(ctx types.CloneContext, auth types.HttpAuth) GitError {
	gitAuth := &http.BasicAuth{Username: auth.User, Password: auth.Password}
	return doClone(auth.Url, ctx, gitAuth)
}

func doClone(url string, ctx types.CloneContext, gitAuth transport.AuthMethod) GitError {
	log.Printf("Git >> start cloning repository %s on branch %s", url, ctx.Branch)
	_, err := git.PlainClone(ctx.Directory, false, &git.CloneOptions{
		URL:               url,
		NoCheckout:        ctx.NoCheckout,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Auth:              gitAuth,
		SingleBranch:      true,
		ReferenceName:     plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", ctx.Branch)), // simple 'naive' implementation. Should maybee improved later
		Progress:          ctx.Progress,
		Tags:              git.NoTags, // for the moment, we don't deals with tags at all
	})
	if err == nil {
		log.Print("Git >> Clone finished without error")
	} else {
		log.Printf("Git >> Clone finished with error : %s", err.Error())
	}
	return fromGitToGitError(err)
}

func fromGitToGitError(err error) GitError {
	if err == nil {
		return nil
	}
	errStr := err.Error()
	switch err {
	case transport.ErrRepositoryNotFound:
		return newGitError(errStr, RepositoryNotFound)
	case transport.ErrAuthenticationRequired:
		return newGitError(errStr, BadCredentials)
	default:
		if strings.Contains(errStr, "no supported methods remain") {
			// this error come directly from clientAuthenticate
			// in client_auth.go from 'golang.org/x/crypto/ssh' package.
			//
			// This error generaly means that the private key used for authentication
			// is not the right one. TODO: maybee use a more specific error ?
			//
			// Because the error is thrown through fmt.Errorf function
			// there is no possibility but checking the text of the error
			// to detect it !
			return newGitError(errStr, BadCredentials)
		} else if strings.Contains(strings.ToLower(errStr), "repository does not exist") {
			// try to catch ssh error related to inexistant repository.
			// This kind of error may be treat as 'unknow error' by the underlying
			// git library (go-git-v4 plumbing/transport/internal/common/common.go)
			//
			// Some Pr may improve this handling, but some case may be missing, that's why
			// we make a try to handle the missing cases here
			return newGitError(errStr, RepositoryNotFound)

		} else if strings.Contains(strings.ToLower(errStr), "couldn't find remote ref") {
			return newGitError(errStr, RepositoryNotFound)
		} else {
			return newGitError(errStr, OtherError)
		}
	}
}
