# Building and Testing

## Prerequisites

* Go 1.15
* Docker ([`docker login`](https://help.github.com/en/packages/using-github-packages-with-your-projects-ecosystem/configuring-docker-for-use-with-github-packages#authenticating-to-github-packages) with your [GitHub token](https://help.github.com/en/github/authenticating-to-github/creating-a-personal-access-token-for-the-command-line#creating-a-token))
* Docker-Compose
* Make
* bash

## Targets

```
# run everything
make all

# linters
make checks

# unit tests
make unit-test

# BDD tests
make bdd-test
```
