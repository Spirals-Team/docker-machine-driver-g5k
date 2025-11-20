module github.com/Spirals-Team/docker-machine-driver-g5k

go 1.24.0

toolchain go1.24.1

require (
	github.com/docker/machine v0.16.2
	github.com/go-resty/resty/v2 v2.16.2
	golang.org/x/crypto v0.45.0
)

require (
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sys v0.38.0 // indirect
	golang.org/x/term v0.37.0 // indirect
)

replace github.com/docker/machine v0.16.2 => gitlab.com/gitlab-org/ci-cd/docker-machine v0.16.2-gitlab.30
