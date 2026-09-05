.PHONY: run dev build package package-deb

BINARY := yayeet
CMD := ./cmd/yayeet

run:
	go run $(CMD)

dev:
	go run -ldflags="-X github.com/Huijiro/YaYeet/internal/buildinfo.UpdatesEnabled=false" $(CMD)

build:
	go build -o $(BINARY) $(CMD)

package:
	@test -n "$(VERSION)" || (echo "VERSION is required" >&2; exit 2)
	./packaging/build-appimage.sh "$(VERSION)"

package-deb:
	@test -n "$(VERSION)" || (echo "VERSION is required" >&2; exit 2)
	./packaging/build-deb.sh "$(VERSION)"
