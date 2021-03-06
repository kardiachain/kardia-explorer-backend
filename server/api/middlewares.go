/*
 *  Copyright 2018 KardiaChain
 *  This file is part of the go-kardia library.
 *
 *  The go-kardia library is free software: you can redistribute it and/or modify
 *  it under the terms of the GNU Lesser General Public License as published by
 *  the Free Software Foundation, either version 3 of the License, or
 *  (at your option) any later version.
 *
 *  The go-kardia library is distributed in the hope that it will be useful,
 *  but WITHOUT ANY WARRANTY; without even the implied warranty of
 *  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 *  GNU Lesser General Public License for more details.
 *
 *  You should have received a copy of the GNU Lesser General Public License
 *  along with the go-kardia library. If not, see <http://www.gnu.org/licenses/>.
 */

package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo"

	"github.com/kardiachain/kardia-explorer-backend/types"
)

func checkPagination() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			limit, err := strconv.Atoi(c.QueryParam("limit"))
			if err != nil {
				limit = 20
			}
			if limit > types.MaximumLimit {
				return echo.NewHTTPError(http.StatusBadRequest, "")
			}
			return next(c)
		}
	}
}
