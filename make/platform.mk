OS := $(shell uname -s)
ifeq ($(OS), Darwin)
  BREW_PREFIX := $(shell brew --prefix &>/dev/null)
  ifeq ($(BREW_PREFIX),)
    $(error Unable to discover brew prefix - do you have brew installed? See https://brew.sh/ for details of how to install)
  endif

  GNUBIN_PATH := $(BREW_PREFIX)/opt/coreutils/libexec/gnubin
  ifeq ($(wildcard $(GNUBIN_PATH)/*),)
    $(error Cannot find GNU coreutils - have you installed them via brew? See https://formulae.brew.sh/formula/coreutils for details)
  endif
  export PATH := $(GNUBIN_PATH):$(PATH)
endif
