# Changelog

All notable changes to this project will be documented in this file. See [convention-change-log](https://github.com/convention-change/convention-change-log) for commit guidelines.

## [1.1.0](https://github.com/convention-change/drone-gitea-cc-release/compare/1.0.1...v1.1.0) (2023-08-06)

### ✨ Features

* pLUGIN_GITEA_PUBLISH_PACKAGE_GO PLUGIN_GITEA_PUBLISH_PACKAGE_PATH_GO gitea 1.20.1+ support ([4d985641](https://github.com/convention-change/drone-gitea-cc-release/commit/4d985641ac1df0d99ee9f872f6b05bcd7a0f4f16))

* suport release_gitea_file_root_path or PLUGIN_RELEASE_GITEA_FILE_ROOT_PATH to change release ([d078b1f9](https://github.com/convention-change/drone-gitea-cc-release/commit/d078b1f9c0d46f569f814b36e5943f93544f2420))

* add gitea_cc_release_plugin has custom ApiRequest for api request ([8809d578](https://github.com/convention-change/drone-gitea-cc-release/commit/8809d578b698ad004443a7846f5334ea4ed8b08f))

* gitea_cc_release_plugin.CreateGoModZipFromDir for build go mod zip file ([0c719b96](https://github.com/convention-change/drone-gitea-cc-release/commit/0c719b9695a0ae077fb9ef5490ea5397293115e7))

### 👷‍ Build System

* change to go1.20.7 to build ([94dfb4f8](https://github.com/convention-change/drone-gitea-cc-release/commit/94dfb4f8a2d28ca063c87ac377cac60d534a6807))

## [1.0.1](https://github.com/convention-change/drone-gitea-cc-release/compare/1.0.0...v1.0.1) (2023-08-05)

### 👷‍ Build System

* update to github.com/convention-change/convention-change-log v1.3.1 ([43706504](https://github.com/convention-change/drone-gitea-cc-release/commit/4370650407fae36cbbe0def8a963206594c0a2b8))

## 1.0.0 (2023-08-04)

### ✨ Features

* add IsBuildDebugOpen to support open at drone build DEBUG ([b604d923](https://github.com/convention-change/drone-gitea-cc-release/commit/b604d923a06a69e36f88667f0e2c24cfc89b5492))

* add full of release and file upload support ([b7a36c71](https://github.com/convention-change/drone-gitea-cc-release/commit/b7a36c71ac6bd1ca05f5d50cdb3266a3711ec394))

* use drone-info-tools v1.25.0 ([5b65bede](https://github.com/convention-change/drone-gitea-cc-release/commit/5b65bedec231d6840b8d9db2b4d8d032dea28bf3))

* clone by git.NewRepositoryClone by Drone.Repo.HttpUrl ([db2035c8](https://github.com/convention-change/drone-gitea-cc-release/commit/db2035c86c933bd070a80e80c58fcdefdfb2b7c2))

### ♻ Refactor

* update const of common flat ([e4a61459](https://github.com/convention-change/drone-gitea-cc-release/commit/e4a61459bab7816f8553c79ea9674e0d448a02dc))

### 👷‍ Build System

* github.com/sinlov/drone-info-tools v1.30.0 ([1a88e1de](https://github.com/convention-change/drone-gitea-cc-release/commit/1a88e1de2af28cae7ad76f9ec0de94777e14e297))

* bump github.com/sinlov-go/go-git-tools from 1.5.0 to 1.8.1 ([0b7efc15](https://github.com/convention-change/drone-gitea-cc-release/commit/0b7efc158e649680a530de398a5477457d6171b5))

* bump github.com/go-git/go-git/v5 from 5.7.0 to 5.8.1 ([41d423ec](https://github.com/convention-change/drone-gitea-cc-release/commit/41d423ecd3a208054069869db3dd0066c3239818))

* test TagLatestByCommitTime CommitLatestTagByTime and git_info.RepositoryFistRemoteInfo ([9634b815](https://github.com/convention-change/drone-gitea-cc-release/commit/9634b815404b352a8ff7797aba78cb2c1cfec387))
