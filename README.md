[![ci](https://github.com/convention-change/drone-gitea-cc-release/workflows/ci/badge.svg?branch=main)](https://github.com/convention-change/drone-gitea-cc-release/actions/workflows/ci.yml)
[![GitHub license](https://img.shields.io/github/license/convention-change/drone-gitea-cc-release)](https://github.com/convention-change/drone-gitea-cc-release)
[![go mod version](https://img.shields.io/github/go-mod/go-version/convention-change/drone-gitea-cc-release?label=go.mod)](https://github.com/convention-change/drone-gitea-cc-release)
[![GoDoc](https://godoc.org/github.com/convention-change/drone-gitea-cc-release?status.png)](https://godoc.org/github.com/convention-change/drone-gitea-cc-release/)
[![GoReportCard](https://goreportcard.com/badge/github.com/convention-change/drone-gitea-cc-release)](https://goreportcard.com/report/github.com/convention-change/drone-gitea-cc-release)
[![codecov](https://codecov.io/gh/convention-change/drone-gitea-cc-release/branch/main/graph/badge.svg)](https://codecov.io/gh/convention-change/drone-gitea-cc-release)
[![docker version semver](https://img.shields.io/docker/v/convention-change/drone-gitea-cc-release?sort=semver)](https://hub.docker.com/r/convention-change/drone-gitea-cc-release/tags?page=1&ordering=last_updated)
[![docker image size](https://img.shields.io/docker/image-size/convention-change/drone-gitea-cc-release)](https://hub.docker.com/r/convention-change/drone-gitea-cc-release)
[![docker pulls](https://img.shields.io/docker/pulls/convention-change/drone-gitea-cc-release)](https://hub.docker.com/r/convention-change/drone-gitea-cc-release/tags?page=1&ordering=last_updated)
[![GitHub release](https://img.shields.io/github/v/release/convention-change/drone-gitea-cc-release?style=social)](https://github.com/convention-change/drone-gitea-cc-release/releases)

## for what

- drone CI release for [gitea](https://docs.gitea.com/) and support [conventional-commits](https://www.conventionalcommits.org/) log

## Contributing

[![Contributor Covenant](https://img.shields.io/badge/contributor%20covenant-v1.4-ff69b4.svg)](.github/CONTRIBUTING_DOC/CODE_OF_CONDUCT.md)
[![GitHub contributors](https://img.shields.io/github/contributors/convention-change/drone-gitea-cc-release)](https://github.com/convention-change/drone-gitea-cc-release/graphs/contributors)

We welcome community contributions to this project.

Please read [Contributor Guide](.github/CONTRIBUTING_DOC/CONTRIBUTING.md) for more information on how to get started.

请阅读有关 [贡献者指南](.github/CONTRIBUTING_DOC/zh-CN/CONTRIBUTING.md) 以获取更多如何入门的信息## Features

## Features

- [ ] more perfect test case coverage
- [ ] more perfect benchmark case
- more see [features/README.md](features/README.md)

## usage

- [x] read [conventional-commits](https://www.conventionalcommits.org/) log
- [x] release for [gitea](https://docs.gitea.com/)

### Pipeline Settings (.drone.yml)

`1.x`

```yaml
steps:
  - name: drone-gitea-cc-release
    image: convention-change/drone-gitea-cc-release:latest
    pull: if-not-exists
    settings:
      debug: false
      webhook:
        # https://docs.drone.io/pipeline/environment/syntax/#from-secrets
        from_secret: webhook_token
      msg_type: your-message-type
      timeout_second: 10 # default 10
    when:
      event: # https://docs.drone.io/pipeline/exec/syntax/conditions/#by-event
        - promote
        - rollback
        - push
        - pull_request
        - tag
      status: # only support failure/success,  both open will send anything
        - failure
        # - success
```

### install cli tools

```bash
# install at ${GOPATH}/bin
$ go install -v github.com/convention-change/drone-gitea-cc-release/cmd/drone-gitea-cc-release@latest
# install version v1.0.0
$ go install -v github.com/convention-change/drone-gitea-cc-release/cmd/drone-gitea-cc-release@v1.0.0
```

or download by [github releases](https://github.com/convention-change/drone-gitea-cc-release/releases)

## env

- minimum go version: go 1.18
- change `go 1.18`, `^1.18`, `1.18.10` to new go version

### libs

| lib                                        | version |
|:-------------------------------------------|:--------|
| https://github.com/stretchr/testify        | v1.8.4  |
| https://github.com/sebdah/goldie           | v2.5.3  |
| https://github.com/joho/godotenv           | v1.4.0  |
| https://github.com/sinlov/drone-info-tools | v1.21.0 |
| https://github.com/urfave/cli/v2           | v2.25.7 |

# dev

## depends

in go mod project

```bash
# warning use private git host must set
# global set for once
# add private git host like github.com to evn GOPRIVATE
$ go env -w GOPRIVATE='github.com'
# use ssh proxy
# set ssh-key to use ssh as http
$ git config --global url."git@github.com:".insteadOf "https://github.com/"
# or use PRIVATE-TOKEN
# set PRIVATE-TOKEN as gitlab or gitea
$ git config --global http.extraheader "PRIVATE-TOKEN: {PRIVATE-TOKEN}"
# set this rep to download ssh as https use PRIVATE-TOKEN
$ git config --global url."ssh://github.com/".insteadOf "https://github.com/"

# before above global settings
# test version info
$ git ls-remote -q https://github.com/convention-change/drone-gitea-cc-release.git

# test depends see full version
$ go list -mod=readonly -v -m -versions github.com/convention-change/drone-gitea-cc-release
# or use last version add go.mod by script
$ echo "go mod edit -require=$(go list -mod=readonly -m -versions github.com/convention-change/drone-gitea-cc-release | awk '{print $1 "@" $NF}')"
$ echo "go mod vendor"
```

```bash
make init dep
```

- see help

```bash
$ make devHelp
```

- test code

add env then test

```bash
export PLUGIN_MSG_TYPE=post \
  export PLUGIN_WEBHOOK=7138d7b3-abc
```

```bash
$ make test testBenchmark
```

- full env example

```bash
export PLUGIN_MSG_TYPE= \
  export PLUGIN_WEBHOOK= \
  export DRONE_REPO=convention-change/drone-gitea-cc-release \
  export DRONE_REPO_NAME=drone-gitea-cc-release \
  export DRONE_REPO_NAMESPACE=convention-change \
  export DRONE_REMOTE_URL=https://github.com/convention-change/drone-gitea-cc-release \
  export DRONE_REPO_OWNER=convention-change \
  export DRONE_COMMIT_AUTHOR=convention-change \
  export DRONE_COMMIT_AUTHOR_AVATAR=  \
  export DRONE_COMMIT_AUTHOR_EMAIL=convention-changegmppt@gmail.com \
  export DRONE_COMMIT_BRANCH=main \
  export DRONE_COMMIT_LINK=https://github.com/convention-change/drone-gitea-cc-release/commit/68e3d62dd69f06077a243a1db1460109377add64 \
  export DRONE_COMMIT_SHA=68e3d62dd69f06077a243a1db1460109377add64 \
  export DRONE_COMMIT_REF=refs/heads/main \
  export DRONE_COMMIT_MESSAGE="mock message commit" \
  export DRONE_STAGE_STARTED=1674531206 \
  export DRONE_STAGE_FINISHED=1674532106 \
  export DRONE_BUILD_STATUS=success \
  export DRONE_BUILD_NUMBER=1 \
  export DRONE_BUILD_LINK=https://drone.xxx.com/convention-change/drone-gitea-cc-release/1 \
  export DRONE_BUILD_EVENT=push \
  export DRONE_BUILD_STARTED=1674531206 \
  export DRONE_BUILD_FINISHED=1674532206
```

- then run

```bash
$ make run
```

- ci to fast check

```bash
$ make ci
```

## docker

```bash
# then test build as test/Dockerfile
$ make dockerTestRestartLatest
# if run error
# like this error
# err: missing webhook, please set webhook
#  fix env settings then test

# see run docker fast
$ make dockerTestRunLatest

# clean test build
$ make dockerTestPruneLatest

# see how to use
$ docker run --rm convention-change/drone-gitea-cc-release:latest -h
```
