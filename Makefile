.PHONY: license tests



build: ## Build the binary file
	@echo "Building the binary file"
	 go build -ldflags="-X 'github.com/karl-cardenas-coding/mywhoop/cmd.VersionString=1.0.0'" -o=whoop -v 

license:
	@echo "Applying license headers..."
	 copywrite headers

opensource:
	@echo "Checking for open source licenses"
	~/go/bin/go-licenses report github.com/karl-cardenas-coding/mywhoop --ignore $$(go list -m) --include_tests \
		--ignore $$(tr -d ' \n' <<<"$${{ inputs.ignore-modules }}") \
		--ignore $$(tr -d ' \n' <<<"$${{ inputs.ignore-modules }}") \
		--ignore $$(go list std | awk 'NR > 1 { printf(",") } { printf("%s",$0) } END { print "" }') \
		--template=docs/open-source.tpl > docs/open-source.md

tests: ## Run tests
	@echo "Running tests"
	go test -race -shuffle on ./...


tests-coverage: ## Start Go Test with code coverage
	@echo "Running Go Tests with code coverage"
	go test -race -shuffle on -cover -coverprofile=coverage.out -covermode=atomic  ./...

view-coverage: ## View the code coverage
	@echo "Viewing the code coverage"
	go tool cover -html=coverage.out


nil: ## Check for nil errors
	@echo "Checking for nil errors"
	~/go/bin/nilaway ./...

## Cleaning

clean: ## Clean the binary file
	@echo "Cleaning the binary file"
	rm -rf dist/
	rm mywhoop_* || true


## Release

release-preview: # Create a preview release
	goreleaser release --snapshot --clean