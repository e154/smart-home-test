// This file is part of the Smart Home
// Program complex distribution https://github.com/e154/smart-home
// Copyright (C) 2016-2023, Filippov Alex
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

package supervisor

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"plugin"
	"sync"

	"github.com/e154/bus"
	"github.com/pkg/errors"
	"go.uber.org/atomic"

	"github.com/e154/smart-home/adaptors"
	"github.com/e154/smart-home/common"
	"github.com/e154/smart-home/common/apperr"
	"github.com/e154/smart-home/common/debug"
	"github.com/e154/smart-home/common/events"
	m "github.com/e154/smart-home/models"
)

var pluginsDir = path.Join("data", "plugins")

type pluginManager struct {
	adaptors       *adaptors.Adaptors
	isStarted      *atomic.Bool
	service        *service
	eventBus       bus.Bus
	enabledPlugins sync.Map
	pluginsWg      *sync.WaitGroup
}

// Start ...
func (p *pluginManager) Start(ctx context.Context) {
	if p.isStarted.Load() {
		return
	}
	defer p.isStarted.Store(true)

	p.loadPluginLibs(ctx)
	p.loadPlugins(ctx)

	log.Info("Started")
}

// Shutdown ...
func (p *pluginManager) Shutdown(ctx context.Context) {

	if !p.isStarted.Load() {
		return
	}
	defer p.isStarted.Store(false)

	p.enabledPlugins.Range(func(name, value any) bool {
		if enabled, _ := value.(bool); !enabled {
			return true
		}
		_ = p.unloadPlugin(ctx, name.(string))
		return true
	})

	p.pluginsWg.Wait()

	log.Info("Shutdown")
}

// GetPlugin ...
func (p *pluginManager) GetPlugin(t string) (plugin interface{}, err error) {

	plugin, err = p.getPlugin(t)

	return
}

func (p *pluginManager) getPlugin(name string) (plugin Pluggable, err error) {

	if item, ok := pluginList.Load(name); ok {
		plugin = item.(Pluggable)
		return
	}

	err = errors.Wrap(apperr.ErrNotFound, fmt.Sprintf("name %s", name))

	return
}

