
PYPKGNAME=pyslicer
GOROOT=$(which go)
GOPATH=$(HOME)/go
GOBIN=$(GOPATH)/bin
PATH:=$(PATH):$(GOROOT):$(GOPATH):$(GOBIN)
MAKEFILE_DIR:=$(shell dirname $(realpath $(firstword $(MAKEFILE_LIST))))
SLICER_ROOT=$(MAKEFILE_DIR)
# Generate the python client too
PY_OUT_DIR=$(SLICER_ROOT)/protos/$(PYPKGNAME)


all: protos

protos: printenv goprotos pyprotos

test:
	cd $(SLICER_ROOT)/ && go test ./... -cover

goprotos:
	echo "Generating GO bindings"
	protoc --go_out=$(SLICER_ROOT) --go_opt=paths=source_relative          \
       --go-grpc_out=$(SLICER_ROOT) --go-grpc_opt=paths=source_relative	\
       --proto_path=$(SLICER_ROOT)                     			\
      $(SLICER_ROOT)/protos/control.proto
	protoc --go_out=$(SLICER_ROOT)/cmd/echosvc/ --go_opt=paths=source_relative          \
       --go-grpc_out=$(SLICER_ROOT)/cmd/echosvc/ --go-grpc_opt=paths=source_relative	\
       --proto_path=$(SLICER_ROOT)/cmd/echosvc/                     			\
      $(SLICER_ROOT)/cmd/echosvc/echosvc.proto

pyprotos:
	echo "Generating Python bindings"
	mkdir -p $(PY_OUT_DIR) $(SLICER_ROOT)/$(PYPKGNAME)
	python3 -m grpc_tools.protoc -I./protos   \
      --python_out="$(PY_OUT_DIR)"          \
      --grpc_python_out="$(PY_OUT_DIR)"     \
      --proto_path=$(SLICER_ROOT)           \
      $(SLICER_ROOT)/protos/control.proto
	@mv $(PY_OUT_DIR)/protos/*.py $(SLICER_ROOT)/$(PYPKGNAME)
	@echo "Cleaning up files..."
	rm -Rf $(PY_OUT_DIR)

printenv:
	@echo MAKEFILE_DIR=$(MAKEFILE_DIR)
	@echo SLICER_ROOT=$(SLICER_ROOT)
	@echo MAKEFILE_LIST=$(MAKEFILE_LIST)
	@echo SLICER_ROOT=$(SLICER_ROOT)
	@echo GOROOT=$(GOROOT)
	@echo GOPATH=$(GOPATH)
	@echo GOBIN=$(GOBIN)
	@echo PYPKGNAME=$(PYPKGNAME)
	@echo PY_OUT_DIR=$(PY_OUT_DIR)

## Setting up the dev db
pgdb:
	docker build -t slicerpgdb -f Dockerfile.pgdb .

runtestdb:
	mkdir -p $(MAKEFILE_DIR)/pgdata_test
	docker run --rm --name slicer-pgdb-container -v ${MAKEFILE_DIR}/pgdata_test:/var/lib/postgresql/data -e POSTGRES_PASSWORD=password -p 5432:5432 slicerpgdb

rundb:
	mkdir -p $(MAKEFILE_DIR)/pgdata
	docker run --rm --name slicer-pgdb-container -p 5432:5432 -v $(MAKEFILE_DIR)/pgdata:/var/lib/postgresql/data slicerpgdb

