// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2024, Filippov Alex
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

package webdav

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/e154/smart-home/pkg/adaptors"
	"github.com/e154/smart-home/pkg/apperr"
	"github.com/e154/smart-home/pkg/events"
	m "github.com/e154/smart-home/pkg/models"
	scripts2 "github.com/e154/smart-home/pkg/scripts"

	"github.com/spf13/afero"
	"go.uber.org/atomic"

	"github.com/e154/bus"
)

type Scripts struct {
	*FS
	adaptors      *adaptors.Adaptors
	scriptService scripts2.ScriptService
	eventBus      bus.Bus
	rootDir       string
	done          chan struct{}
	isStarted     *atomic.Bool
	isSyncFiles   *atomic.Bool
	sync.Mutex
	fileInfos map[string]*FileInfo
}

func NewScripts(fs *FS) *Scripts {
	return &Scripts{
		FS:          fs,
		rootDir:     "scripts",
		isStarted:   atomic.NewBool(false),
		isSyncFiles: atomic.NewBool(false),
	}
}

func (s *Scripts) Start(adaptors *adaptors.Adaptors, scriptService scripts2.ScriptService, eventBus bus.Bus) {
	if !s.isStarted.CompareAndSwap(false, true) {
		return
	}

	s.adaptors = adaptors
	s.eventBus = eventBus
	s.scriptService = scriptService
	s.fileInfos = make(map[string]*FileInfo)

	s.preload()

	s.done = make(chan struct{})
	go func() {
		ticker := time.NewTicker(time.Second * 5)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				s.syncFiles()
			case <-s.done:
				return
			}
		}
	}()

	_ = eventBus.Subscribe("system/models/scripts/+", s.eventHandler)
}

func (s *Scripts) Shutdown() {
	if !s.isStarted.CompareAndSwap(true, false) {
		return
	}

	s.fileInfos = nil
	close(s.done)
	_ = s.eventBus.Unsubscribe("system/models/scripts/+", s.eventHandler)
}

// eventHandler ...
func (s *Scripts) eventHandler(_ string, message interface{}) {

	switch msg := message.(type) {
	case events.EventUpdatedScriptModel:
		go s.eventUpdateScript(msg)
	case events.EventRemovedScriptModel:
		go s.eventRemoveScript(msg)
	case events.EventCreatedScriptModel:
		go s.eventAddScript(msg)
	}
}

func (s *Scripts) eventAddScript(msg events.EventCreatedScriptModel) {
	if msg.Owner == events.OwnerSystem {
		return
	}
	filePath := s.getFilePath(msg.Script)
	if err := afero.WriteFile(s.Fs, filePath, []byte(msg.Script.Source), 0644); err != nil {
		log.Error(err.Error())
		return
	}
	_ = s.Fs.Chtimes(filePath, msg.Script.CreatedAt, msg.Script.CreatedAt)
	info, err := s.Fs.Stat(filePath)
	if err != nil {
		log.Error(err.Error())
		return
	}
	s.Lock()
	s.fileInfos[filePath] = &FileInfo{
		Size:      info.Size(),
		ModTime:   info.ModTime(),
		LastCheck: time.Now(),
	}
	s.Unlock()
}

func (s *Scripts) onRemoveHandler(ctx context.Context, filePath string) (err error) {
	s.Lock()
	defer s.Unlock()
	if err = s.removeScript(ctx, filePath); err != nil {
		return
	}
	_ = s.Fs.RemoveAll(filePath)
	delete(s.fileInfos, filePath)
	return
}

func (s *Scripts) eventRemoveScript(msg events.EventRemovedScriptModel) {
	if msg.Owner == events.OwnerSystem {
		return
	}
	filePath := s.getFilePath(msg.Script)
	_ = s.Fs.RemoveAll(filePath)
	s.Lock()
	delete(s.fileInfos, filePath)
	s.Unlock()
}

func (s *Scripts) eventUpdateScript(msg events.EventUpdatedScriptModel) {
	if msg.Owner == events.OwnerSystem {
		return
	}
	filePath := s.getFilePath(msg.Script)
	_ = afero.WriteFile(s.Fs, filePath, []byte(msg.Script.Source), 0644)
	_ = s.Fs.Chtimes(filePath, msg.Script.UpdatedAt, msg.Script.UpdatedAt)
	info, err := s.Fs.Stat(filePath)
	if err != nil {
		log.Error(err.Error())
		return
	}
	s.Lock()
	s.fileInfos[filePath] = &FileInfo{
		Size:      info.Size(),
		ModTime:   msg.Script.UpdatedAt,
		LastCheck: time.Now(),
	}
	if msg.OldScript != nil && msg.OldScript.Name != msg.Script.Name {
		filePath = s.getFilePath(msg.OldScript)
		_ = s.Fs.RemoveAll(filePath)
		delete(s.fileInfos, filePath)
	}
	s.Unlock()

}

func (s *Scripts) preload() {
	log.Info("Preload script list")

	var recordDir = filepath.Join(rootDir, s.rootDir)

	_ = s.Fs.MkdirAll(recordDir, 0755)

	var page int64
	var scripts []*m.Script
	const perPage = 500
	var err error

LOOP:

	if scripts, _, err = s.adaptors.Script.List(context.Background(), perPage, perPage*page, "desc", "id", nil, nil); err != nil {
		log.Error(err.Error())
		return
	}

	for _, script := range scripts {
		filePath := s.getFilePath(script)
		if err = afero.WriteFile(s.Fs, filePath, []byte(script.Source), 0644); err != nil {
			log.Error(err.Error())
		}
		if err = s.Fs.Chtimes(filePath, script.CreatedAt, script.UpdatedAt); err != nil {
			log.Error(err.Error())
		}
	}

	if len(scripts) != 0 {
		page++
		goto LOOP
	}

	s.Lock()
	defer s.Unlock()
	err = afero.Walk(s.Fs, filepath.Join(rootDir, s.rootDir), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		s.fileInfos[path] = &FileInfo{
			Size:      info.Size(),
			ModTime:   info.ModTime(),
			LastCheck: time.Now(),
		}

		return nil
	})
}

