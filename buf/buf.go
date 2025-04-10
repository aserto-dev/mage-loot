package buf

import (
	"encoding/json"
	"fmt"
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

type labelResult struct {
	NextPage string  `json:"next_page"`
	Labels   []Label `json:"labels"`
}

type Label struct {
	Name        string `json:"name"`
	Commit      string `json:"commit"`
	CreatedTime string `json:"create_time"`
}

// Arg represents a protoc CLI argument.
type Arg func(*bufArgs)

var (
	ui        = clui.NewUI()
	ErrNoTags = errors.New("no tags found in repository")
)

// Run runs the protoc CLI.
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

	err := os.MkdirAll(deps.ExtTmpDir(), 0o700)
	if err != nil {
		return "", errors.Wrapf(err, "failed to create tmp dir")
	}

	file, err := os.CreateTemp(deps.ExtTmpDir(), ".netrc*")
	if err != nil {
		return "", err
	}

	err = file.Chmod(0o700)
	if err != nil {
		return "", err
	}

	bufToken := testutil.VaultValue("buf.build", "ASERTO_BUF_TOKEN")
	_, err = fmt.Fprintf(file, "machine buf.build\npassword %s", bufToken)
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

// Gets the 10 most recent labels sorted from newest to oldest.
func GetTags(repository string) ([]Label, error) {
	bufDep := deps.GoDepOutput("buf")
	out, err := bufDep("registry", "module", "label", "list", repository, "--format", "json")
	if err != nil {
		ui.Problem().Msg(fmt.Sprintf("Error retrieving tags for %s. Message: %s, Error: %s", repository, out, err.Error()))
		return nil, err
	}
	result := labelResult{}
	err = json.Unmarshal([]byte(out), &result)
	if err != nil {
		return nil, errors.Wrapf(err, "invalid get tags response, probably no tags pushed")
	}

	return result.Labels, nil
}

// Gets the latest commit based on create_date.
func GetLatestTag(repository string) (Label, error) {
	tags, err := GetTags(repository)
	if err != nil {
		return Label{}, err
	}

	if len(tags) == 0 {
		return Label{}, errors.Wrap(ErrNoTags, repository)
	}

	latestTag := tags[0]
	for _, tag := range tags {

		latestTime, err := time.Parse(time.RFC3339, latestTag.CreatedTime)
		if err != nil {
			return Label{}, err
		}
		tagTime, err := time.Parse(time.RFC3339, tag.CreatedTime)
		if err != nil {
			return Label{}, err
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
