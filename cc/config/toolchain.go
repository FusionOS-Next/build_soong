// Copyright 2015 Google Inc. All rights reserved.
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

package config

import (
	"fmt"
	"strings"

	"android/soong/android"
)

func init() {
	pctx.StaticVariable("DarwinAvailableLibraries", strings.Join(darwinAvailableLibraries, " "))
	pctx.StaticVariable("LinuxAvailableLibraries", strings.Join(linuxAvailableLibraries, " "))
	pctx.StaticVariable("WindowsAvailableLibraries", strings.Join(windowsAvailableLibraries, " "))
}

type toolchainFactory func(arch android.Arch) Toolchain

var toolchainFactories = make(map[android.OsType]map[android.ArchType]toolchainFactory)

func registerToolchainFactory(os android.OsType, arch android.ArchType, factory toolchainFactory) {
	if toolchainFactories[os] == nil {
		toolchainFactories[os] = make(map[android.ArchType]toolchainFactory)
	}
	toolchainFactories[os][arch] = factory
}

type toolchainContext interface {
	Os() android.OsType
	Arch() android.Arch
}

func FindToolchainWithContext(ctx toolchainContext) Toolchain {
	t, err := findToolchain(ctx.Os(), ctx.Arch())
	if err != nil {
		panic(err)
	}
	return t
}

func FindToolchain(os android.OsType, arch android.Arch) Toolchain {
	t, err := findToolchain(os, arch)
	if err != nil {
		panic(err)
	}
	return t
}

func findToolchain(os android.OsType, arch android.Arch) (Toolchain, error) {
	factory := toolchainFactories[os][arch.ArchType]
	if factory == nil {
		return nil, fmt.Errorf("Toolchain not found for %s arch %q", os.String(), arch.String())
	}
	return factory(arch), nil
}

type Toolchain interface {
	Name() string

	IncludeFlags() string

	ClangTriple() string
	ToolchainCflags() string
	ToolchainLdflags() string
	Asflags() string
	Cflags() string
	Cppflags() string
	Ldflags() string
	Lldflags() string
	InstructionSetFlags(string) (string, error)

	ndkTriple() string

	YasmFlags() string

	Is64Bit() bool

	ShlibSuffix() string
	ExecutableSuffix() string

	LibclangRuntimeLibraryArch() string

	AvailableLibraries() []string

	CrtBeginStaticBinary() []string
	CrtBeginSharedBinary() []string
	CrtBeginSharedLibrary() []string
	CrtEndStaticBinary() []string
	CrtEndSharedBinary() []string
	CrtEndSharedLibrary() []string
	CrtPadSegmentSharedLibrary() []string

	// DefaultSharedLibraries returns the list of shared libraries that will be added to all
	// targets unless they explicitly specify system_shared_libs.
	DefaultSharedLibraries() []string

	Bionic() bool
	Glibc() bool
	Musl() bool
}

type toolchainBase struct {
}

func (t *toolchainBase) ndkTriple() string {
	return ""
}

func NDKTriple(toolchain Toolchain) string {
	triple := toolchain.ndkTriple()
	if triple == "" {
		// Use the clang triple if there is no explicit NDK triple
		triple = toolchain.ClangTriple()
	}
	return triple
}

func (toolchainBase) InstructionSetFlags(s string) (string, error) {
	if s != "" {
		return "", fmt.Errorf("instruction_set: %s is not a supported instruction set", s)
	}
	return "", nil
}

func (toolchainBase) ToolchainCflags() string {
	return ""
}

func (toolchainBase) ToolchainLdflags() string {
	return ""
}

func (toolchainBase) Asflags() string {
	return ""
}

func (toolchainBase) YasmFlags() string {
	return ""
}

func (toolchainBase) LibclangRuntimeLibraryArch() string {
	return ""
}

type toolchainNoCrt struct{}

func (toolchainNoCrt) CrtBeginStaticBinary() []string       { return nil }
func (toolchainNoCrt) CrtBeginSharedBinary() []string       { return nil }
func (toolchainNoCrt) CrtBeginSharedLibrary() []string      { return nil }
func (toolchainNoCrt) CrtEndStaticBinary() []string         { return nil }
func (toolchainNoCrt) CrtEndSharedBinary() []string         { return nil }
func (toolchainNoCrt) CrtEndSharedLibrary() []string        { return nil }
func (toolchainNoCrt) CrtPadSegmentSharedLibrary() []string { return nil }

func (toolchainBase) DefaultSharedLibraries() []string {
	return nil
}

func (toolchainBase) Bionic() bool {
	return false
}

func (toolchainBase) Glibc() bool {
	return false
}

func (toolchainBase) Musl() bool {
	return false
}

type toolchain64Bit struct {
}

func (toolchain64Bit) Is64Bit() bool {
	return true
}

type toolchain32Bit struct {
}

func (toolchain32Bit) Is64Bit() bool {
	return false
}

func variantOrDefault(variants map[string]string, choice string) string {
	if ret, ok := variants[choice]; ok {
		return ret
	}
	return variants[""]
}

func addPrefix(list []string, prefix string) []string {
	for i := range list {
		list[i] = prefix + list[i]
	}
	return list
}

func LibclangRuntimeLibrary(library string) string {
	return "libclang_rt." + library
}

func BuiltinsRuntimeLibrary() string {
	return LibclangRuntimeLibrary("builtins")
}

func AddressSanitizerRuntimeLibrary() string {
	return LibclangRuntimeLibrary("asan")
}

func AddressSanitizerStaticRuntimeLibrary() string {
	return LibclangRuntimeLibrary("asan.static")
}

func AddressSanitizerCXXStaticRuntimeLibrary() string {
	return LibclangRuntimeLibrary("asan_cxx.static")
}

func HWAddressSanitizerRuntimeLibrary() string {
	return LibclangRuntimeLibrary("hwasan")
}

func HWAddressSanitizerStaticLibrary() string {
	return LibclangRuntimeLibrary("hwasan_static")
}

func UndefinedBehaviorSanitizerRuntimeLibrary() string {
	return LibclangRuntimeLibrary("ubsan_standalone")
}

func UndefinedBehaviorSanitizerMinimalRuntimeLibrary() string {
	return LibclangRuntimeLibrary("ubsan_minimal")
}

func ThreadSanitizerRuntimeLibrary() string {
	return LibclangRuntimeLibrary("tsan")
}

func ScudoRuntimeLibrary() string {
	return LibclangRuntimeLibrary("scudo")
}

func ScudoMinimalRuntimeLibrary() string {
	return LibclangRuntimeLibrary("scudo_minimal")
}

func LibFuzzerRuntimeLibrary() string {
	return LibclangRuntimeLibrary("fuzzer")
}

func LibFuzzerRuntimeInterceptors() string {
	return LibclangRuntimeLibrary("fuzzer_interceptors")
}
