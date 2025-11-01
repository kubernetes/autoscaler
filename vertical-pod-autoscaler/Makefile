
BIN_DIR := bin
TEST_DIR := test
TOOLS_DIR := hack/tools
TOOLS_BIN_DIR := $(abspath $(TOOLS_DIR)/$(BIN_DIR))
GOLANGCI_LINT := golangci-lint

GOLANGCI_LINT_KAL_BIN := golangci-kube-api-linter
GOLANGCI_LINT_KAL := $(abspath $(TOOLS_BIN_DIR)/$(GOLANGCI_LINT_KAL_BIN))

GOLANGCI_LINT_BIN := golangci-lint
GOLANGCI_LINT := $(abspath $(TOOLS_BIN_DIR)/$(GOLANGCI_LINT_BIN)-$(GOLANGCI_LINT_VER))
GOLANGCI_LINT_PKG := github.com/golangci/golangci-lint/v2/cmd/golangci-lint


$(GOLANGCI_LINT_KAL):
	cd $(TOOLS_DIR); golangci-lint custom

.PHONY: lint-api
lint-api: $(GOLANGCI_LINT_KAL)
	$(GOLANGCI_LINT_KAL) run -v --config .golangci-kal.yml $(GOLANGCI_LINT_EXTRA_ARGS)

.PHONY: lint-api-fix
lint-api-fix: $(GOLANGCI_LINT_KAL)
	GOLANGCI_LINT_EXTRA_ARGS=--fix $(MAKE) lint-api

.PHONY: show-git-info
show-git-info:
	echo '=============== local git status ==============='
	git status
	echo '=============== last commit ==============='
	git log -1
	echo '=============== bulding from the above ==============='

$(GOLANGCI_LINT): # Build golangci-lint from tools folder.
	GOBIN=$(TOOLS_BIN_DIR) $(GO_INSTALL) $(GOLANGCI_LINT_PKG) $(GOLANGCI_LINT_BIN) $(GOLANGCI_LINT_VER)

.PHONY: clean
clean: ## Remove all generated binaries
	rm -rf $(TOOLS_BIN_DIR)
