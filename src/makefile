.PHONY: clean debug prod debug-image prod-image

gocmd = CGO_ENABLED=0 go
dockercmd = docker build --build-arg 

debug-image: debug
	@$(dockercmd) "name=debug_bin" -t jumpdir:debug .

prod-image: prod
	@$(dockercmd) "name=prod_bin" -t jumpdir:prod .

debug:
	$(gocmd) build -tags debug -o debug_bin
prod:
	$(gocmd) build -o prod_bin

clean:
	rm -f debug prod
