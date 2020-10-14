debugDeps = src

.PHONY: stop run-debug ${debugDeps}

run-debug: docker-compose.yml docker-compose.debug.yml ${debugDeps}
	docker-compose -f docker-compose.yml -f docker-compose.debug.yml up -d

stop:
	docker-compose -f docker-compose.yml -f docker-compose.debug.yml down

$(debugDeps):
	$(MAKE) -C $@ debug-image

