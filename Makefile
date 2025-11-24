all: gen compile run-silent
.PHONY: all 

vmlinux:
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > vmlinux.h

gen: vmlinux
	go generate

compile: gen
	go build -o testserver ./cmd/testserver
	go build -o goroutinepadding ./cmd/goroutinepadding
	go build -o xgotop ./cmd/xgotop

samplingtest: gen compile
	./scripts/test_sampling.sh

weboverheadtest: gen compile
	./scripts/test_web_overhead.sh -r "1 2 4 8 16" -p "1" --only-web --flood -n 20000

run:
	sudo ./xgotop -b ./testserver -rw 1 -pw 1

run-silent:
	sudo ./xgotop -b ./testserver -rw 1 -pw 1 -s

# TODO: This date call is not working in the Makefile, fix it.
# runlog:
# 	sudo ./xgotop -b ./testserver -rw 3 -pw 5 2>&1 | tee "xgotop_$(date +%Y-%m-%d-%H-%M-%S).log"

# Web targets
web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

xgotop-web:
	go build -o xgotop ./cmd/xgotop

run-web:
	sudo ./xgotop -b ./testserver -rw 1 -pw 1 -web -web-port 8080 -storage-format binary -storage-dir ./sessions

clean:
	- rm ebpf_arm64*.go
	- rm ebpf_arm64*.o
	- rm testserver
	- rm goroutinepadding
	- rm xgotop
	- rm -rf web/dist
	- rm -rf sessions