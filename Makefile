	.PHONY : build run fresh test clean

BIN := pmm_grafana_backup_tool

build:
	go build -o ${BIN} -ldflags="-s -w"

clean:
	go clean
	- rm -f ${BIN}

dist:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(MAKE) build

fresh: clean build run

lint:
	gofmt -s -w .
	find . -name "*.go" -exec ${GOPATH}/bin/golint {} \;

run_backup:
	./${BIN} backup

run_restore:
	./${BIN} backup

test: lint
	go test

docker_build:
	docker build . -t aemdeveloper/pmm_grafana_backup_tool_test:0.0.1

docker_run_backup:
	docker run --rm --name pmm_grafana_backup_tool \
                   -e GRAFANA_TOKEN=eyJrIjoiS043WW9ubzg4Nnl5TkJlcE5jZFZDek8xUnBacWhmeTciLCJuIjoiYWRtaW4iLCJpZCI6MX0= \
                   -e GRAFANA_URL=http://host.docker.internal/graph/api/ \
                   -v /Users/dmytro.liakhov/dev/percona/pmm_grafana_backup/pmm_grafana_backup_tool/dashboards_docker:/opt/pmm_grafana_backup_tool/_OUTPUT_  \
                   $(docker build -q .)
