.PHONY: run build package package-deb

BINARY := yayeet
CMD := ./cmd/yayeet

run:
	go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

package:
	./packaging/build-appimage.sh

package-deb:
	@test -n "$(VERSION)" || (echo "VERSION is required" >&2; exit 2)
	./packaging/build-deb.sh "$(VERSION)"
