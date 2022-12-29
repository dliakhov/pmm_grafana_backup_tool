FROM golang:alpine as builder

WORKDIR /build

COPY . .

RUN apk add --no-cache make
RUN make dist

FROM alpine:3.17

ENV RESTORE false

WORKDIR /

COPY --from=builder /build/pmm_grafana_backup_tool .

#ENTRYPOINT ["./pmm_grafana_backup_tool", "backup"]
CMD sh -c 'if [ "$RESTORE" = true ]; then ./pmm_grafana_backup_tool restore; else ./pmm_grafana_backup_tool backup; fi'
