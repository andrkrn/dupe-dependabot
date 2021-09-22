package main

import (
	"os"

	"github.com/andrkrn/dupe-dependabot/internal/service"
)

func main() {
	s := service.NewService(os.Getenv("GITHUB_TOKEN"))
	s.Run()
}
