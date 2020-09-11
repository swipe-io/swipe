package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/swipe-io/swipe/v2/internal/git"
)

func main() {
	g := git.NewGIT()
	tags, err := g.GetTags()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	logs, err := g.GetLogs()
	if err != nil {
		log.Println(err.Error())
		os.Exit(1)
	}
	r := regexp.MustCompile("(v|)([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})")
	for _, l := range logs {
		if l.Tag != "" {
			continue
		}
		extractedTag := r.FindString(l.Subject)
		if extractedTag == "" {
			continue
		}
		for _, t := range tags {
			tagName := strings.TrimPrefix(t.Name, "v")
			if extractedTag == tagName {
				if err := g.DeleteTag(t.Name); err != nil {
					log.Println(err.Error())
					os.Exit(1)
				}
				if err := g.AssignTag(t.Name, l.Commit, t.Subject, t.Date); err != nil {
					log.Println(err.Error())
					os.Exit(1)
				}
				fmt.Println(tagName, l.Commit)
			}
		}
	}
}
