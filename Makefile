build:
	go mod tidy
	go build -o ampp

CPATH:="-Xlinker -rpath=/opt/clib-"
CPATH2:=$(CPATH)$(CVERS)

lab:
ifeq ($(CVERS),)
	@echo "it's empty, do not export"
else
	@echo "it's not empty, export"
	@echo $(CPATH2)
	set(ENV{CGO_LDFLAGS}$(CPATH2))
	@echo $(CGO_LDFLAGS)
endif
