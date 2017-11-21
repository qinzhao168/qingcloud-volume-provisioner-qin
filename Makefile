# tab space is 4
# GitHub viewer defaults to 8, change with ?ts=4 in URL

# Vars describing project
NAME= qingcloud-volume-provisioner
GIT_REPOSITORY= github.com/yunify/qingcloud-volume-provisioner
DOCKER_IMAGE_NAME?= qingcloud/qingcloud-volume-provisioner
GOPATH ?= $(HOME)/go/
BASE= $(GOPATH)/src/$(GIT_REPOSITORY)
BIN      = $(GOPATH)/bin

export GOPATH
GO      = go
GOFMT   = gofmt
TIMEOUT = 15
V = 0
Q = $(if $(filter 1,$V),,@)
M = $(shell printf "\033[34;1m�\033[0m")
# Generate vars to be included from external script
# Allows using bash to generate complex vars, such as project versions
GENERATE_VERSION_INFO_SCRIPT	= ./generate_version.sh
GENERATE_VERSION_INFO_OUTPUT	= version_info

# Define newline needed for subsitution to allow evaluating multiline script output
define newline


endef

# Call the version_info script with keyvalue option and evaluate the output
# Will import the keyvalue pairs and make available as Makefile variables
# Use dummy variable to only have execute once
$(eval $(subst #,$(newline),$(shell $(GENERATE_VERSION_INFO_SCRIPT) keyvalue | tr '\n' '#')))
# Call the verson_info script with json option and store result into output file and variable
# Will only execute once due to ':='
#GENERATE_VERSION_INFO			:= $(shell $(GENERATE_VERSION_INFO_SCRIPT) json | tee $(GENERATE_VERSION_INFO_OUTPUT))

# Set defaults for needed vars in case version_info script did not set
# Revision set to number of commits ahead
VERSION							?= 0.0
COMMITS							?= 0
REVISION						?= $(COMMITS)
BUILD_LABEL						?= unknown_build
BUILD_DATE						?= $(shell date -u +%Y%m%d.%H%M%S)
GIT_SHA1						?= unknown_sha1

IMAGE_LABLE         ?= $(BUILD_LABEL)
# Vars for export ; generate list of ENV vars based on matching export prefix
# Use strip to get rid of excessive spaces due to the foreach / filter / if logic
EXPORT_VAR_PREFIX               = EXPORT_VAR_
EXPORT_VARS                     = $(strip $(foreach v,$(filter $(EXPORT_VAR_PREFIX)%,$(.VARIABLES)),$(if $(filter environment%,$(origin $(v))),$(v))))

# Vars for go phase
# All vars which being with prefix will be included in ldflags
# Defaulting to full static build
GO_VARIABLE_PREFIX				= GO_VAR_
GO_VAR_BUILD_LABEL				:= $(BUILD_LABEL)
GO_VAR_VERSION                  := $(VERSION)
GO_VAR_GIT_SHA1                 := $(GIT_SHA1)
GO_VAR_BUILD_LABEL              := $(BUILD_LABEL)
GO_LDFLAGS						= $(foreach v,$(filter $(GO_VARIABLE_PREFIX)%, $(.VARIABLES)),-X github.com/yunify/qingcloud-volume-provisioner/qingcloud.$(patsubst $(GO_VARIABLE_PREFIX)%,%,$(v))=$(value $(value v)))
GO_BUILD_FLAGS					= -a -tags netgo -installsuffix nocgo -ldflags "$(GO_LDFLAGS)"

#src
qingcloud-volume-provisioner_pkg = $(subst $(GIT_REPOSITORY)/,,$(shell go list -f '{{ join .Deps "\n" }}' $(GIT_REPOSITORY)/cmd/qingcloud-volume-provisioner | grep "^$(GIT_REPOSITORY)" |grep -v "^$(GIT_REPOSITORY)/vendor/" ))
qingcloud-volume-provisioner_pkg += cmd/qingcloud-volume-provisioner
qingcloud-flex-volume_pkg = $(subst $(GIT_REPOSITORY)/,,$(shell go list -f '{{ join .Deps "\n" }}' $(GIT_REPOSITORY)/cmd/qingcloud-flex-volume | grep "^$(GIT_REPOSITORY)" |grep -v "^$(GIT_REPOSITORY)/vendor/" ))
qingcloud-flex-volume_pkg += cmd/qingcloud-flex-volume
#TEST_PACKAGES = $(shell go list -f '{{ join .Deps "\n" }}' $(GIT_REPOSITORY) | grep "^$(GIT_REPOSITORY)" |grep -v "^$(GIT_REPOSITORY)/vendor/" )
#TEST_PACKAGES += $(GIT_REPOSITORY)

# default just build binary
default							: go-build
# Define targets
TESTPKGS = $(shell env GOPATH=$(GOPATH) $(GO) list -f '{{ if or .TestGoFiles .XTestGoFiles }}{{ .ImportPath }}{{ end }}' $(PKGS))
# default just build binary
all : $(BASE) go-build ; $(info $(M) executable is built ...)
# target for debugging / printing variables
print-%							:
								@echo '$*=$($*)'

# perform go build on project
go-build						: bin/qingcloud-volume-provisioner bin/qingcloud-flex-volume

bin/qingcloud-volume-provisioner                     : $(foreach dir,$(qingcloud-volume-provisioner_pkg),$(wildcard $(dir)/*.go)) Makefile
								mkdir -p bin
								go build -o bin/qingcloud-volume-provisioner $(GO_BUILD_FLAGS) $(GIT_REPOSITORY)/cmd/qingcloud-volume-provisioner

bin/qingcloud-flex-volume                            : $(foreach dir,$(qingcloud-flex-volume_pkg),$(wildcard $(dir)/*.go)) Makefile
								mkdir -p bin
								go build -o bin/qingcloud-flex-volume $(GO_BUILD_FLAGS) $(GIT_REPOSITORY)/cmd/qingcloud-flex-volume

bin/qingcloud-flex-volume.tar.gz              : bin/qingcloud-flex-volume
								tar -C bin/ -czf bin/qingcloud-flex-volume.tar.gz qingcloud-flex-volume

bin/.docker-images-build-timestamp                         : bin/qingcloud-volume-provisioner Makefile Dockerfile
								docker build -q -t $(DOCKER_IMAGE_NAME):$(IMAGE_LABLE) -t dockerhub.qingcloud.com/$(DOCKER_IMAGE_NAME):$(IMAGE_LABLE) . > bin/.docker-images-build-timestamp

release : bin/qingcloud-flex-volume.tar.gz bin/.docker-images-build-timestamp

bin/.docker_label               : bin/.docker-images-build-timestamp
								docker push $(DOCKER_IMAGE_NAME):$(IMAGE_LABLE)
								docker push dockerhub.qingcloud.com/$(DOCKER_IMAGE_NAME):$(IMAGE_LABLE)
								echo $(DOCKER_IMAGE_NAME):$(IMAGE_LABLE) > bin/.docker_label

install-docker                  : bin/.docker_label

publish                         : bin/.docker_label
								docker tag `cat bin/.docker_label` $(DOCKER_IMAGE_NAME):latest
								docker tag `cat bin/.docker_label` dockerhub.qingcloud.com/$(DOCKER_IMAGE_NAME):latest
								docker push $(DOCKER_IMAGE_NAME):latest
								docker push dockerhub.qingcloud.com/$(DOCKER_IMAGE_NAME):latest

clean                           :
								rm -rf bin/ && if -f bin/.docker-images-build-timestamp then docker rmi `cat bin/.docker-images-build-timestamp`


.PHONY							: default all go-build clean release install-docker
