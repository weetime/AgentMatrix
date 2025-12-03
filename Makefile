DOCKER_IMAGE=quanzhenglong.com/camp/nova

CMD_PATH=./cmd/
BUILD_PATH=./build/
DIRS = $(shell ls $(CMD_PATH))

VERSION=$(shell git describe --tags --always)

DOCKERFILE=./Dockerfile

ARCH=amd64

.PHONY: go_get
go_get: generate_mod
	go install github.com/google/wire/cmd/wire@latest
	go install entgo.io/ent/cmd/ent@latest
	go install github.com/go-kratos/kratos/cmd/kratos/v2@latest && kratos upgrade
	go install github.com/envoyproxy/protoc-gen-validate@latest

.PHONY: generate
# generate mod, proto, ent, cmd
generate: generate_proto generate_ent generate_wire

.PHONY: generate_mod
generate_mod:
	go mod tidy

.PHONY: generate_proto
generate_proto:
	kratos proto client .

.PHONY: generate_ent
generate_ent:
	CGO_ENABLED=0 ent generate --feature sql/modifier --feature sql/upsert --feature sql/lock --feature intercept,schema/snapshot ./internal/data/ent/schema

.PHONY: generate_wire
generate_wire: generate_mod
	@ for dir in $(DIRS); \
	do \
  if [ "$$dir" != "$(CMD_PATH)" ]; then \
	cd ${CMD_PATH}$$dir/ && wire . && cd ../../;\
  fi \
	done

.PHONY: all
all: image

.PHONY: all_local
all_local: linux image

.PHONY: local
local:
	@ for dir in $(DIRS); \
	do \
  if [ "$$dir" != "$(CMD_PATH)" ]; then \
	go build -o $(BUILD_PATH)$$dir ${CMD_PATH}$$dir; \
	echo "build "$$dir; \
  fi \
	done

.PHONY: clean
clean:
	rm -rf $(BUILD_PATH)

.PHONY: test
test:
	go test ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: image
image:
	docker buildx build --platform linux/${ARCH} \
	  --build-arg ARCH=${ARCH} \
	  --file ${DOCKERFILE} \
	  -t ${DOCKER_IMAGE}:${VERSION}-${ARCH} \
	  --push .

.PHONY: linux
linux:
	@ for dir in $(DIRS); \
	do \
  if [ "$$dir" != "$(CMD_PATH)" ]; then \
	GOARCH=${ARCH} GOOS=linux CGO_ENABLED=0 go build -o $(BUILD_PATH)$$dir ${CMD_PATH}$$dir; \
	echo "build "$$dir; \
  fi \
	done

.PHONY: init_workspace
init_workspace:
	sed -e 's/anyone/$(shell hostname)/g' .nocalhost/.env | xargs -I {} echo "{}" > .nocalhost/.env_lock

.PHONY: check_init
check_init:
	@$(shell if [ ! -f .nocalhost/.env_lock ];then echo "请先在本地执行 make init_workplace";fi;)

.PHONY: run_debug
run_debug:
	kratos run

.PHONY: save_image
save_image:
	docker save -o $(BUILD_PATH)nova_release_${VERSION}.tar ${DOCKER_IMAGE}:${VERSION}-${ARCH}; \
	gzip -f $(BUILD_PATH)nova_release_${VERSION}_${ARCH}.tar; \

.PHONY: release
release:generate linux image save_image

.PHONY: build
build: go_get generate linux
