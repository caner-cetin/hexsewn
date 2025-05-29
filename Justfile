#!/usr/bin/env just --justfile
default: build

name := "hexsewn"

build_dir := "dist"
build_flags := "-trimpath -ldflags='-s -w'"


list:
    @just --list

clean:
    rm -rf {{build_dir}}

setup:
    mkdir -p {{build_dir}}

tidy:
    go mod tidy


build: clean setup tidy
    #!/usr/bin/env sh
    GOOS=linux GOARCH=amd64 go build {{build_flags}} -o {{build_dir}}/{{name}}-linux-amd64
    GOOS=linux GOARCH=arm64 go build {{build_flags}} -o {{build_dir}}/{{name}}-linux-arm64

    GOOS=darwin GOARCH=amd64 go build {{build_flags}} -o {{build_dir}}/{{name}}-darwin-amd64
    GOOS=darwin GOARCH=arm64 go build {{build_flags}} -o {{build_dir}}/{{name}}-darwin-arm64

    GOOS=windows GOARCH=amd64 go build {{build_flags}} -o {{build_dir}}/{{name}}-windows-amd64.exe
    GOOS=windows GOARCH=arm64 go build {{build_flags}} -o {{build_dir}}/{{name}}-windows-arm64.exe

    # CGO_ENABLED=1 GOOS=android GOARCH=arm64 go build {{build_flags}} -o {{build_dir}}/{{name}}-android-arm64
    # CGO_ENABLED=1 GOOS=android GOARCH=arm go build {{build_flags}} -o {{build_dir}}/{{name}}-android-arm

    # CGO_ENABLED=1 GOOS=ios GOARCH=arm64 go build {{build_flags}} -buildmode=c-archive -o {{build_dir}}/{{name}}-ios-arm64


    chmod +x {{build_dir}}/{{name}}-linux-*
    chmod +x {{build_dir}}/{{name}}-darwin-*

build-current: tidy setup
    go build {{build_flags}} -o {{build_dir}}/{{name}}
    chmod +x {{build_dir}}/{{name}}

package: build
    #!/usr/bin/env sh
    cd {{build_dir}}

    tar czf {{name}}-linux-amd64.tar.gz {{name}}-linux-amd64
    tar czf {{name}}-linux-arm64.tar.gz {{name}}-linux-arm64

    tar czf {{name}}-darwin-amd64.tar.gz {{name}}-darwin-amd64
    tar czf {{name}}-darwin-arm64.tar.gz {{name}}-darwin-arm64

    zip {{name}}-windows-amd64.zip {{name}}-windows-amd64.exe
    zip {{name}}-windows-arm64.zip {{name}}-windows-arm64.exe

lint:
    go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.0.2
    golangci-lint run --config .golangci.yml