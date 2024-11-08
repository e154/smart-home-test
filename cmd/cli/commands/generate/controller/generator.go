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

package controller

import (
	"os"
	"path"
	"strings"
	"text/template"

	"github.com/e154/smart-home/internal/common"
	"github.com/e154/smart-home/pkg/logger"

	"github.com/e154/smart-home/cmd/cli/commands/generate"
	"github.com/spf13/cobra"
)

var (
	log = logger.MustGetLogger("controller")
)

var controllerTpl = `//CODE GENERATED AUTOMATICALLY

package {{.Package}}

import (
	"context"
	"net/http"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/triggers/known/emptypb"

	"{{.Dir}}/api/stub/api"	
)

// {{.Name}} ...
type Controller{{.Name}} struct {
	*ControllerCommon
}

func NewController{{.Name}}(common *ControllerCommon) Controller{{.Name}} {
	return Controller{{.Name}}{
		ControllerCommon: common,
	}
}

// AddEntity{{.Name}} ...
func (c Controller{{.Name}}) AddEntity{{.Name}}(_ context.Context, req *api.New{{.Name}}Request) (*api.{{.Name}}, error) {

	image, errs, err := c.endpoint.{{.EndpointName}}.AddEntity(c.dto.{{.Name}}.FromNew{{.Name}}Request(req))
	if len(errs) > 0 {
		return nil, c.prepareErrors(errs)
	}

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.dto.{{.Name}}.To{{.Name}}(image), nil
}

// Get{{.Name}} ...
func (c Controller{{.Name}}) Get{{.Name}}ById(_ context.Context, req *api.Get{{.Name}}Request) (*api.{{.Name}}, error) {

	image, err := c.endpoint.{{.EndpointName}}.GetById(int64(req.Id))
	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.dto.{{.Name}}.To{{.Name}}(image), nil
}

// UpdateEntity{{.Name}}ById ...
func (c Controller{{.Name}}) UpdateEntity{{.Name}}ById(_ context.Context, req *api.UpdateEntity{{.Name}}Request) (*api.{{.Name}}, error) {

	image, errs, err := c.endpoint.{{.EndpointName}}.UpdateEntity(c.dto.{{.Name}}.FromUpdate{{.Name}}Request(req))
	if len(errs) > 0 {
		return nil, c.prepareErrors(errs)
	}

	if err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return c.dto.{{.Name}}.To{{.Name}}(image), nil
}

// Get{{.Name}}List ...
func (c Controller{{.Name}}) Get{{.Name}}List(_ context.Context, req *api.PaginationRequest) (*api.Get{{.Name}}ListResult, error) {

	pagination := c.Pagination(req.Page, req.Limit, req.Sort)
	items, total, err := c.endpoint.{{.EndpointName}}.List(ctx, pagination)
	if err != nil {
		return nil, c.error(ctx, nil, err)
	}

	return c.dto.{{.EndpointName}}.GetTaskList(items, uint64(total), pagination), nil
}

// Delete{{.Name}}ById ...
func (c Controller{{.Name}}) Delete{{.Name}}ById(_ context.Context, req *api.Delete{{.Name}}Request) (*emptypb.Empty, error) {

	if err := c.endpoint.{{.EndpointName}}.Delete(int64(req.Id)); err != nil {
		if err.Error() == "record not found" {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

`

var (
	controllerCmd = &cobra.Command{
		Use:   "c",
		Short: "controller generator",
		Long:  "$ cli g c [-endpoint=endpointName] [controllerName]",
	}
	endpointName string
	packageName  = "controllers"
)

func init() {
	generate.Generate.AddCommand(controllerCmd)
	controllerCmd.Flags().StringVarP(&endpointName, "endpoint", "e", "EndpointName", "EndpointName")
	controllerCmd.Run = func(cmd *cobra.Command, args []string) {

		if len(args) == 0 {
			log.Error("Wrong number of arguments. Run: cli help generate")
			return
		}

		currpath, _ := os.Getwd()

		g := Generator{}
		g.Generate(args[0], currpath)
	}
}

// Generator ...
type Generator struct{}

// Generate ...
func (e Generator) Generate(controllerName, currpath string) {

	log.Infof("Using '%s' as controller name", controllerName)
	log.Infof("Using '%s' as package name", packageName)

	fp := path.Join(currpath, "api", "controllers")

	e.addController(fp, controllerName)
}

func (e Generator) addController(fp, controllerName string) {

	if _, err := os.Stat(fp); os.IsNotExist(err) {
		// Create the controller's directory
		if err := os.MkdirAll(fp, 0777); err != nil {
			log.Errorf("Could not create controllers directory: %s", err.Error())
			return
		}
	}

	fpath := path.Join(fp, strings.ToLower(controllerName)+".go")
	f, err := os.OpenFile(fpath, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0666)
	if err != nil {
		log.Errorf("Could not create controller file: %s", err.Error())
		return
	}
	defer f.Close()

	templateData := struct {
		Package      string
		Name         string
		List         string
		EndpointName string
		Dir          string
	}{
		Dir:          common.Dir(),
		Package:      packageName,
		Name:         controllerName,
		EndpointName: endpointName,
	}
	t := template.Must(template.New("controller").Parse(controllerTpl))

	if t.Execute(f, templateData) != nil {
		log.Error(err.Error())
	}

	common.FormatSourceCode(fpath)
}
