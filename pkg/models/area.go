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

package models

import (
	"time"
)

// Area ...
type Area struct {
	Polygon     []Point   `json:"polygon"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Name        string    `json:"name" validate:"required"`
	Description string    `json:"description"`
	Center      Point     `json:"center"`
	Id          int64     `json:"id"`
	Zoom        float32   `json:"zoom"`
	Resolution  float32   `json:"resolution"`
}

type AreaPayload struct {
	Center     Point   `json:"center"`
	Zoom       float32 `json:"zoom"`
	Resolution float32 `json:"resolution"`
}
