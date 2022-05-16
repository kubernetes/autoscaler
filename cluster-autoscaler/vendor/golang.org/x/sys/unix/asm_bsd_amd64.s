// Copyright 2021 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

<<<<<<<< HEAD:cluster-autoscaler/vendor/golang.org/x/sys/unix/asm_bsd_amd64.s
//go:build (darwin || dragonfly || freebsd || netbsd || openbsd) && gc
// +build darwin dragonfly freebsd netbsd openbsd
========
//go:build (darwin || freebsd || netbsd || openbsd) && gc
// +build darwin freebsd netbsd openbsd
>>>>>>>> cluster-autoscaler-release-1.22:cluster-autoscaler/vendor/golang.org/x/sys/unix/asm_bsd_arm64.s
// +build gc

#include "textflag.h"

<<<<<<<< HEAD:cluster-autoscaler/vendor/golang.org/x/sys/unix/asm_bsd_amd64.s
// System call support for AMD64 BSD
========
// System call support for ARM64 BSD
>>>>>>>> cluster-autoscaler-release-1.22:cluster-autoscaler/vendor/golang.org/x/sys/unix/asm_bsd_arm64.s

// Just jump to package syscall's implementation for all these functions.
// The runtime may know about them.

TEXT	·Syscall(SB),NOSPLIT,$0-56
	JMP	syscall·Syscall(SB)

TEXT	·Syscall6(SB),NOSPLIT,$0-80
	JMP	syscall·Syscall6(SB)

TEXT	·Syscall9(SB),NOSPLIT,$0-104
	JMP	syscall·Syscall9(SB)

TEXT	·RawSyscall(SB),NOSPLIT,$0-56
	JMP	syscall·RawSyscall(SB)

TEXT	·RawSyscall6(SB),NOSPLIT,$0-80
	JMP	syscall·RawSyscall6(SB)