func (s *Scripts) getFilePath(script *m.Script) string {
	return filepath.Join(rootDir, s.rootDir, getFileName(script))
}

func (s *Scripts) removeScript(ctx context.Context, path string) (err error) {
	scriptName := extractScriptName(path)
	var script *m.Script
	script, err = s.adaptors.Script.GetByName(ctx, scriptName)
	if err == nil {
		log.Infof("remove script: %s, (id: %d)", script.Name, script.Id)
		if err = s.adaptors.Script.Delete(ctx, script.Id); err != nil {
			return
		}
		s.eventBus.Publish(fmt.Sprintf("system/models/scripts/%d", script.Id), events.EventRemovedScriptModel{
			Common: events.Common{
				Owner: events.OwnerSystem,
			},
			ScriptId: script.Id,
			Script:   script,
		})
	}
	return
}

func (s *Scripts) createScript(ctx context.Context, name string, fileInfo os.FileInfo) (err error) {
	scriptName := extractScriptName(fileInfo.Name())
	lang := getScriptLang(fileInfo.Name())

	if lang == "" {
		err = fmt.Errorf("bad file name %s", scriptName)
		return
	}

	var source []byte
	source, err = afero.ReadFile(s.Fs, name)
	if err != nil {
		return
	}

	if _, err = s.adaptors.Script.GetByName(ctx, scriptName); err == nil {
		return
	}

	log.Infof("create script: %s", scriptName)

	script := &m.Script{
		Name:   scriptName,
		Lang:   lang,
		Source: string(source),
	}
	engine, err := s.scriptService.NewEngine(script)
	if err != nil {
		return
	}
	if err = engine.Compile(); err != nil {
		err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrScriptCompile)
		return
	}

	if _, err = s.adaptors.Script.Add(ctx, script); err != nil {
		return err
	}

	s.eventBus.Publish(fmt.Sprintf("system/models/scripts/%d", script.Id), events.EventCreatedScriptModel{
		Common: events.Common{
			Owner: events.OwnerSystem,
		},
		ScriptId: script.Id,
		Script:   script,
	})

	return
}

func (s *Scripts) updateScript(ctx context.Context, name string, fileInfo os.FileInfo) (err error) {
	scriptName := extractScriptName(fileInfo.Name())
	lang := getScriptLang(fileInfo.Name())

	if lang == "" {
		err = errors.New("bad extension")
		return
	}

	var source []byte
	source, err = afero.ReadFile(s.Fs, name)
	if err != nil {
		return
	}

	log.Infof("update script: %s", scriptName)

	var script *m.Script
	script, err = s.adaptors.Script.GetByName(ctx, scriptName)
	if err == nil {
		script.Source = string(source)
		script.Lang = lang

		var engine scripts2.Engine
		engine, err = s.scriptService.NewEngine(script)
		if err != nil {
			return
		}
		if err = engine.Compile(); err != nil {
			err = fmt.Errorf("%s: %w", err.Error(), apperr.ErrScriptCompile)
			return
		}
		if err = s.adaptors.Script.Update(ctx, script); err != nil {
			return err
		}
		s.eventBus.Publish(fmt.Sprintf("system/models/scripts/%d", script.Id), events.EventUpdatedScriptModel{
			Common: events.Common{
				Owner: events.OwnerSystem,
			},
			ScriptId: script.Id,
			Script:   script,
		})
		return
	}
	return
}

func (s *Scripts) syncFiles() {
	if !s.isSyncFiles.CompareAndSwap(false, true) {
		return
	}
	defer s.isSyncFiles.Store(false)

	s.Lock()
	defer s.Unlock()

	for _, fileInfo := range s.fileInfos {
		fileInfo.IsInitialized = false
	}

	_ = afero.Walk(s.Fs, "/webdav/scripts", func(path string, info os.FileInfo, err error) error {
		if !s.isStarted.Load() {
			return errors.New("not started")
		}
		if info.IsDir() {
			return nil
		}
		fileInfo, ok := s.fileInfos[path]
		if ok {
			fileInfo.IsInitialized = true
			if info.ModTime().After(fileInfo.ModTime) || fileInfo.Size != info.Size() {
				log.Infof("File %s has changed.", path)
				fileInfo.Size = info.Size()
				fileInfo.ModTime = info.ModTime()
				fileInfo.LastCheck = time.Now()

				if _err := s.updateScript(context.Background(), path, info); _err != nil {
					if errors.Is(_err, apperr.ErrScriptNotFound) {
						fileInfo.IsInitialized = false
					}
					log.Error(_err.Error())
				}
			}
		} else {
			if _err := s.createScript(context.Background(), path, info); _err != nil {
				log.Error(_err.Error())
			}
			s.fileInfos[path] = &FileInfo{
				Size:          info.Size(),
				ModTime:       info.ModTime(),
				LastCheck:     time.Now(),
				IsInitialized: true,
			}
		}
		return nil
	})

	if !s.isStarted.Load() {
		return
	}

	for path, fileInfo := range s.fileInfos {
		if !s.isStarted.Load() {
			return
		}
		if !fileInfo.IsInitialized {
			if err := s.removeScript(context.Background(), path); err != nil {
				//log.Error(err.Error())
			}
			delete(s.fileInfos, path)
		}
	}
}
