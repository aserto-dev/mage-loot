# mage-loot

[![Built with Mage](https://magefile.org/badge.svg)](https://magefile.org)
![ci](https://github.com/aserto-dev/mage-loot/workflows/ci/badge.svg?branch=main)
[![Coverage Status](https://coveralls.io/repos/github/aserto-dev/mage-loot/badge.svg?branch=main&t=4v6ABX&service=github)](https://coveralls.io/github/aserto-dev/mage-loot?branch=main)

Collection of reusable, useful code for magefiles.

> **FAQ:** The `Depfile` used by `mage-loot` is what you probably want as a starting point for your project.

A bit about mage-loot, and what it contains:

- Dependency management using `Depfile`.
- Helpers for automating the `buf`, `dotnet`, and `protoc` CLIs.
- Opinionated build functions for `go` and `docker`.
- Opinionated helper functions for go projects, like linting, testing and code generation.

## Depfile

One problem that we’ve had to repeatedly solve is making sure the human and robot teams all use the same tools, binaries and libraries when building projects.
There are some solutions out there, but they didn’t fit our needs for various reasons (some are too hard to adopt, some don’t have enough features).

With Depfile, we’re aiming for enough features so that you can reliably build complex projects, but simple enough that you understand how to use it in less than 10 min.

As a developer, you have to buy into [mage](https://github.com/magefile/mage).
Next to your `magefile`, you add a `Depfile`. It’s a yaml file with the following structure:

```yaml
---
go:
  tool:
    importPath: "so.me/import/path"
    version: "v1.0.0"

bin:
  useful-binary:
    url: 'https://so.me/url/v{{.Version}}/protoc-{{.Version}}-{{if eq .OS "darwin"}}osx{{else}}{{.OS}}{{end}}-x86_64.zip'
    version: "1.1.0"
    sha:
      linux-amd64: "d4246a5136cf9cd1abc851c521a1ad6b8884df4feded8b9cbd5e2a2226d4b357"
      darwin-amd64: "68901eb7ef5b55d7f2df3241ab0b8d97ee5192d3902c59e7adf461adc058e9f1"
    zipPaths:
    - "bin/the.binary"
lib:
  useful-lib:
    url: "https://so.me/url/v{{.Version}}.tgz"
    version: "2.5.0"
    sha: "e8334c270a479f55ad9f264e798680ac536f473d7711593f6eadab3df2d1ddc3"
    libPrefix: "lib-{{.Version}}"
    tgzPaths:
    - "*/glob/patterns/*.proto"
```

You’ll notice we support 3 types of dependencies:
- go
- binaries
- libraries

For **go** tools (where you need a go tool but it’s not a dependency of your app), we just use `go` to install the version you specify.

For **binaries**, we download the file or archive from a location we calculate based on a template and a version, OS and architecture.
If you’re downloading an archive, you can specify which file to extract from it. Binaries will live in a `.ext/bin` directory inside your project.
We also need you to give us the SHA of the artifact we’re downloading, so we can make sure there’s no trickery!

For **libraries**, we assume you’re downloading archives, either zip or tgz. You can again use a template for the download URL, but there’s no differentiation on architecture or OS. Libraries live in `.ext/lib`. You can use globbing patterns to select which files to unpack from the archive.
Again, we need a SHA to verify integrity.

You can use the Depfile from [mage-loot](https://github.com/aserto-dev/mage-loot/blob/main/Depfile) itself as an example to get you started.
