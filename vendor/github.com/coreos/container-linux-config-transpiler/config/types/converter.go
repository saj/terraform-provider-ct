// Copyright 2017 CoreOS, Inc.
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

package types

import (
	"reflect"

	"github.com/coreos/container-linux-config-transpiler/config/astyaml"
	ignTypes "github.com/coreos/ignition/config/v2_0/types"
	"github.com/coreos/ignition/config/validate"
	"github.com/coreos/ignition/config/validate/report"
)

type converterFor2_0 func(in Config, ast validate.AstNode, out ignTypes.Config, platform string) (ignTypes.Config, report.Report, validate.AstNode)

var convertersFor2_0 []converterFor2_0

func register2_0(f converterFor2_0) {
	convertersFor2_0 = append(convertersFor2_0, f)
}

func ConvertAs2_0(in Config, platform string, ast validate.AstNode) (ignTypes.Config, report.Report) {
	// convert our tree from having yaml tags to having json tags, so when Validate() is
	// called on the tree, it can find the keys in the ignition structs (which are denoted
	// by `json` tags)
	if asYamlNode, ok := ast.(astyaml.YamlNode); ok {
		asYamlNode.ChangeTreeTag("json")
		ast = asYamlNode
	}

	out := ignTypes.Config{
		Ignition: ignTypes.Ignition{
			Version: ignTypes.IgnitionVersion{Major: 2, Minor: 0},
		},
	}

	r := report.Report{}

	for _, convert := range convertersFor2_0 {
		var subReport report.Report
		out, subReport, ast = convert(in, ast, out, platform)
		r.Merge(subReport)
	}
	if r.IsFatal() {
		return ignTypes.Config{}, r
	}

	validationReport := validate.Validate(reflect.ValueOf(out), ast, nil, false)
	r.Merge(validationReport)
	if r.IsFatal() {
		return ignTypes.Config{}, r
	}

	return out, r
}
