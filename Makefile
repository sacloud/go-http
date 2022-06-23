#
# Copyright 2021-2022 The sacloud/go-http authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
AUTHOR          ?="The sacloud/go-http authors"
COPYRIGHT_YEAR  ?="2021-2022"
COPYRIGHT_FILES ?=$$(find . -name "*.go" -print | grep -v "/vendor/")

default: gen fmt set-license goimports lint test

.PHONY: test
test:
	TESTACC= go test ./... $(TESTARGS) -v -timeout=120m -parallel=8 -race;

.PHONY: testacc
testacc:
	TESTACC=1 go test ./... $(TESTARGS) -v -timeout=120m -parallel=8 ;

.PHONY: tools
tools:
	go install github.com/rinchsan/gosimports/cmd/gosimports@latest
	go install golang.org/x/tools/cmd/stringer@latest
	go install github.com/sacloud/addlicense@latest
	go install github.com/client9/misspell/cmd/misspell@latest
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.46.2/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.46.2

.PHONY: clean
clean:
	find . -type f -name "*_gen.go" -delete

.PHONY: gen
gen: _gen fmt goimports set-license

.PHONY: _gen
_gen:
	go generate ./...

.PHONY: goimports gosimports
goimports: gosimports
gosimports: fmt
	gosimports -l -w .

.PHONY: fmt
fmt:
	find . -name '*.go' | grep -v vendor | xargs gofmt -s -w

.PHONY: set-license
set-license:
	@addlicense -c $(AUTHOR) -y $(COPYRIGHT_YEAR) $(COPYRIGHT_FILES)

.PHONY: lint lint-go
lint: lint-go textlint

lint-go:
	golangci-lint run ./...

.PHONY: textlint
textlint:
	@docker run --rm -v $$PWD:/work -w /work ghcr.io/sacloud/textlint-action:v0.0.1 .
