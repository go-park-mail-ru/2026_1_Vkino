PHONY: init
init:
	cp .github/hooks/* .git/hooks
	chmod +x .git/hooks/*

test:
	go test ./... -cover

cover:
	-go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

clean:
	rm -f coverage.out

cover-total:
	@echo "=== Total project coverage ==="
	@go test ./... -coverprofile=coverage.out > /dev/null 2>&1 || true
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}'
