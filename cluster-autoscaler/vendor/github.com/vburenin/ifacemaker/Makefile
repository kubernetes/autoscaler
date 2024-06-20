include vars.mk

build:
	docker build . --build-arg IFACEMAKER_VERSION=$(IFACEMAKER_VERSION) -t $(IMAGE_NAME):$(IFACEMAKER_VERSION)
