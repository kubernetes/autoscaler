TAG?=dev
REGISTRY?=staging-k8s.gcr.io
COMPONENT=stress
FULL_COMPONENT=vpa-${COMPONENT}

ALL_ARCHITECTURES?=amd64 arm arm64 ppc64le s390x
export DOCKER_CLI_EXPERIMENTAL=enabled

.PHONY: docker-build
docker-build: $(addprefix docker-build-,$(ALL_ARCHITECTURES))

.PHONY: docker-build-*
docker-build-%:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker build --pull -t ${REGISTRY}/${FULL_COMPONENT}-$*:${TAG} --build-arg ARCH=$* .

.PHONY: docker-push
docker-push: $(addprefix sub-push-,$(ALL_ARCHITECTURES)) push-multi-arch;

.PHONY: sub-push-*
sub-push-%: docker-build-% do-push-% ;

.PHONY: do-push-*
do-push-%:
ifndef REGISTRY
	ERR = $(error REGISTRY is undefined)
	$(ERR)
endif
ifndef TAG
	ERR = $(error TAG is undefined)
	$(ERR)
endif
	docker push ${REGISTRY}/${FULL_COMPONENT}-$*:${TAG}

.PHONY: push-multi-arch
push-multi-arch:
	docker manifest create --amend $(REGISTRY)/${FULL_COMPONENT}:$(TAG) $(shell echo $(ALL_ARCHITECTURES) | sed -e "s~[^ ]*~$(REGISTRY)/${FULL_COMPONENT}\-&:$(TAG)~g")
	@for arch in $(ALL_ARCHITECTURES); do docker manifest annotate --arch $${arch} $(REGISTRY)/${FULL_COMPONENT}:$(TAG) $(REGISTRY)/${FULL_COMPONENT}-$${arch}:${TAG}; done
	docker manifest push --purge $(REGISTRY)/${FULL_COMPONENT}:$(TAG)

.PHONY: release
release: docker-build docker-push
	@echo "Full in-docker release ${FULL_COMPONENT}:${TAG} completed"
