package provider

import (
	"github.com/google/go-github/v53/github"
)

type Config struct {
	Owner string
	Client *github.Client
}