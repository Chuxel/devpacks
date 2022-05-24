registry = ghcr.io
publisher = chuxel
repository = devpacks
buildpacks = nodejs finalize npminstall npmstart cpython pythonutils pipinstall
buildpack-stages = build detect

all: build package

build: build-buildpacks build-extractor

package: package-buildpacks

build-buildpacks:
	for buildpack in $(buildpacks); do \
		for stage in $(buildpack-stages); do \
			go build -o ./bin/$$buildpack/bin/$$stage ./cmd/$$buildpack/$$stage/main.go; \
		done; \
		cp -fR ./assets/$$buildpack/* ./bin/$$buildpack/ || echo No assets to copy.; \
	done

build-extractor:
	go build -o ./bin/devcontainer-extractor ./cmd/devcontainer-extractor/main.go

package-buildpacks:
	for buildpack in $(buildpacks); do \
 		pack buildpack package "$(registry)/$(publisher)/$(repository)/buildpack-$$buildpack" --pull-policy if-not-present -p ./bin/$$buildpack; \
 	done
