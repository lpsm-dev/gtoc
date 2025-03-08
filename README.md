GTOC is a CLI tool that generates a hierarchical index of markdown files in your Git repository and updates a specified markdown file with the generated index. It respects `.gitignore` rules and provides various customization options.

## Features

- Generate a hierarchical index of markdown files in your Git repository.
- Update a specified markdown file with the generated index.
- Respects `.gitignore` rules.
- Customizable directory depth and file patterns.
- Option to exclude specific paths.
- Dry-run mode to preview changes without writing.

## Installation

To install GTOC, you need to have Go installed on your machine. Then, you can install GTOC using the following command:

```sh
go install github.com/lpsm-dev/gtoc@latest
```

## Usage

GTOC provides a `generate` command to generate and update the markdown index. Below are some examples of how to use it:

### Basic Usage

Generate a markdown index and update the `README.md` file:

```sh
gtoc generate --file README.md
```

### Custom Depth and Pattern

Generate a markdown index with a maximum directory depth of 2 and update the `docs/index.md` file:

```sh
gtoc generate --file docs/index.md --depth 2 --pattern "docs/**/*.md"
```

### Exclude Paths

Generate a markdown index and exclude specific paths:

```sh
gtoc generate --file README.md --exclude "docs/exclude-this.md,docs/exclude-that.md"
```

### Dry-Run Mode

Preview the changes without writing to the file:

```sh
gtoc generate --file README.md --dry-run
```

## Command-Line Options

- `--file`: Path to the markdown file to update (required).
- `--depth`: Maximum directory depth (0 for unlimited).
- `--pattern`: Glob pattern to filter markdown files (default: `**/*.md`).
- `--exclude`: Comma-separated list of paths to exclude.
- `--dry-run`: Preview changes without writing.

## Example

Here is an example of how the generated index might look in your `README.md` file:

```md
<!-- START_MDTOC -->

# Documentation Index

## docs

- [Introduction](docs/introduction.md)
- [Getting Started](docs/getting-started.md)
- [API Reference](docs/api-reference.md)

## guides

- [User Guide](guides/user-guide.md)
- [Developer Guide](guides/developer-guide.md)

<!-- END_MDTOC -->
```

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue if you encounter any problems or have suggestions for improvements.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Acknowledgements

- [Cobra](https://github.com/spf13/cobra) - A library for creating powerful modern CLI applications.
- [Go](https://golang.org) - The Go programming language.
