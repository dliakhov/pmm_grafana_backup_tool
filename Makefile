.PHONY : build clean dist fresh run_backup run_restore test docker_build docker_run_backup

BIN := pmm_grafana_backup_tool
GRAFANA_TOKEN ?=
GRAFANA_URL ?=

build:
	go build -o ${BIN} -ldflags="-s -w"

clean:
	go clean
	- rm -f ${BIN}

dist:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(MAKE) build

fresh: clean build run

run_backup:
	./${BIN} backup

run_restore:
	./${BIN} restore

test:
	go test

docker_build:
	docker build . -t aemdeveloper/pmm_grafana_backup_tool_test:0.0.1

docker_run_backup:
	docker run --rm --name pmm_grafana_backup_tool \
                   -e GRAFANA_TOKEN=${GRAFANA_TOKEN} \
                   -e GRAFANA_URL=${GRAFANA_URL} \
                   -v /Users/dmytro.liakhov/dev/percona/pmm_grafana_backup/pmm_grafana_backup_tool/dashboards_docker:/opt/pmm_grafana_backup_tool/_OUTPUT_  \
                   $(docker build -q .)
