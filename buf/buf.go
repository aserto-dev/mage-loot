package buf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aserto-dev/clui"
	"github.com/aserto-dev/mage-loot/deps"
	"github.com/aserto-dev/mage-loot/testutil"
	"github.com/pkg/errors"
)

type bufArgs struct {
	withLogin bool
	args      []string
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
	return RunWithEnv(nil, args...)
}

// Run runs the protoc CLI with the given environment variables.
func RunWithEnv(env map[string]string, args ...Arg) (err error) {
	bufArgs := &bufArgs{}

	for _, arg := range args {
		arg(bufArgs)
	}

	finalArgs := []string{}

	finalArgs = append(finalArgs, bufArgs.args...)

	ui.Normal().
		Msg(">>> executing buf " + strings.Join(bufArgs.args, " "))

	if bufArgs.withLogin {
		netrcFile, err := getLoginFile()
		if err != nil {
			return err
		}

		if env == nil {
			env = make(map[string]string)
		}

		env["NETRC"] = netrcFile
		defer func() {
			if os.Getenv("NETRC") != "" {
				return
			}
			err = os.Remove(netrcFile)
		}()
	}

	return deps.GoDepWithEnv(env, "buf")(finalArgs...)
}

func getLoginFile() (string, error) {
	if os.Getenv("NETRC") != "" {
		return os.Getenv("NETRC"), nil
	}

	fmt.Println("Using vault to get buf credentials")

	file, err := ioutil.TempFile(filepath.Join(deps.ExtTmpDir()), ".netrc*")
	if err != nil {
		return "", err
	}

	err = file.Chmod(0700)
	if err != nil {
		return "", err
	}

	bufToken := testutil.VaultValue("buf.build", "ASERTO_BUF_TOKEN")
	_, err = file.WriteString(fmt.Sprintf("machine buf.build\npassword %s", bufToken))
	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

// AddArg adds a simple argument.
func AddArg(arg string) func(*bufArgs) {
	return func(o *bufArgs) {
		o.args = append(o.args, arg)
	}
}

func WithLogin() func(*bufArgs) {
	return func(o *bufArgs) {
		o.withLogin = true
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
func GetTags(repository string) ([]Tag, error) {
	bufDep := deps.GoDepOutput("buf")
	out, err := bufDep("beta", "registry", "tag", "list", repository, "--format", "json", "--reverse")
	if err != nil {
		ui.Problem().Msg(fmt.Sprintf("Error retrieving tags for %s. Message: %s, Error: %s", repository, out, err.Error()))
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

func getBufPath(protoPluginPaths []string) (string, error) {
	path := os.Getenv("PATH")
	pathSeparator := string(os.PathListSeparator)

	for _, p := range protoPluginPaths {
		pluginFolderPath := p
		fileInfo, err := os.Stat(pluginFolderPath)
		if err != nil {
			return "", err
		}

		if !fileInfo.IsDir() {
			pluginFolderPath = filepath.Dir(pluginFolderPath)
		}

		path = path + pathSeparator + pluginFolderPath
	}
	return path, nil
}
