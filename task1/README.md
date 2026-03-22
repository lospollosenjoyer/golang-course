# pugovkin

> _**pugovkin** is inspired by poking nose into everything_

pugovkin is a simple CLI-tool to get some basic info from public GitHub repositories

### Requirements

- Go (1.25.2)

### Install

```sh
git clone https://github.com/lospollosenjoyer/golang-course.git
git switch task1
cd task1
```

### Build

```sh
make build
```

### Run

```sh
./pugovkin [-t <timeout in seconds>] [<repo-url> ...]
```

### Using GitHub fine-grained token

You may use your own GitHub fine-grained token for greater rate limits and access to private repositories. For that, you should add it as an environment variable `PUGOVKIN_GITHUB_TOKEN`:

```sh
export PUGOVKIN_GITHUB_TOKEN=<your-token-here>
./pugovkin ...
```

or just:

```sh
PUGOVKIN_GITHUB_TOKEN=<your-token-here> ./pugovkin ...
```
