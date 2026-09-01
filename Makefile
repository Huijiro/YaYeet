.PHONY: run build package

BINARY := jirolauncher
CMD := ./cmd/jirolauncher

run:
	go run $(CMD)

build:
	go build -o $(BINARY) $(CMD)

package:
	./packaging/build-appimage.sh
