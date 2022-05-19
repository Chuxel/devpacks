registry = ghcr.io
publisher = chuxel
repository = devpacks
buildpacks = nodejs finalize npminstall

all: build package

build:
	for buildpack in $(buildpacks); do \
		go build -o ./bin/$$buildpack/bin/build ./cmd/$$buildpack/build/main.go; \
		go build -o ./bin/$$buildpack/bin/detect ./cmd/$$buildpack/detect/main.go; \
		cp -fR ./assets/$$buildpack/* ./bin/$$buildpack/ || echo No assets to copy.; \
	done

package:
	for buildpack in $(buildpacks); do \
 		pack buildpack package "$(registry)/$(publisher)/$(repository)/buildpack-$$buildpack" --pull-policy if-not-present -p ./bin/$$buildpack; \
 	done
