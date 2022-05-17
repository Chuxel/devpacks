buildpacks = nodejs finalize

all: build

build:
	for buildpack in $(buildpacks); do \
		go build -o ./bin/$$buildpack/bin/build ./cmd/$$buildpack/build/main.go; \
		go build -o ./bin/$$buildpack/bin/detect ./cmd/$$buildpack/detect/main.go; \
		cp -fR ./assets/$$buildpack/* ./bin/$$buildpack/ || echo No assets to copy.; \
	done
