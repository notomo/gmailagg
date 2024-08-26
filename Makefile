TEST_OPTS:=
test:
	rm -rf /tmp/gmailaggtest
	go test ${TEST_OPTS} -v ./...

lint:
	go vet ./...

CONFIG:=example_config.json
LOG_DIR:=/tmp/gmailagg
_execute:
	rm -rf ${LOG_DIR}
	go run main.go --config=${CONFIG} --log-dir=${LOG_DIR} ${GMAILAGG_ARGS}
auth:
	$(MAKE) _execute GMAILAGG_ARGS="auth"
auth_dry_run:
	$(MAKE) _execute GMAILAGG_ARGS="auth --dry-run"
run:
	$(MAKE) _execute GMAILAGG_ARGS="run"
dry_run:
	$(MAKE) _execute GMAILAGG_ARGS="run --dry-run"
