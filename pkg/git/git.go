package git

import (
	"sort"
	"strings"
	"time"

	"github.com/tsuyoshiwada/go-gitcmd"
)

const separator = "@@__GIT_SWIPE__@@"

type Tag struct {
	Name    string
	Subject string
	Date    time.Time
}

type GIT struct {
}

func (g *GIT) GetTags() ([]Tag, error) {
	git := gitcmd.New(nil)
	out, err := git.Exec("for-each-ref",
		"--format",
		"%(refname)"+separator+"%(subject)"+separator+"%(taggerdate)"+separator+"%(authordate)",
		"refs/tags",
	)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	var tags []Tag
	for _, line := range lines {
		tokens := strings.Split(line, separator)
		if len(tokens) != 4 {
			continue
		}
		name := strings.Replace(tokens[0], "refs/tags/", "", 1)
		subject := strings.TrimSpace(tokens[1])
		date, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", tokens[2])
		if err != nil {
			date, err = time.Parse("Mon Jan 2 15:04:05 2006 -0700", tokens[2])
			if err != nil {
				return nil, err
			}
		}
		tags = append(tags, Tag{
			Name:    name,
			Subject: subject,
			Date:    date,
		})
	}
	sort.Slice(tags, func(i, j int) bool {
		return !tags[i].Date.Before(tags[j].Date)
	})
	return tags, nil
}

func NewGIT() *GIT {
	return &GIT{}
}
