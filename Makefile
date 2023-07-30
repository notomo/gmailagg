TEST_OPTS:=
test:
	rm -rf /tmp/gmailaggtest
	go test ${TEST_OPTS} -v ./...

lint:
	go vet ./...

CONFIG:=gmailagg.yaml
LOG_DIR:=/tmp/gmailagg
_execute:
	rm -rf ${LOG_DIR}
	go run main.go --config=${CONFIG} --log-dir=${LOG_DIR} ${GMAILAGG_ARGS}
auth:
	$(MAKE) _execute GMAILAGG_ARGS="auth"
run:
	$(MAKE) _execute GMAILAGG_ARGS="run"
dry_run:
	$(MAKE) _execute GMAILAGG_ARGS="run --dry-run"

INFLUXDB_ORG:=example-org
INFLUXDB_USER:=writer
INFLUXDB_BUCKET:=gmailagg
influxdb_setup_auth:
	$(MAKE) influxdb_user
	$(MAKE) influxdb_token
influxdb_user:
	docker compose exec influxdb influx user create --name="${INFLUXDB_USER}" --password="${INFLUXDB_PASSWORD}" --org="${INFLUXDB_ORG}"
influxdb_token:
	docker compose exec influxdb influx auth create --org="${INFLUXDB_ORG}" --user="${INFLUXDB_USER}" --write-buckets
influxdb_token_list:
	docker compose exec influxdb influx auth list --org="${INFLUXDB_ORG}"
influxdb_clear_bucket:
	docker compose exec influxdb influx delete --org="${INFLUXDB_ORG}" --bucket="${INFLUXDB_BUCKET}" --start=2009-01-02T23:00:00Z --stop=2099-01-02T23:00:00Z
