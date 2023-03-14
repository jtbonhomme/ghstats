# ghstats

Github Statistics

## Installation

```sh
go install github.com/jtbonhomme/ghstats/cmd/ghstats@latest
export GH_API_TOKEN=<PERSONAL ACCESS TOKEN>
```

## Usage

```
Usage of ghstats
  -o string
        organisation name
  -r string
        repository list separed by comma
```

Example:

```sh
ghstats -o orga1 -r repos1,repos2
```

