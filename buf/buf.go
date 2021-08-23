package buf

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/pkg/errors"
)

type bufArgs struct {
	args []string
}

type tagResult struct {
	Results []Tag `json:"results"`
}

type Tag struct {
	Name        string `json:"name"`
	Commit      string `json:"commit"`
	CreatedTime string `json:"create_time"`
}

// Arg represents a protoc CLI argument
type Arg func(*bufArgs)

var (
	ui = clui.NewUI()
)

// Run runs the protoc CLI
func Run(args ...Arg) error {
	bufArgs := &bufArgs{}

	for _, arg := range args {
		arg(bufArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, bufArgs.args...)

	ui.Normal().
		Msg(">>> executing buf " + strings.Join(bufArgs.args, " "))

	return deps.GoDep("buf")(finalArgs...)
}

// AddArg adds a simple argument.
func AddArg(arg string) func(*bufArgs) {
	return func(o *bufArgs) {
		o.args = append(o.args, arg)
	}
}

func AddPaths(paths []string) func(*bufArgs) {
	return func(o *bufArgs) {
		for _, p := range paths {
			o.args = append(o.args, "--path", p)
		}
	}
}

// Gets all the tags from a buf repository
func GetTags(repostory string) ([]Tag, error) {
	bufDep := deps.GoDepOutput("buf")
	out, err := bufDep("beta", "registry", "tag", "list", repostory, "--format", "json")
	if err != nil {
		ui.Problem().Msg(fmt.Sprintf("Error retrieving tags for %s. Message: %s, Error: %s", repostory, out, err.Error()))
		return nil, err
	}
	result := tagResult{}
	err = json.Unmarshal([]byte(out), &result)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid get tags response, probably no tags pushed")
	}

	return result.Results, nil
}

// Gets the latest commit based on create_date
func GetLatestTag(repository string) (Tag, error) {
	tags, err := GetTags(repository)
	if err != nil {
		return Tag{}, err
	}

	if len(tags) == 0 {
		return Tag{}, fmt.Errorf("no tags found for repository %s", repository)
	}

	latestTag := tags[0]
	for _, tag := range tags {

		latestTime, err := time.Parse(time.RFC3339, latestTag.CreatedTime)
		if err != nil {
			return Tag{}, err
		}
		tagTime, err := time.Parse(time.RFC3339, tag.CreatedTime)
		if err != nil {
			return Tag{}, err
		}
		if latestTime.Before(tagTime) {
			latestTag = tag
		}
	}
	return latestTag, nil
}
