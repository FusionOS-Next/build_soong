// Copyright 2018 Google Inc. All rights reserved.
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

package aconfig

import (
	"strings"
	"testing"

	"android/soong/android"
)

func TestAconfigDeclarations(t *testing.T) {
	bp := `
		aconfig_declarations {
			name: "module_name",
			package: "com.example.package",
			container: "com.android.foo",
			exportable: true,
			srcs: [
				"foo.aconfig",
				"bar.aconfig",
			],
		}
	`
	result := runTest(t, android.FixtureExpectsNoErrors, bp)

	module := result.ModuleForTests("module_name", "").Module().(*DeclarationsModule)

	// Check that the provider has the right contents
	depData, _ := android.SingletonModuleProvider(result, module, android.AconfigDeclarationsProviderKey)
	android.AssertStringEquals(t, "package", depData.Package, "com.example.package")
	android.AssertStringEquals(t, "container", depData.Container, "com.android.foo")
	android.AssertBoolEquals(t, "exportable", depData.Exportable, true)
	if !strings.HasSuffix(depData.IntermediateCacheOutputPath.String(), "/intermediate.pb") {
		t.Errorf("Missing intermediates proto path in provider: %s", depData.IntermediateCacheOutputPath.String())
	}
	if !strings.HasSuffix(depData.IntermediateDumpOutputPath.String(), "/intermediate.txt") {
		t.Errorf("Missing intermediates text path in provider: %s", depData.IntermediateDumpOutputPath.String())
	}
}

func TestAconfigDeclarationsWithExportableUnset(t *testing.T) {
	bp := `
		aconfig_declarations {
			name: "module_name",
			package: "com.example.package",
			container: "com.android.foo",
			srcs: [
				"foo.aconfig",
				"bar.aconfig",
			],
		}
	`
	result := runTest(t, android.FixtureExpectsNoErrors, bp)

	module := result.ModuleForTests("module_name", "").Module().(*DeclarationsModule)
	depData, _ := android.SingletonModuleProvider(result, module, android.AconfigDeclarationsProviderKey)
	android.AssertBoolEquals(t, "exportable", depData.Exportable, false)
}

func TestAconfigDeclarationsWithContainer(t *testing.T) {
	bp := `
		aconfig_declarations {
			name: "module_name",
			package: "com.example.package",
			container: "com.android.foo",
			srcs: [
				"foo.aconfig",
			],
		}
	`
	result := runTest(t, android.FixtureExpectsNoErrors, bp)

	module := result.ModuleForTests("module_name", "")
	rule := module.Rule("aconfig")
	android.AssertStringEquals(t, "rule must contain container", rule.Args["container"], "--container com.android.foo")
}

func TestMandatoryProperties(t *testing.T) {
	testCases := []struct {
		name          string
		expectedError string
		bp            string
	}{
		{
			name: "Srcs missing from aconfig_declarations",
			bp: `
				aconfig_declarations {
					name: "my_aconfig_declarations_foo",
					package: "com.example.package",
					container: "otherapex",
				}`,
			expectedError: `missing source files`,
		},
		{
			name: "Package missing from aconfig_declarations",
			bp: `
				aconfig_declarations {
					name: "my_aconfig_declarations_foo",
					container: "otherapex",
					srcs: ["foo.aconfig"],
				}`,
			expectedError: `missing package property`,
		},
		{
			name: "Container missing from aconfig_declarations",
			bp: `
				aconfig_declarations {
					name: "my_aconfig_declarations_foo",
					package: "com.example.package",
					srcs: ["foo.aconfig"],
				}`,
			expectedError: `missing container property`,
		},
	}
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			errorHandler := android.FixtureExpectsAtLeastOneErrorMatchingPattern(test.expectedError)
			android.GroupFixturePreparers(PrepareForTestWithAconfigBuildComponents).
				ExtendWithErrorHandler(errorHandler).
				RunTestWithBp(t, test.bp)
		})
	}
}
