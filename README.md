[![ci](https://github.com/convention-change/drone-gitea-cc-release/workflows/ci/badge.svg?branch=main)](https://github.com/convention-change/drone-gitea-cc-release/actions/workflows/ci.yml)

[![go mod version](https://img.shields.io/github/go-mod/go-version/convention-change/drone-gitea-cc-release?label=go.mod)](https://github.com/convention-change/drone-gitea-cc-release)
[![GoDoc](https://godoc.org/github.com/convention-change/drone-gitea-cc-release?status.png)](https://godoc.org/github.com/convention-change/drone-gitea-cc-release/)
[![GoReportCard](https://goreportcard.com/badge/github.com/convention-change/drone-gitea-cc-release)](https://goreportcard.com/report/github.com/convention-change/drone-gitea-cc-release)

[![docker hub version semver](https://img.shields.io/docker/v/sinlov/drone-gitea-cc-release?sort=semver)](https://hub.docker.com/r/sinlov/drone-gitea-cc-release/tags?page=1&ordering=last_updated)
[![docker hub image size](https://img.shields.io/docker/image-size/sinlov/drone-gitea-cc-release)](https://hub.docker.com/r/sinlov/drone-gitea-cc-release)
[![docker hub image pulls](https://img.shields.io/docker/pulls/sinlov/drone-gitea-cc-release)](https://hub.docker.com/r/sinlov/drone-gitea-cc-release/tags?page=1&ordering=last_updated)

[![GitHub license](https://img.shields.io/github/license/convention-change/drone-gitea-cc-release)](https://github.com/convention-change/drone-gitea-cc-release)
[![GitHub release](https://img.shields.io/github/v/release/convention-change/drone-gitea-cc-release?style=social)](https://github.com/convention-change/drone-gitea-cc-release/releases)

## for what

- drone CI release for [gitea](https://docs.gitea.com/) and `must with git tag push`
  support [conventional-commits](https://www.conventionalcommits.org/) log

## Contributing

[![Contributor Covenant](https://img.shields.io/badge/contributor%20covenant-v1.4-ff69b4.svg)](.github/CONTRIBUTING_DOC/CODE_OF_CONDUCT.md)
[![GitHub contributors](https://img.shields.io/github/contributors/convention-change/drone-gitea-cc-release)](https://github.com/convention-change/drone-gitea-cc-release/graphs/contributors)

We welcome community contributions to this project.

Please read [Contributor Guide](.github/CONTRIBUTING_DOC/CONTRIBUTING.md) for more information on how to get started.

请阅读有关 [贡献者指南](.github/CONTRIBUTING_DOC/zh-CN/CONTRIBUTING.md) 以获取更多如何入门的信息## Features

## Features

- [X] release for [gitea](https://docs.gitea.com/) and `only support with git tag` pushed
- [X] gitea client token use [Access Token](https://docs.gitea.com/development/api-usage#authentication)
- [X] upload release files by glob pattern
- [X] support upload check sum file
- [X] support [conventional-commits](https://www.conventionalcommits.org/) log
- [X] publish package support
    - [X] go package publish, must update gitea:1.20.1+ and set `gitea_publish_package_go` to `true` doc see [https://docs.gitea.com/zh-cn/usage/packages/go](https://docs.gitea.com/zh-cn/usage/packages/go)
    - [X] when go package publish, can remove folder by `gitea_publish_go_remove_paths` or `PLUGIN_GITEA_PUBLISH_GO_REMOVE_PATHS` (support v1.3.+)
- [ ] more perfect test case coverage
- more see [CHANGELOG.md](https://github.com/convention-change/drone-gitea-cc-release/blob/main/CHANGELOG.md)

## usage

- [x] read [conventional-commits](https://www.conventionalcommits.org/) log
- [x] get your [gitea Access Token](https://docs.gitea.com/development/api-usage#authentication)
  for `release_gitea_api_key`
- [x] add `.drone.yml` config as pipeline

### Pipeline Settings (.drone.yml)

`1.x`

- fast use as docker

```yaml
kind: pipeline
type: docker
name: basic-docker

steps:
  - name: gitea-cc-release
    image: sinlov/drone-gitea-cc-release:1.3.1 # https://hub.docker.com/r/sinlov/drone-gitea-cc-release/tags
    pull: if-not-exists
    settings:
      prerelease: true # default true
      release_gitea_base_url: https://gitea.xxxx.com
      release_gitea_api_key:
        from_secret: release_gitea_api_key
      release_gitea_files: # release as files by glob pattern
        - "doc/*.md"
      release_gitea_files_checksum: # generate specific checksums, support [ md5 sha1 sha256 sha512 adler32 crc32 blake2b blake2s ]
        - md5
        - sha1
        - sha256
      release_gitea_file_exists_do: "overwrite" # default skip, support [ fail skip overwrite ]
      release_gitea_note_by_convention_change: true # default false, like tools https://github.com/convention-change/convention-change-log read change log
      # gitea_publish_package_go: true # gitea 1.20.1+ support publish go package, default false
      # gitea_publish_package_path_go: "" # publish go package dir to find go.mod, if not set will use git root path
      # gitea_publish_go_remove_paths: ['dist'] # publish go package remove paths, this path under gitea_publish_package_path_go, vars like dist,target
    when:
      event: # https://docs.drone.io/pipeline/exec/syntax/conditions/#by-event
        - tag
```

- full config

```yaml
kind: pipeline
type: docker
name: basic-docker

steps:
  - name: gitea-cc-release
    image: sinlov/drone-gitea-cc-release:1.3.1 # https://hub.docker.com/r/sinlov/drone-gitea-cc-release/tags
    pull: if-not-exists
    settings:
      # debug: true # plugin debug switch
      prerelease: true # default true
      draft: true # default false
      release_gitea_base_url: https://gitea.xxxx.com
      release_gitea_api_key:
        from_secret: release_gitea_api_key
      # release_gitea_insecure: false # default false, visit base-url via insecure https protocol
      release_gitea_file_root_path: "" # release as files by glob pattern root path, if not setting will get cwd folder by PLUGIN_RELEASE_GITEA_ROOT_FOLDER_PATH
      release_gitea_files: # release as files by glob pattern
        - "doc/*.md"
        - "**/*.zip"
      release_gitea_files_checksum: # generate specific checksums, support [ md5 sha1 sha256 sha512 adler32 crc32 blake2b blake2s ]
        - md5
        - sha1
        - sha256
      release_gitea_file_exists_do: "overwrite" # default skip, support [ fail skip overwrite ]
      release_gitea_note_by_convention_change: true # default false, like tools https://github.com/convention-change/convention-change-log read change log
      # release_read_change_log_file: CHANGELOG.md # default CHANGELOG.md
      # release_gitea_title: "" # if set release_gitea_note_by_convention_change true will cover this, use "" will use tag
      # release_gitea_note: "" # if set release_gitea_note_by_convention_change true will cover this
      gitea_publish_package_go: true # gitea 1.20.1+ support publish go package, default false
      gitea_publish_package_path_go: "" # publish go package dir to find go.mod, if not set will use git root path
      gitea_publish_go_remove_paths: ['dist'] # publish go package remove paths, this path under gitea_publish_package_path_go, vars like dist,target
    when:
      event: # https://docs.drone.io/pipeline/exec/syntax/conditions/#by-event
        - tag
```

- `1.x` drone-exec only support env
- download
  by [https://github.com/convention-change/drone-gitea-cc-release/releases](https://github.com/convention-change/drone-gitea-cc-release/releases)
  to get platform binary, then has local path
- binary path like `C:\Drone\drone-runner-exec\plugins\drone-feishu-group-robot.exe` can be drone run env
  like `EXEC_DRONE_GITEA-CC-RELEASE`
- env:EXEC_DRONE_GITEA-CC-RELEASE can set at file which define
  as [DRONE_RUNNER_ENVFILE](https://docs.drone.io/runner/exec/configuration/reference/drone-runner-envfile/) to support
  each platform to send feishu message

```yaml
steps:
  - name: notification-feishu-group-robot-exec # must has env EXEC_DRONE_GITEA-CC-RELEASE and exec tools
    environment:
      # PLUGIN_DEBUG: false
      PLUGIN_PRERELEASE: true # default true
      PLUGIN_DRAFT: true # default false
      PLUGIN_RELEASE_GITEA_BASE_URL: https://gitea.xxxx.com
      PLUGIN_RELEASE_GITEA_API_KEY:
        from_secret: release_gitea_api_key
      # PLUGIN_RELEASE_GITEA_INSECURE: false # default false, visit base-url via insecure https protocol
      # PLUGIN_RELEASE_GITEA_FILE_ROOT_PATH: "" # release as files by glob pattern root path, if not setting will get cwd folder by PLUGIN_RELEASE_GITEA_ROOT_FOLDER_PATH
      PLUGIN_RELEASE_GITEA_FILES: "doc/*.md,**/*.zip" # release as files by glob pattern
      PLUGIN_RELEASE_GITEA_FILES_CHECKSUM: "md5,sha1,sha256" # generate specific checksums, support [ md5 sha1 sha256 sha512 adler32 crc32 blake2b blake2s ]
      PLUGIN_RELEASE_GITEA_FILE_EXISTS_DO: "overwrite" # default skip, support [ fail skip overwrite ]
      PLUGIN_RELEASE_GITEA_NOTE_BY_CONVENTION_CHANGE: true # default false, like tools https://github.com/convention-change/convention-change-log read change log
      # PLUGIN_RELEASE_READ_CHANGE_LOG_FILE: CHANGELOG.md # default CHANGELOG.md
      PLUGIN_RELEASE_GITEA_TITLE: "" # if set release_gitea_note_by_convention_change true will cover this, use "" will use tag
      PLUGIN_RELEASE_GITEA_NOTE: "" # if set release_gitea_note_by_convention_change true will cover this
      # PLUGIN_GITEA_PUBLISH_PACKAGE_GO: true # gitea 1.20.1+ support publish go package, default false
      # PLUGIN_GITEA_PUBLISH_PACKAGE_PATH_GO: "" # publish go package dir to find go.mod, if not set will use git root path
      # PLUGIN_GITEA_PUBLISH_GO_REMOVE_PATHS: "dist" # publish go package remove paths, this path under gitea_publish_package_path_go, vars like dist,target
    commands:
      - ${EXEC_DRONE_GITEA-CC-RELEASE} `
        ""
    when:
      event: # https://docs.drone.io/pipeline/exec/syntax/conditions/#by-event
        - tag
```

### install cli tools

```bash
# install at ${GOPATH}/bin
$ go install -v github.com/convention-change/drone-gitea-cc-release/cmd/drone-gitea-cc-release@latest
# install version v1.3.1
$ go install -v github.com/convention-change/drone-gitea-cc-release/cmd/drone-gitea-cc-release@v1.3.1
```

or download by [github releases](https://github.com/convention-change/drone-gitea-cc-release/releases)

## env

- minimum go version: go 1.20
- change `go 1.20`, `^1.20`, `1.20.7` to new go version

### libs

| lib                                 | version |
|:------------------------------------|:--------|
| https://github.com/stretchr/testify | v1.8.4  |
| https://github.com/sebdah/goldie    | v2.5.3  |
| https://github.com/joho/godotenv    | v1.4.0  |
| https://gitea.com/gitea/go-sdk      | v0.15.1 |

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
