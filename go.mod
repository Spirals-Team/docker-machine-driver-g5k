module github.com/Spirals-Team/docker-machine-driver-g5k

go 1.25.0

require (
	github.com/docker/machine v0.16.2
	github.com/go-resty/resty/v2 v2.16.2
	golang.org/x/crypto v0.51.0
)

require (
	golang.org/x/net v0.55.0 // indirect
	golang.org/x/sys v0.45.0 // indirect
	golang.org/x/term v0.43.0 // indirect
)

replace github.com/docker/machine v0.16.2 => gitlab.com/gitlab-org/ci-cd/docker-machine v0.16.2-gitlab.30
