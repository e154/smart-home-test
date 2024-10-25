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

package adaptors

import (
	"context"

	m "github.com/e154/smart-home/models"
)

// TemplateRepo ...
type TemplateRepo interface {
	UpdateOrCreate(ctx context.Context, ver *m.Template) (err error)
	Create(ctx context.Context, ver *m.Template) (err error)
	UpdateStatus(ctx context.Context, ver *m.Template) (err error)
	GetList(ctx context.Context, templateType m.TemplateType) (items []*m.Template, err error)
	GetByName(ctx context.Context, name string) (ver *m.Template, err error)
	GetItemByName(ctx context.Context, name string) (ver *m.Template, err error)
	GetItemsSortedList(ctx context.Context) (count int64, items []string, err error)
	Delete(ctx context.Context, name string) (err error)
	GetItemsTree(ctx context.Context) (tree []*m.TemplateTree, err error)
	UpdateItemsTree(ctx context.Context, tree []*m.TemplateTree) (err error)
	Search(ctx context.Context, query string, limit, offset int) (list []*m.Template, total int64, err error)
	GetMarkers(ctx context.Context, template *m.Template) (err error)
	Render(ctx context.Context, name string, params map[string]interface{}) (render *m.TemplateRender, err error)
}
