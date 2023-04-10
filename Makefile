VERSION=2.0

revsocks: dep
	go build

dep:
	#go get -u ./...
	go get

tools:
	go install github.com/mitchellh/gox@latest
	go install github.com/tcnksm/ghr@latest


ver:
	echo version $(VERSION)

gittag:
	git tag v$(VERSION)
	git push --tags origin master

clean:
	rm -rf dist

dist:
	mkdir -p dist

gox:
	CGO_ENABLED=0 gox -osarch="!darwin/386" -ldflags="-s -w" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

draft:
	ghr -draft v$(VERSION) dist/



