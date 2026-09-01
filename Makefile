.PHONY: run build package

BINARY := yayeet
CMD := ./cmd/yayeet

run:
	go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

package:
	./packaging/build-appimage.sh
