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
	go run main.go --config=${CONFIG} --log-dir=${LOG_DIR} --token=gs://gmailagg-oauth/token.json ${GMAILAGG_ARGS}
auth:
	$(MAKE) _execute GMAILAGG_ARGS="auth"
auth_dry_run:
	$(MAKE) _execute GMAILAGG_ARGS="auth --dry-run"
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

PROJECT:=gmailagg
REGION:=us-west1
REPOSITORY_ID:=gmailagg-app
REGISTRY:=${REGION}-docker.pkg.dev
IMAGE_TAG:=latest
IMAGE:=${REGISTRY}/${PROJECT}/${REPOSITORY_ID}/job:${IMAGE_TAG}

setup_terraform_backend:
	gsutil mb -b on -c standard -p ${PROJECT} -l ${REGION} gs://gmailagg-tfstate
setup_docker_auth:
	gcloud auth configure-docker ${REGISTRY}
setup_repository:
	gcloud --project ${PROJECT} artifacts repositories create ${REPOSITORY_ID} \
	  --repository-format=docker \
	  --location=${REGION}
setup_cleanup_policy:
	gcloud artifacts repositories set-cleanup-policies ${REPOSITORY_ID} \
	  --project ${PROJECT} \
	  --location ${REGION} \
	  --policy ./infra/repository_cleanup_policy.json \
	  --no-dry-run \
	  --overwrite

BUILD_DIR:= .local/build
build:
	mkdir -p ${BUILD_DIR}
	cp -f ./infra/start.sh ${BUILD_DIR}/start.sh
	CGO_ENABLED=0 go build -o ${BUILD_DIR}/gmailagg main.go
	docker build -f Dockerfile -t ${IMAGE} ${BUILD_DIR}

push:
	docker push ${IMAGE}
