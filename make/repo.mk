REPO_ROOT := $(CURDIR)

LOCAL_DIR := $(REPO_ROOT)/.local

GIT_COMMIT := $(shell git rev-parse "HEAD^{commit}")
export GIT_TAG ?= $(shell git describe --tags "$(GIT_COMMIT)^{commit}" --match v* --abbrev=0 2>/dev/null)
export GIT_CURRENT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)

export GITHUB_ORG ?= $(shell grep -E '^module ' $(REPO_ROOT)/go.mod | cut -d'/' -f2)
export GITHUB_REPOSITORY ?= $(shell grep -E '^module ' $(REPO_ROOT)/go.mod | cut -d'/' -f3)

ifneq ($(shell git status --porcelain 2>/dev/null; echo $$?), 0)
	export GIT_TREE_STATE := dirty
else
	export GIT_TREE_STATE :=
endif
