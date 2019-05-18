VERSION=$(shell git describe --tags)

build:
	go build -ldflags "-X github.com/lemonade-command/lemonade/lemon.Version=$(VERSION)"

install:
	go install -ldflags "-X github.com/lemonade-command/lemonade/lemon.Version=$(VERSION)"

release:
	xgo --targets=windows/386,windows/amd64 -ldflags="-s -w -H=windowsgui -X github.com/hanxi/lemonade/lemon.Version=$(VERSION)" .
	xgo --targets=linux/amd64,darwin/amd64 -ldflags="-s -w -X github.com/hanxi/lemonade/lemon.Version=$(VERSION)" .
	mkdir -p dist/lemonade_windows_386/
	mkdir -p dist/lemonade_windows_amd64/
	mkdir -p dist/lemonade_linux_amd64/
	mkdir -p dist/lemonade_darwin_amd64/
	mv -f lemonade-windows-4.0-386.exe dist/lemonade_windows_386/lemonade.exe
	mv -f lemonade-windows-4.0-amd64.exe dist/lemonade_windows_amd64/lemonade.exe
	mv -f lemonade-linux-amd64 dist/lemonade_linux_amd64/lemonade
	mv -f lemonade-darwin-10.6-amd64 dist/lemonade_darwin_amd64/lemonade
	zip pkg/lemonade_windows_386.zip dist/lemonade_windows_386/lemonade.exe -j
	zip pkg/lemonade_windows_amd64.zip dist/lemonade_windows_amd64/lemonade.exe -j
	tar zcvf pkg/lemonade_linux_amd64.tar.gz -C dist/lemonade_linux_amd64/ lemonade
	tar zcvf pkg/lemonade_darwin_amd64.tar.gz -C dist/lemonade_darwin_amd64/ lemonade

clean:
	rm -rf dist/
	rm -f pkg/*.tar.gz pkg/*.zip
