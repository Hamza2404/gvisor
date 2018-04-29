// Copyright 2018 Google Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package filter defines all syscalls the sandbox is allowed to make
// to the host, and installs seccomp filters to prevent prohibited
// syscalls in case it's compromised.
package filter

import (
	"fmt"

	"gvisor.googlesource.com/gvisor/pkg/log"
	"gvisor.googlesource.com/gvisor/pkg/seccomp"
	"gvisor.googlesource.com/gvisor/pkg/sentry/platform"
	"gvisor.googlesource.com/gvisor/pkg/sentry/platform/kvm"
	"gvisor.googlesource.com/gvisor/pkg/sentry/platform/ptrace"
)

// Install installs seccomp filters for based on the given platform.
func Install(p platform.Platform, whitelistFS, console, hostNetwork bool) error {
	s := allowedSyscalls

	// Set of additional filters used by -race and -msan. Returns empty
	// when not enabled.
	s = append(s, instrumentationFilters()...)

	if whitelistFS {
		Report("direct file access allows unrestricted file access!")
		s = append(s, whitelistFSFilters()...)
	}
	if console {
		Report("console is enabled: syscall filters less restrictive!")
		s = append(s, consoleFilters()...)
	}
	if hostNetwork {
		Report("host networking enabled: syscall filters less restrictive!")
		s = append(s, hostInetFilters()...)
	}

	switch p := p.(type) {
	case *ptrace.PTrace:
		s = append(s, ptraceFilters()...)
	case *kvm.KVM:
		s = append(s, kvmFilters()...)
	default:
		return fmt.Errorf("unknown platform type %T", p)
	}

	// TODO: Set kill=true when SECCOMP_RET_KILL_PROCESS is supported.
	return seccomp.Install(s, false)
}

// Report writes a warning message to the log.
func Report(msg string) {
	log.Warningf("*** SECCOMP WARNING: %s", msg)
}