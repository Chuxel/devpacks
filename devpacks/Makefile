registry = ghcr.io
publisher = chuxel
repository = devpacks
buildpacks = nodejs finalize npminstall npmbuild npmstart cpython pythonutils pipinstall goutils procfile
buildpack-stages = build detect
extractor-archs = amd64 arm64
extractor-os = linux darwin windows

all: buildpacks build-extractor

build: build-buildpacks build-extractor

package: package-buildpacks

buildpacks: build-buildpacks package-buildpacks

build-buildpacks:
	for buildpack in $(buildpacks); do \
		for stage in $(buildpack-stages); do \
			GOARCH="amd64" GOOS="linux" go build -o ./bin/$$buildpack/bin/$$stage ./cmd/$$buildpack/$$stage/main.go; \
		done; \
		cp -fR ./assets/$$buildpack/* ./bin/$$buildpack/ || echo No assets to copy.; \
	done

build-extractor:
	for os in $(extractor-os); do \
		for arch in $(extractor-archs); do \
			GOARCH="$$arch" GOOS="$$os" go build -o ./bin/devcontainer-extractor/devcontainer-extractor-$$os-$$arch ./cmd/devcontainer-extractor/main.go; \
		done; \
	done
	for arch in $(extractor-archs); do \
		rm -f ./bin/devcontainer-extractor/devcontainer-extractor-windows-$$arch.exe; \
		mv ./bin/devcontainer-extractor/devcontainer-extractor-windows-$$arch ./bin/devcontainer-extractor/devcontainer-extractor-windows-$$arch.exe; \
	done; \

package-buildpacks:
	for buildpack in $(buildpacks); do \
		pack buildpack package "$(registry)/$(publisher)/$(repository)/buildpack-$$buildpack" --pull-policy if-not-present -p ./bin/$$buildpack; \
	done

publish-buildpacks:
	for buildpack in $(buildpacks); do \
		docker push "$(registry)/$(publisher)/$(repository)/buildpack-$$buildpack"; \
	done