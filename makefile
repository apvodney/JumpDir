debugDeps = src
composeBase = -f docker-compose.yml
composeDebug = $(composeBase) -f docker-compose.debug.yml

.PHONY: down run-debug $(debugDeps)

run-debug: down docker-compose.yml docker-compose.debug.yml $(debugDeps)
	docker-compose $(composeDebug) up -d

down:
	docker-compose $(composeDebug) down

$(debugDeps):
	$(MAKE) -C $@ debug-image

