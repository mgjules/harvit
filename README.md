# harvit

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](https://godoc.org/github.com/mgjules/harvit)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge)](LICENSE)

Harvit scrapes data from different sources (e.g websites, APIs) and transforms it. 


## Contents

  - [Requirements](#requirements)
  - [Usage](#usage)
  - [License](#license)
  - [Stability](#stability)


## Requirements

- [Go 1.18+](https://golang.org/doc/install)
- [Mage](https://github.com/magefile/mage) - replacement for Makefile in Go.
- [Golangci-lint](https://github.com/golangci/golangci-lint) - Fast Go linters runner.
- [Ginkgo](https://github.com/onsi/ginkgo) - Esspressive testing framework.
- [Docker](https://www.docker.com) - Containerization.

## Usage

1. Download the latest release for your OS and architecture and extract the binary.
2. Run the binary.
  ```sh
  $ ./harvit
  ```


## License

Harvit is Apache 2.0 licensed.


## Stability

This project follows [SemVer](http://semver.org/) strictly and is not yet `v1`.

Breaking changes might be introduced until `v1` is released.

This project follows the [Go Release Policy](https://golang.org/doc/devel/release.html#policy). Each major version of Go is supported until there are two newer major releases.
