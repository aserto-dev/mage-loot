package deps

import (
	"bytes"
	"runtime"
	"text/template"

	"github.com/pkg/errors"
)

type deps struct {
	Version string
	Arch    string
	OS      string
}

func parseStringTemplate(tpl, version string) string {

	d := deps{
		Version: version,
		Arch:    runtime.GOARCH,
		OS:      runtime.GOOS,
	}
	t := template.Must(template.New("tml").Parse(tpl))

	var buf bytes.Buffer
	err := t.Execute(&buf, d)
	if err != nil {
		panic(errors.Wrap(err, "failed to render template with version"))
	}

	return buf.String()
}

func parseArrayTemplate(tpls []string, version string) []string {

	var out []string
	for _, tpl := range tpls {
		out = append(out, parseStringTemplate(tpl, version))
	}
	return out
}