func (p *pluginManager) loadPluginLibs(ctx context.Context) error {

	list, _, err := p.ListPluginsDir(ctx)
	if err != nil {
		log.Error(err.Error())
		return err
	}

	for _, plugin := range list {
		if err := p.LoadPluginLib(plugin); err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

func (p *pluginManager) loadPlugins(ctx context.Context) {

	var page int64
	var loadList []*m.Plugin
	const perPage = 500
	var err error

LOOP:
	loadList, _, err = p.adaptors.Plugin.List(context.Background(), perPage, perPage*page, "", "", common.Bool(true), nil)
	if err != nil {
		log.Error(err.Error())
		return
	}

	for _, pl := range loadList {
		go func(pl *m.Plugin) {
			if err = p.loadPlugin(ctx, pl.Name); err != nil {
				log.Errorf("plugin name '%s', %s", pl.Name, err.Error())
			}
		}(pl)
	}

	if len(loadList) != 0 {
		page++
		goto LOOP
	}

	log.Info("all plugins loaded ...")
}

func (p *pluginManager) loadPlugin(ctx context.Context, name string) (err error) {

	if p.PluginIsLoaded(name) {
		err = errors.Wrap(apperr.ErrPluginIsLoaded, name)
		return
	}
	if item, ok := pluginList.Load(name); ok {
		plugin := item.(Pluggable)
		log.Infof("load plugin '%v'", plugin.Name())
		if err = plugin.Load(ctx, p.service); err != nil {
			err = errors.Wrap(err, "load plugin")
			return
		}
	} else {
		err = apperr.ErrNotFound
		return
	}

	p.enabledPlugins.Store(name, true)

	p.pluginsWg.Add(1)

	p.eventBus.Publish("system/plugins/"+name, events.EventPluginLoaded{
		PluginName: name,
	})

	return
}

func (p *pluginManager) unloadPlugin(ctx context.Context, name string) (err error) {

	if !p.PluginIsLoaded(name) {
		err = errors.Wrap(apperr.ErrPluginNotLoaded, name)
		return
	}

	if item, ok := pluginList.Load(name); ok {
		plugin := item.(Pluggable)
		log.Infof("unload plugin %v", plugin.Name())
		_ = plugin.Unload(ctx)
	} else {
		err = errors.Wrap(apperr.ErrNotFound, fmt.Sprintf("name %s", name))
	}

	p.enabledPlugins.Store(name, false)

	p.pluginsWg.Done()

	p.eventBus.Publish("system/plugins/+", events.EventPluginUnloaded{
		PluginName: string(name),
	})

	return
}

// Install ...
func (p *pluginManager) Install(ctx context.Context, t string) {

	pl, _ := p.adaptors.Plugin.GetByName(context.Background(), t)
	if pl.Enabled {
		return
	}

	plugin, err := p.getPlugin(t)
	if err != nil {
		return
	}

	if plugin.Type() != PluginInstallable {
		return
	}

	installable, ok := plugin.(Installable)
	if !ok {
		return
	}

	if err := installable.Install(); err != nil {
		log.Error(err.Error())
		return
	}

	_ = p.adaptors.Plugin.CreateOrUpdate(context.Background(), &m.Plugin{
		Name:    plugin.Name(),
		Version: plugin.Version(),
		Enabled: true,
		System:  plugin.Type() == PluginBuiltIn,
	})

	if err = p.loadPlugin(ctx, plugin.Name()); err != nil {
		log.Error(err.Error())
	}
}

// Uninstall ...
func (p *pluginManager) Uninstall(name string) {

}

// EnablePlugin ...
func (p *pluginManager) EnablePlugin(ctx context.Context, name string) (err error) {
	if err = p.loadPlugin(ctx, name); err != nil {
		return
	}
	if _, ok := pluginList.Load(name); !ok {
		err = errors.Wrap(apperr.ErrNotFound, fmt.Sprintf("name %s", name))
		return
	}
	var plugin *m.Plugin
	if plugin, err = p.adaptors.Plugin.GetByName(context.Background(), name); err != nil {
		err = errors.Wrap(apperr.ErrPluginGet, fmt.Sprintf("name %s", name))
		return
	}
	plugin.Enabled = true
	if err = p.adaptors.Plugin.CreateOrUpdate(context.Background(), plugin); err != nil {
		err = errors.Wrap(apperr.ErrPluginUpdate, fmt.Sprintf("name %s", name))
	}
	return
}

// DisablePlugin ...
func (p *pluginManager) DisablePlugin(ctx context.Context, name string) (err error) {
	if err = p.unloadPlugin(ctx, name); err != nil {
		return
	}
	if _, ok := pluginList.Load(name); !ok {
		err = errors.Wrap(apperr.ErrNotFound, fmt.Sprintf("name %s", name))
		return
	}
	var plugin *m.Plugin
	if plugin, err = p.adaptors.Plugin.GetByName(context.Background(), name); err != nil {
		err = errors.Wrap(apperr.ErrPluginGet, fmt.Sprintf("name %s", name))
		return
	}
	plugin.Enabled = false
	if err = p.adaptors.Plugin.CreateOrUpdate(context.Background(), plugin); err != nil {
		err = errors.Wrap(apperr.ErrPluginUpdate, fmt.Sprintf("name %s", name))
	}
	return
}

// PluginList ...
func (p *pluginManager) PluginList() (list []PluginInfo, total int64, err error) {

	list = make([]PluginInfo, 0)
	pluginList.Range(func(key, value interface{}) bool {
		total++
		plugin := value.(Pluggable)
		list = append(list, PluginInfo{
			Name:    plugin.Name(),
			Version: plugin.Version(),
			Enabled: p.PluginIsLoaded(plugin.Name()),
			System:  plugin.Type() == PluginBuiltIn,
		})
		return true
	})
	return
}

func (p *pluginManager) PluginIsLoaded(name string) (loaded bool) {
	if value, ok := p.enabledPlugins.Load(name); ok {
		loaded = value.(bool)
	}
	return
}

func (p *pluginManager) GetPluginReadme(ctx context.Context, name string, note *string, lang *string) (result []byte, err error) {
	var plugin Pluggable
	plugin, err = p.getPlugin(name)
	if err != nil {
		return
	}
	result, err = plugin.Readme(note, lang)
	return
}

func (p *pluginManager) LoadPluginLib(pluginInfo *PluginFileInfo) error {

	log.Infof("load external library %s", pluginInfo.Name)
	plugin, err := plugin.Open(path.Join(pluginsDir, pluginInfo.Name, "plugin.so"))
	if err != nil {
		return err
	}

	newFunc, err := plugin.Lookup("New")
	if err != nil {
		return err
	}

	pluggable, ok := newFunc.(func() Pluggable)
	if !ok {
		return errors.New("unexpected type from module symbol")
	}

	RegisterPlugin(pluginInfo.Name, pluggable)

	return nil
}

func (p *pluginManager) ListPluginsDir(ctx context.Context) (list PluginFileInfos, total int64, err error) {

	_ = filepath.Walk(pluginsDir, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return nil
		}
		if info.Name() == ".gitignore" || !info.IsDir() {
			return nil
		}
		if info.Name()[0:1] == "." {
			return nil
		}
		if info.Name() == "plugins" {
			return nil
		}
		list = append(list, &PluginFileInfo{
			Name:     info.Name(),
			Size:     info.Size(),
			FileMode: info.Mode(),
			ModTime:  info.ModTime(),
		})
		return nil
	})
	total = int64(len(list))
	//sort.Sort(list)

	debug.Println(list)
	return
}

func (p *pluginManager) UploadPlugin(ctx context.Context, reader *bufio.Reader, fileName string) (newFile *m.Plugin, err error) {

	//var list []*m.Backup
	//if list, _, err = b.List(ctx, 999, 0, "", ""); err != nil {
	//	return
	//}
	//
	//for _, file := range list {
	//	if fileName == file.Name {
	//		err = apperr.ErrBackupNameNotUnique
	//		return
	//	}
	//}
	//
	//buffer := bytes.NewBuffer(make([]byte, 0))
	//part := make([]byte, 128)
	//
	//var count int
	//for {
	//	if count, err = reader.Read(part); err != nil {
	//		break
	//	}
	//	buffer.Write(part[:count])
	//}
	//if err != io.EOF {
	//	return
	//}
	//
	//contentType := http.DetectContentType(buffer.Bytes())
	//log.Infof("Content-type from buffer, %s", contentType)
	//
	////create destination file making sure the path is writeable.
	//var dst *os.File
	//filePath := filepath.Join(b.cfg.Path, fileName)
	//if dst, err = os.Create(filePath); err != nil {
	//	return
	//}
	//
	//defer dst.Close()
	//
	////copy the uploaded file to the destination file
	//if _, err = io.Copy(dst, buffer); err != nil {
	//	return
	//}
	//
	//size, _ := common.GetFileSize(filePath)
	//newFile = &m.Backup{
	//	Name:     fileName,
	//	Size:     size,
	//	MimeType: contentType,
	//}
	//
	//b.eventBus.Publish("system/services/backup", events.EventUploadedBackup{
	//	Name: fileName,
	//})
	//
	//go b.RestoreFromChunks()

	return
}

func (p *pluginManager) RemovePlugin(ctx context.Context, pluginName string) (err error) {

	return
}
