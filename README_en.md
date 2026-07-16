<!-- BEGIN_DOCS -->
<div align="center">

[🇧🇷 Versão em Português](README.md)

<a name="readme-top"></a>

Hello Human 👽! Welcome to my repository 👋

<img src="https://github.com/lpsm-dev/lpsm-dev/blob/5cf57b68283a857a105430d0d6c9290ee657a012/.github/assets/go-cli.png" width="350"/>

**Generate beautiful documentation for your Git repositories**

[![CI](https://github.com/lpsm-dev/gtoc/actions/workflows/ci.yaml/badge.svg)](https://github.com/lpsm-dev/gtoc/actions/workflows/ci.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/lpsm-dev/gtoc)](https://goreportcard.com/report/github.com/lpsm-dev/gtoc)
[![Commitizen friendly](https://img.shields.io/badge/commitizen-friendly-brightgreen.svg)](https://www.conventionalcommits.org/en/v1.0.0/)
[![Semantic Release](https://img.shields.io/badge/%20%20%F0%9F%93%A6%F0%9F%9A%80-semantic--release-e10079.svg)](https://semantic-release.gitbook.io/semantic-release/usage/configuration)
[![Built with Devbox](https://jetpack.io/img/devbox/shield_galaxy.svg)](https://jetpack.io/devbox/docs/contributor-quickstart/)

📌 Star this repository to follow updates and news ( ≖‿ ≖ )

</div>

> [!NOTE]
>
> **NOTICE**: This repository is constantly evolving. If you find a bug or have suggestions, please open an [issue](https://github.com/lpsm-dev/gtoc/issues/new/choose) or send a [pull request](https://github.com/lpsm-dev/gtoc/pulls).

<!-- START_TABLE_OF_CONTENTS -->

- [Overview](#overview)
  - [Goal](#goal)
  - [Context and Motivation](#context-and-motivation)
- [Implementation](#implementation)
  - [Requirements](#requirements)
  - [Installation](#installation)
  - [Usage](#usage)
- [Contributing](#contributing)
- [Versioning](#versioning)
- [Troubleshooting](#troubleshooting)
- [Show your support](#show-your-support)

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- END_TABLE_OF_CONTENTS -->

# Overview

`gtoc` is a CLI written in Go that generates and keeps up to date the table of contents of Markdown files. It reads the file's headings, builds a hierarchical index with GitHub-compatible anchors and inserts the result between HTML markers, idempotently: running the command twice produces the same output.

## Goal

Eliminate manual maintenance of tables of contents in long READMEs and docs. `gtoc` takes care of:

- Generating the index from the file's actual headings (`#` through `######`);
- Creating anchors exactly the way GitHub does, including accented characters (`Instalação` becomes `#instalação`) and duplicate headings (`-1`, `-2` suffixes);
- Skipping headings inside code blocks and inside the TOC itself;
- Updating the existing block in place, preserving the rest of the file and its permissions.

## Context and Motivation

Well-structured READMEs are easier to navigate, but hand-maintained tables of contents go stale every time a section is added or renamed. Existing tools (such as doctoc) solve part of the problem, but this project was born to: (1) ship as a single binary with no Node dependency; (2) handle accented anchors correctly for Portuguese content; and (3) serve as a playground for Go CLI development best practices.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Implementation

## Requirements

None to use the binaries published on the releases page. To build from source you need Go `1.25+` (see the exact version in [go.mod](go.mod)).

## Installation

Via `go install`:

```bash
go install github.com/lpsm-dev/gtoc@latest
```

Via binary: download the file for your platform from the [releases](https://github.com/lpsm-dev/gtoc/releases) page and put it on your `PATH`. Once installed, upgrade with the CLI itself:

```bash
gtoc upgrade
```

Via Docker (local build):

```bash
docker build -t gtoc .
docker run --rm -v "$(pwd)":/work gtoc generate README.md
```

## Usage

Generate or update a file's table of contents:

```bash
gtoc generate README.md            # updates the file in place
gtoc generate README.md --dry-run  # only prints what would be generated
gtoc generate README.md --depth 3  # limits heading depth
gtoc generate README.md --exclude "draft,private"
```

The TOC is inserted (and later updated) between the markers below. On the first run without markers it is added at the top of the file:

```markdown
<!-- START_TABLE_OF_CONTENTS -->
<!-- END_TABLE_OF_CONTENTS -->
```

Apply README formatting best practices (`BEGIN_DOCS`/`END_DOCS` markers, `readme-top` anchor and "back to top" links at the end of every `#` section):

```bash
gtoc analyze --file README.md
```

`generate` flags:

| Flag | Default | Description |
| --------- | ------- | ------------------------------------------------------------------ |
| `--file` | - | Path to the Markdown file (or pass it as a positional argument) |
| `--depth` | `0` | Maximum heading depth (`0` = unlimited) |
| `--exclude` | - | Comma-separated heading texts to exclude (case-insensitive substring match) |
| `--dry-run` | `false` | Print the result without writing to the file |
| `--pretty` | `false` | In dry-run, render the whole formatted file in the terminal |

Global flags: `--log-level` (`debug`, `info`, `warn`, `error`, `fatal`), `--log-format` (`text`, `json`) and `--log-no-colors`.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Contributing

Would you like to contribute? Great! There is a contribution guide to help you. Click [here](CONTRIBUTING.md) to read it.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Versioning

To check the change history, see the [**CHANGELOG.md**](CHANGELOG.md) file.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Troubleshooting

If you run into any problem, open an [issue](https://github.com/lpsm-dev/gtoc/issues/new/choose) in this project.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

# Show your support

<div align="center">

Give this project a ⭐️ if it helped you!

<img src="https://github.com/lpsm-dev/lpsm-dev/blob/0062b174ec9877e6dfc78817f314b4a0690f63ff/.github/assets/yoda.gif" width="225"/>

<br>
<br>

Made with 💜 by the **DevOps Team** :wave: inspired by [readme-md-generator](https://github.com/kefranabg/readme-md-generator)

</div>

<p align="right">(<a href="#readme-top">back to top</a>)</p>

<!-- END_DOCS -->
