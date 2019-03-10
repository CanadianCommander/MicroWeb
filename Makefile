SHELL = /bin/bash
GOBUILD = go build
GOTEST = go test
GOGET = go get
BINARY_FILE = microweb.a
GOBUILD_FLAGS = -o $(BINARY_FILE)
GOTEST_FLAGS = -v -timeout=60s -count 1

# build microweb
MAIN_PKG = ./cmd/microweb/
SRC_GO = ./cmd/microweb/*.go ./pkg/cache/*.go ./pkg/database/*.go ./pkg/logger/*.go ./pkg/mwSettings/*.go ./pkg/pluginUtil/*.go ./pkg/templateHelper/*.go ./pkg/session/*.go

# test (./cmd/microweb must come first)
TEST_PKGS = ./cmd/microweb ./pkg/cache ./pkg/logger ./pkg/mwSettings ./pkg/templateHelper ./pkg/session

$(BINARY_FILE): $(SRC_GO)
	$(GOBUILD) $(GOBUILD_FLAGS) $(MAIN_PKG)

.PHONY: build
build: $(BINARY_FILE)

.PHONY: clean
clean:
	rm $(BINARY_FILE)

.PHONY: test
test:
	$(GOTEST) $(GOTEST_FLAGS) ./cmd/microweb #<- this installs testing infrastructure and must be run first
	$(GOTEST) $(GOTEST_FLAGS) $(TEST_PKGS) || (echo '=== TEST FAILED ==='; exit 1) && echo "=== TEST PASS ==="

.PHONY: getdep
getdep:
	$(GOGET) $(MAIN_PKG)

.PHONY: install
install:
	sudo -E ./install/install.sh
