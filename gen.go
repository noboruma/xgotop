package main

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go -type go_runtime_event_t -target arm64 -output-dir cmd/xgotop ebpf xgotop.bpf.c
