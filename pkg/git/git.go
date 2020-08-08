package git

import (
	"bytes"
	"io/ioutil"
	"os/exec"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

var tagRegexp = regexp.MustCompile("(v|)([0-9]{1,3}.[0-9]{1,3}.[0-9]{1,3})")

const separator = "@@__GIT_SWIPE__@@"

type Tag struct {
	Name    string
	Hash    string
	Subject string
	Date    time.Time
}

type Log struct {
	Commit  string
	Subject string
	Date    time.Time
	Tag     string
}

type GIT struct {
}

type Option func(*execOptions)

type execOptions struct {
	env  []string
	args []string
}

func Env(env ...string) Option {
	return func(o *execOptions) {
		o.env = env
	}
}

func Args(args ...string) Option {
	return func(o *execOptions) {
		o.args = args
	}
}

func (g *GIT) exec(subCmd string, opts ...Option) (string, error) {
	o := &execOptions{}
	for _, opt := range opts {
		opt(o)
	}
	arr := append([]string{subCmd}, o.args...)
	var out bytes.Buffer
	cmd := exec.Command("git", arr...)
	cmd.Stdout = &out
	cmd.Stderr = ioutil.Discard
	cmd.Env = o.env

	err := cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
		if waitStatus, ok := exitError.Sys().(syscall.WaitStatus); ok {
			if waitStatus.ExitStatus() != 0 {
				return "", err
			}
		}
	}

	return strings.TrimRight(strings.TrimSpace(out.String()), "\000"), nil
}

func (g *GIT) GetLogs() ([]Log, error) {
	out, err := g.exec("log",
		Args(
			`--date=iso-strict`,
			"--pretty=format:%H"+separator+"%ad"+separator+"%s"+separator+"%d",
		),
	)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(out, "\n")
	var logs []Log
	for _, line := range lines {
		tokens := strings.Split(line, separator)
		if len(tokens) != 4 {
			continue
		}
		commit := strings.TrimSpace(tokens[0])
		subject := strings.TrimSpace(tokens[2])
		date, err := parseDate(tokens[1])
		if err != nil {
			return nil, err
		}

		logs = append(logs, Log{
			Commit:  commit,
			Subject: subject,
			Date:    date,
			Tag:     strings.TrimSpace(tagRegexp.FindString(tokens[3])),
		})
	}
	sort.Slice(logs, func(i, j int) bool {
		return !logs[i].Date.Before(logs[j].Date)
	})
	return logs, nil
}

func (g *GIT) DeleteTag(name string) error {
	_, err := g.exec("tag", Args("-d", name))
	return err
}

func (g *GIT) AssignTag(name, commit, subject string, date time.Time) error {
	opts := []Option{
		Args("-m", subject, name, commit),
	}
	if !date.IsZero() {
		opts = append(opts, Env("GIT_COMMITTER_DATE="+date.Format(time.RFC3339)))
	}
	_, err := g.exec("tag", opts...)
	return err
}

func (g *GIT) GetTags() ([]Tag, error) {
	out, err := g.exec("for-each-ref",
		Args(
			"--format",
			"%(refname)"+separator+"%(subject)"+separator+"%(taggerdate:iso-strict)"+separator+"%(*authordate:iso-strict)"+separator+"%(objectname)",
			"refs/tags",
		),
	)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(out, "\n")
	var tags []Tag
	for _, line := range lines {
		tokens := strings.Split(line, separator)
		if len(tokens) != 5 {
			continue
		}
		name := strings.Replace(tokens[0], "refs/tags/", "", 1)
		subject := strings.TrimSpace(tokens[1])
		hash := strings.TrimSpace(tokens[4])
		date, err := parseDate(tokens[3])
		if err != nil {
			return nil, err
		}
		tags = append(tags, Tag{
			Name:    name,
			Hash:    hash,
			Subject: subject,
			Date:    date,
		})
	}
	sort.Slice(tags, func(i, j int) bool {
		return !tags[i].Date.Before(tags[j].Date)
	})
	return tags, nil
}

func parseDate(dateStr string) (time.Time, error) {
	date, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return time.Time{}, err
	}
	return date, nil
}

func NewGIT() *GIT {
	return &GIT{}
}
