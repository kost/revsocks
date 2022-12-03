VERSION=1.1.1

revsocks: dep
	go build

dep:
	#go get -u ./...
	go get

tools:
	go get github.com/mitchellh/gox
	get github.com/tcnksm/ghr

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
	CGO_ENABLED=0 gox -ldflags="-s -w" -output="dist/{{.Dir}}_{{.OS}}_{{.Arch}}"

draft:
	ghr -draft v$(VERSION) dist/



