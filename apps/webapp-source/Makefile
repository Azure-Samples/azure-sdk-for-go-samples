PACKAGE = github.com/Azure-Samples/azure-sdk-for-go-samples/apps/basic_web_app

REGISTRY = local
IMAGE = go-sample
TAG = latest

PORT = 8080
IMAGE_URI = $(REGISTRY)/$(IMAGE):$(TAG)
C = $(IMAGE)-tester
HOST = localhost:$(PORT)

binary:
	mkdir -p out
	go build -o out/server .
	./out/server &
	curl http://$(HOST)/?name=josh && echo ""
	pkill --euid $(USER) --newest --exact server

container:
	docker build -t $(IMAGE_URI) .
	# docker push $(IMAGE_URI)
	docker run -d --rm \
		--name $C \
		--publish "$(PORT):8080" \
		$(IMAGE_URI)
	curl "http://$(HOST)/?name=josh" && echo ""
	docker container logs $C
	docker container stop $C
	echo "start a new container with"
	echo "    \`docker run [-d|-it] -p $(PORT):8080 $(IMAGE_URI)\`"

.PHONY: binary container
