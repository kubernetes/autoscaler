export BUILD_TAGS=brightbox
export REGISTRY?=brightbox
export GOARCH?=$(shell go env GOARCH)
ifndef TAG
	override TAG=dev
endif
export TAG
ifeq ($(TAG), dev)
	deploydeps=build
endif

.PHONY: deploy
deploy: examples/config.rb $(deploydeps)
	helm repo update
	ruby $< | \
	helm template release autoscaler/cluster-autoscaler \
	--namespace kube-system -f - | \
	kubectl -n kube-system apply -f -

.PHONY: remove
remove: examples/config.rb
	helm repo update
	ruby $< | \
	helm template release autoscaler/cluster-autoscaler \
	--namespace kube-system -f - | \
	kubectl -n kube-system delete -f -

.PHONY: secret
secret: ${HOME}/.docker/config.json
	-kubectl create secret generic regcred \
	    --from-file=.dockerconfigjson=$? \
	    --type=kubernetes.io/dockerconfigjson

../../cluster-autoscaler: brightbox_cloud_provider.go brightbox_node_group.go
	$(MAKE) -C $(@D) container
	docker tag ${REGISTRY}/cluster-autoscaler-${BUILD_TAGS}-${GOARCH}:${TAG} ${REGISTRY}/cluster-autoscaler-${BUILD_TAGS}:${TAG}
	docker push ${REGISTRY}/cluster-autoscaler-${BUILD_TAGS}:${TAG}

.PHONY: build
build: ../../cluster-autoscaler

.PHONY: clean
clean: 
	$(MAKE) -C ../.. $@
