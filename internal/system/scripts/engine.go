// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2024, Filippov Alex
//
// This library is free software: you can redistribute it and/or
// modify it under the terms of the GNU Lesser General Public
// License as published by the Free Software Foundation; either
// version 3 of the License, or (at your option) any later version.
//
// This library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Library General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public
// License along with this library.  If not, see
// <https://www.gnu.org/licenses/>.

package scripts

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/e154/smart-home/internal/system/scripts/require"
	"github.com/e154/smart-home/pkg/apperr"
	. "github.com/e154/smart-home/pkg/common"
	m "github.com/e154/smart-home/pkg/models"
	"github.com/e154/smart-home/pkg/scripts"

	"github.com/hashicorp/go-multierror"
	"go.uber.org/atomic"
)

var _ scripts.Engine = (*Engine)(nil)

// Engine ...
type Engine struct {
	model      *m.Script
	script     scripts.IScript
	buf        []string
	IsRun      atomic.Bool
	functions  *Pull
	structures *Pull
}

// NewEngine ...
func NewEngine(s *m.Script, functions, structures *Pull, loader require.SourceLoader) (engine *Engine, err error) {

	if s == nil {
		s = &m.Script{
			Lang: ScriptLangJavascript,
		}
	}

	if functions == nil {
		functions = NewPull()
	}

	if structures == nil {
		structures = NewPull()
	}

	engine = &Engine{
		model:      s,
		buf:        make([]string, 0),
		functions:  functions,
		structures: structures,
	}

	if s.Lang == "" {
		err = fmt.Errorf("%s: %w", fmt.Sprintf("language not specified"), apperr.ErrNotFound)
		return
	}

	switch s.Lang {
	case ScriptLangTs, ScriptLangCoffee, ScriptLangJavascript:
		engine.script = NewJavascript(engine, loader)
	default:
		err = fmt.Errorf("%s: %w", fmt.Sprintf("i don't know this language: \"%s\"", s.Lang), apperr.ErrNotFound)
		return
	}

	err = engine.script.Init()

	return
}

// Compile ...
func (s *Engine) Compile() (err error) {
	if err = s.script.Compile(); err != nil {
		err = fmt.Errorf("script id: %d: %w", s.model.Id, err)
	}
	return
}

// PushStruct ...
func (s *Engine) PushStruct(name string, i interface{}) {
	s.script.PushStruct(name, i)
}

// PushFunction ...
func (s *Engine) PushFunction(name string, i interface{}) {
	s.script.PushFunction(name, i)
}

// EvalString ...
func (s *Engine) EvalString(str ...string) (result string, errs error) {
	var err error
	if len(str) == 0 {
		if result, err = s.script.Do(); err != nil {
			errs = multierror.Append(err, errs)
		}
		return
	}
	for _, st := range str {
		if result, err = s.script.EvalString(st); err != nil {
			errs = multierror.Append(err, errs)
		}
	}
	return
}

// EvalScript ...
func (s *Engine) EvalScript(script *m.Script) (result string, err error) {
	programName := strconv.Itoa(int(script.Id))
	if result, err = s.script.RunProgram(programName); err == nil {
		return
	}
	if errors.Is(err, apperr.ErrNotFound) {
		if err = s.script.CreateProgram(programName, script.Compiled); err != nil {
			err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrInternal)
			return
		}
		result, err = s.script.RunProgram(programName)
	}
	return
}

// DoFull ...
func (s *Engine) DoFull() (res string, err error) {
	if !s.IsRun.CompareAndSwap(false, true) {
		return
	}
	defer s.IsRun.Store(false)

	var result string
	result, err = s.script.Do()
	if err != nil {
		err = fmt.Errorf("do full: %w", err)
		return
	}
	for _, r := range s.buf {
		res += r + "\n"
	}

	res += result + "\n"

	// reset buffer
	s.buf = []string{}

	return
}

// Do ...
func (s *Engine) Do() (string, error) {
	return s.script.Do()
}

// AssertFunction ...
func (s *Engine) AssertFunction(f string, arg ...interface{}) (result string, err error) {
	if !s.IsRun.CompareAndSwap(false, true) {
		return
	}
	defer s.IsRun.Store(false)

	result, err = s.script.AssertFunction(f, arg...)
	if err != nil {
		if s.ScriptId() != 0 {
			err = fmt.Errorf("script id:%d: %w", s.ScriptId(), err)
			return
		}
		return
	}

	// reset buffer
	s.buf = []string{}

	return
}

// Print ...
func (s *Engine) Print(v ...interface{}) {
	fmt.Println(v...)
	s.buf = append(s.buf, fmt.Sprint(v...))
}

// Get ...
func (s *Engine) Get() scripts.IScript {
	return s.script
}

// File ...
func (s *Engine) File(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Engine) ScriptId() int64 {
	return s.model.Id
}

func (s *Engine) Script() *m.Script {
	return s.model
}
