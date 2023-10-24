.PHONY: gendocs
gendocs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs

.PHONY: test
test:
	DOCKER_API_VERSION=1.41 go test -v ./...

.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run