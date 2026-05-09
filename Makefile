
.PHONY: all deploy build scale-up scale-down status logs clean test-load help init stop update

STACK_NAME := demo
APP_NAME := app
SCALER_PORT := 3619

all: build deploy

init:
	docker info | grep -q "Swarm: active" || docker swarm init

build:
	docker build --pull --no-cache -t demo-app:latest ./app
	docker build --pull --no-cache -t scaler:latest ./scaler
	docker pull gcr.io/cadvisor/cadvisor:v0.47.0

deploy: init
	docker stack deploy -c compose.yml $(STACK_NAME)
	docker service update --force $(STACK_NAME)_$(APP_NAME) || true

scale-up:
	@echo "Trigger scale-up webhook"
	@curl -4 -s http://localhost:$(SCALER_PORT)/scale-up || echo "Fail to connect to the target"

scale-down:
	@echo "Trigger scale-down webhook"
	@curl -4 -s http://localhost:$(SCALER_PORT)/scale-down || echo "Fail to connect to the target"

status:
	docker stack services $(STACK_NAME)

logs:
	docker service logs -f $(STACK_NAME)_$(APP_NAME) || echo "Service not running"

stop:
	docker stack services $(STACK_NAME) -q | xargs -r -I{} docker service scale {}=0 || true
	docker swarm leave --force || true

clean:
	docker stack rm $(STACK_NAME)
	sleep 5
	docker ps -a --filter "label=com.docker.stack.namespace=$(STACK_NAME)" -q | xargs -r docker rm -f

test-load:
	k6 run load-test.js

help:
	@echo "Targets:"
	@echo "  all        - build and deploy"
	@echo "  init       - initialize docker swarm"
	@echo "  build      - build app and scaler images"
	@echo "  deploy     - deploy stack"
	@echo "  scale-up   - trigger scale up via scaler"
	@echo "  scale-down - trigger scale down via scaler"
	@echo "  status     - show stack services"
	@echo "  logs       - tail app logs"
	@echo "  clean      - remove stack and containers"
	@echo "  stop       - stop stack and leave swarm (no delete)"
	@echo "  test-load  - run load test with ab"


update:
	docker service update --force $(filter-out $@,$(MAKECMDGOALS))