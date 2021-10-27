.DEFAULT_GOAL := help

INTERACTIVE := $(shell [ -t 0 ] && echo 1)

# Helper variables
V = 0
Q = $(if $(filter 1,$V),,@)
ifeq ($(INTERACTIVE),1)
	M := $(shell printf "\033[34;1m▶\033[0m")
else
	M := =>
endif

.PHONY: help
help: ## Shows this help message
	$(Q) awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_\-.]+:.*?##/ { printf "  \033[36m%-15s\033[0m\t %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
