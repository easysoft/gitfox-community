// Copyright 2023 Harness, Inc.
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

package foreachref

import (
	"errors"
	"io"
	"testing"
)

func TestParserErrReturnsReaderError(t *testing.T) {
	pipeRead, pipeWrite := io.Pipe()
	defer pipeRead.Close()

	format := NewFormat("refname")
	parser := format.Parser(pipeRead)

	errExpected := errors.New("git command failed")
	_ = pipeWrite.CloseWithError(errExpected)

	if got := parser.Next(); got != nil {
		t.Fatalf("expected nil result on reader error, got %#v", got)
	}

	if err := parser.Err(); !errors.Is(err, errExpected) {
		t.Fatalf("expected error %v, got %v", errExpected, err)
	}
}
