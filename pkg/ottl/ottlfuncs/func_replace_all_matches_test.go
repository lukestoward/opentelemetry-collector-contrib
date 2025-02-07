// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ottlfuncs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"

	"github.com/open-telemetry/opentelemetry-collector-contrib/pkg/ottl"
)

func Test_replaceAllMatches(t *testing.T) {
	input := pcommon.NewMap()
	input.PutStr("test", "hello world")
	input.PutStr("test2", "hello")
	input.PutStr("test3", "goodbye")

	target := &ottl.StandardGetSetter[pcommon.Map]{
		Getter: func(ctx pcommon.Map) (interface{}, error) {
			return ctx, nil
		},
		Setter: func(ctx pcommon.Map, val interface{}) error {
			val.(pcommon.Map).CopyTo(ctx)
			return nil
		},
	}

	tests := []struct {
		name        string
		target      ottl.GetSetter[pcommon.Map]
		pattern     string
		replacement string
		want        func(pcommon.Map)
	}{
		{
			name:        "replace only matches",
			target:      target,
			pattern:     "hello*",
			replacement: "hello {universe}",
			want: func(expectedMap pcommon.Map) {
				expectedMap.PutStr("test", "hello {universe}")
				expectedMap.PutStr("test2", "hello {universe}")
				expectedMap.PutStr("test3", "goodbye")
			},
		},
		{
			name:        "no matches",
			target:      target,
			pattern:     "nothing*",
			replacement: "nothing {matches}",
			want: func(expectedMap pcommon.Map) {
				expectedMap.PutStr("test", "hello world")
				expectedMap.PutStr("test2", "hello")
				expectedMap.PutStr("test3", "goodbye")
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scenarioMap := pcommon.NewMap()
			input.CopyTo(scenarioMap)

			exprFunc, err := ReplaceAllMatches(tt.target, tt.pattern, tt.replacement)
			assert.NoError(t, err)

			result, err := exprFunc(scenarioMap)
			assert.NoError(t, err)
			assert.Nil(t, result)

			expected := pcommon.NewMap()
			tt.want(expected)

			assert.Equal(t, expected, scenarioMap)
		})
	}
}

func Test_replaceAllMatches_bad_input(t *testing.T) {
	input := pcommon.NewValueStr("not a map")
	target := &ottl.StandardGetSetter[interface{}]{
		Getter: func(ctx interface{}) (interface{}, error) {
			return ctx, nil
		},
		Setter: func(ctx interface{}, val interface{}) error {
			t.Errorf("nothing should be set in this scenario")
			return nil
		},
	}

	exprFunc, err := ReplaceAllMatches[interface{}](target, "*", "{replacement}")
	assert.NoError(t, err)
	result, err := exprFunc(input)
	assert.NoError(t, err)
	assert.Nil(t, result)
	assert.Equal(t, pcommon.NewValueStr("not a map"), input)
}

func Test_replaceAllMatches_get_nil(t *testing.T) {
	target := &ottl.StandardGetSetter[interface{}]{
		Getter: func(ctx interface{}) (interface{}, error) {
			return ctx, nil
		},
		Setter: func(ctx interface{}, val interface{}) error {
			t.Errorf("nothing should be set in this scenario")
			return nil
		},
	}

	exprFunc, err := ReplaceAllMatches[interface{}](target, "*", "{anything}")
	assert.NoError(t, err)
	result, err := exprFunc(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
