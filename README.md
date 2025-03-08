# gtoc - Git Table of Contents Generator

<div align="center">

![gtoc logo](/placeholder.svg?height=150&width=150)

**Generate beautiful documentation indexes for your Git repositories**

[![Go Report Card](https://goreportcard.com/badge/github.com/lpsm-dev/gtoc)](https://goreportcard.com/report/github.com/lpsm-dev/gtoc)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

</div>

## 📖 Overview

`gtoc` is a command-line tool that automatically generates a hierarchical index of markdown files in your Git repository. It helps maintain organized documentation by creating and updating a table of contents in your README or other markdown files.

<!-- START_GTOC -->

# Documentation Index

## Features

- Automatically scans your Git repository for markdown files
- Generates a hierarchical index based on directory structure
- Extracts titles from H1 headers in each file
- Updates a specified markdown file with the generated index
- Respects `.gitignore` rules
- Customizable depth, patterns, and exclusions

<!-- END_GTOC -->

## 🚀 Installation

### Using Go

\`\`\`bash
go install github.com/lpsm-dev/gtoc@latest
\`\`\`

### From Source

\`\`\`bash
git clone https://github.com/lpsm-dev/gtoc.git
cd gtoc
go build
\`\`\`

## 🔧 Usage

The basic command structure is:

\`\`\`bash
gtoc generate [file] [flags]
\`\`\`

You can specify the file either as the first argument or using the `--file` flag.

### Examples

Generate an index for your README.md:

\`\`\`bash
gtoc generate README.md
\`\`\`

Or using the flag:

\`\`\`bash
gtoc generate --file README.md
\`\`\`

Generate an index for a specific documentation file with depth limit:

\`\`\`bash
gtoc generate docs/index.md --depth 2 --pattern "docs/\*_/_.md"
\`\`\`

Preview changes without writing to the file:

\`\`\`bash
gtoc generate README.md --dry-run
\`\`\`

Exclude specific paths:

\`\`\`bash
gtoc generate README.md --exclude "vendor/_,node_modules/_"
\`\`\`

### Available Flags

| Flag        | Description                                                                   | Default   |
| ----------- | ----------------------------------------------------------------------------- | --------- |
| `--file`    | Path to the markdown file to update (can also be specified as first argument) | -         |
| `--depth`   | Maximum directory depth (0 for unlimited)                                     | 0         |
| `--pattern` | Glob pattern to filter markdown files                                         | `**/*.md` |
| `--exclude` | Comma-separated list of paths to exclude                                      | -         |
| `--dry-run` | Preview changes without writing                                               | `false`   |

## 📋 How It Works

1. `gtoc` finds the Git repository root
2. It scans for markdown files matching your pattern
3. It extracts titles from H1 headers in each file
4. It generates a hierarchical index based on directory structure
5. It updates your specified file with the index between special markers:

\`\`\`markdown

<!-- START_GTOC -->

# Documentation Index

...

<!-- END_GTOC -->

\`\`\`

If you run the tool again, it will update the existing index section while preserving the rest of the file content.

## 📊 Example Output

Before:
\`\`\`markdown

# My Project

Some description here.

## Getting Started

...
\`\`\`

After running `gtoc generate README.md`:
\`\`\`markdown

# My Project

<!-- START_GTOC -->

# Documentation Index

- [Installation Guide](docs/installation.md)
- [API Reference](docs/api.md)

## Examples

- [Basic Example](examples/basic.md)
- [Advanced Example](examples/advanced.md)

<!-- END_GTOC -->

Some description here.

## Getting Started

...
\`\`\`

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - A Commander for modern Go CLI interactions
- All contributors who help improve this tool

---

<div align="center">
Made with ❤️ by <a href="https://github.com/lpsm-dev">LPSM</a>
</div>
