# Copyright 2018 The OpenPitrix Authors. All rights reserved.
# Use of this source code is governed by a Apache license
# that can be found in the LICENSE file.

TARG:=libconfd-simple-app

GO_FILES=$(sort $(wildcard ./*.go))

KILL_APP:=ps aux | grep $(TARG) | grep -v grep | awk '{print $$2}' | xargs kill -9
SHOW_APP:=ps aux | grep $(TARG) | grep -v grep

.PHONY: help
help:
	@echo "Please use \`make <target>\` where <target> is one of"
	@echo "  check-config to check app config"
	@echo "  check-server to check app server"
	@echo "  start        to start app server"
	@echo "  restart      to restart app server"
	@echo "  stop         to stop app server"
	@echo "  clean        to stop app server and clean the temp files"
	@echo "  help         to show this help informatrion"


.PHONY: check-config
check-config: $(TARG).out
	./$(TARG).out -check-config
	@echo "ok"

.PHONY: check-server
check-server: $(TARG).out
	./$(TARG).out -check-server
	@echo "ok"

.PHONY: start
start: $(TARG).out
	nohup ./$(TARG).out > _app.log 2>&1 &

.PHONY: restart
restart: $(TARG).out stop
	nohup ./$(TARG).out > _app.log 2>&1 &

.PHONY: stop
stop:
	-$(KILL_APP)

.PHONY: ps
ps:
	-$(SHOW_APP)

.PHONY: clean
clean: stop
	-rm $(TARG).out
	-rm _app.log


$(TARG).out: $(GO_FILES)
	go fmt ./...
	go vet ./...
	go build -o $@
