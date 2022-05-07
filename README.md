# harvit

[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=for-the-badge)](https://godoc.org/github.com/mgjules/harvit)
[![Conventional Commits](https://img.shields.io/badge/Conventional%20Commits-1.0.0-yellow.svg?style=for-the-badge)](https://conventionalcommits.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg?style=for-the-badge)](LICENSE)

Harvit harvests data from different sources (e.g websites, APIs), converts and transforms it.


## Contents

- [harvit](#harvit)
  - [Contents](#contents)
  - [Requirements](#requirements)
  - [Usage](#usage)
  - [Example](#example)
    - [plan.yml](#planyml)
    - [transformers/sample.js](#transformerssamplejs)
    - [Result](#result)
  - [License](#license)
  - [Stability](#stability)


## Requirements

- [Go 1.18+](https://golang.org/doc/install)
- [Mage](https://github.com/magefile/mage) - replacement for Makefile in Go.
- [Golangci-lint](https://github.com/golangci/golangci-lint) - Fast Go linters runner.
- [Ginkgo](https://github.com/onsi/ginkgo) - Esspressive testing framework.
- [Docker](https://www.docker.com) - Containerization.

## Usage

Harvit uses a `plan` in yaml format (see example below) to define the data source, fields and the transformer to be performed.

```shell
$ ./harvit harvest [command options] plan
```

```
NAME:
   harvit harvest - Let's harvest some data!

USAGE:
   harvit harvest [command options] plan

OPTIONS:
   --debug     whether running in PROD or DEBUG mode (default: false) [$HARVIT_DEBUG]
   --help, -h  show help (default: false)
```

## Example

```shell
$ ./harvit harvest | jq
```

### plan.yml

```yaml
source: https://mgjules.dev
type: website
fields:
  - name: firstJobName
    type: raw
    selector: "#experience > div:nth-child(2) > ul > li:nth-child(1) > div.flex.flex-wrap.items-center.justify-between > h3"
  - name: secondJobStartYear
    type: datetime
    selector: "#experience > div:nth-child(2) > ul > li:nth-child(2) > div.flex.flex-wrap.items-center.justify-between > span"
    regex: \d{2}/(\d{4})\s→
    format: Y
  - name: secondJobEndDateTime
    type: datetime
    selector: "#experience > div:nth-child(2) > ul > li:nth-child(2) > div.flex.flex-wrap.items-center.justify-between > span"
    regex: →\s(?:[a-zA-Z]+|(\d{2}/\d{4}))
    format: m/Y
    timezone: Indian/Mauritius
  - name: topLinks
    type: text
    selector: "body > div.relative.px-4.pt-4.sm\\:pt-16.print\\:pt-0.sm\\:px-6.lg\\:px-8 > div.max-w-4xl.mx-auto.text-lg > div:nth-child(2) > div.flex.flex-wrap.items-center.justify-center.gap-x-4.gap-y-2.print\\:hidden > a > div > span"
  - name: experiencePlaces
    type: text
    selector: "#experience > div:nth-child(2) > ul > li > div.flex.flex-wrap.items-center.justify-between > h3"
  - name: contributionsYears
    type: datetime
    selector: "#contributions > div:nth-child(2) > ul > li > div > span"
    regex: (\d{4})
    format: Y
  - name: contributionsYearsNumbers
    type: number
    selector: "#contributions > div:nth-child(2) > ul > li > div > span"
    regex: (\d{4})
  - name: interestsTitle
    type: text
    selector: "#interests > div:nth-child(2) > ul > li > span"
transformer: transformers/sample.js
```

### transformers/sample.js

```js
data['interestsTitle'] = data['interestsTitle'].map(v => v === 'Space Exploration' ? 'SpaceX' : v);
```

### Result

```json
{
  "contributionsYears": [
    "2022-01-01T00:00:00+04:00",
    "2021-01-01T00:00:00+04:00",
    "2020-01-01T00:00:00+04:00",
    "2020-01-01T00:00:00+04:00",
    "2019-01-01T00:00:00+04:00",
    "2019-01-01T00:00:00+04:00",
    "2019-01-01T00:00:00+04:00",
    "2019-01-01T00:00:00+04:00"
  ],
  "contributionsYearsNumbers": [
    2022,
    2021,
    2020,
    2020,
    2019,
    2019,
    2019,
    2019
  ],
  "experiencePlaces": [
    "Ringier SA",
    "Bocasay",
    "La Sentinelle Digital Ltd",
    "Expat-Blog Ltd",
    "Noveo IT Ltd"
  ],
  "firstJobName": "<h3 class=\"my-0\">Ringier SA</h3>",
  "interestsTitle": [
    "SpaceX",
    "Artificial Intelligence",
    "Skateboarding",
    "Anime",
    "Gaming",
    "Movie"
  ],
  "secondJobEndDateTime": "2021-02-01T00:00:00+04:00",
  "secondJobStartYear": "2020-01-01T00:00:00+04:00",
  "topLinks": [
    "Developer",
    "Github",
    "LinkedIn",
    "Mail",
    "Mauritius"
  ]
}
```


## License

Harvit is Apache 2.0 licensed.


## Stability

This project follows [SemVer](http://semver.org/) strictly and is not yet `v1`.

Breaking changes might be introduced until `v1` is released.

This project follows the [Go Release Policy](https://golang.org/doc/devel/release.html#policy). Each major version of Go is supported until there are two newer major releases.
