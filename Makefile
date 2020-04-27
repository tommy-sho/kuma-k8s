

PROJECT  := $(shell gcloud projects list| sed -n 2P |cut -d' ' -f1)
ZONE     := asia-northeast1-c
CLUSTER  := kuma-k8s
SERVICES := gateway backend backend2
name     := hoge

.PHONY: build push apply delete proto

proto:
	protoc  \
	--go_out=plugins=grpc:. \
    ./proto/*.proto

build:
	for service in ${SERVICES}; do \
		docker image build -t kuma_k8s/$$service:latest app/$$service/; \
	done


apply:
	kubectl apply -f manifest/

delete:
	kubectl delete -f manifest/


run:
	@cd pkg/client && \
	go run main.go -name=${name}


gke-init:
	gcloud container clusters create ${CLUSTER} --zone=${ZONE} --num-nodes=3 --preemptible --disk-size=10
	gcloud container clusters get-credentials ${CLUSTER} --zone=${ZONE}

gke-yaml:
	for service in ${SERVICES}; do \
		   sed -e 's/kuma_k8s/gcr.io\/${PROJECT}/g' ./manifest/$$service.yaml > ./manifest/$$service-gke.yaml; \
	done

image-push:
	for service in ${SERVICES}; do \
		docker tag kuma_k8s/$$service:latest gcr.io/${PROJECT}/$$service:latest; \
		docker push gcr.io/${PROJECT}/$$service:latest;\
	done