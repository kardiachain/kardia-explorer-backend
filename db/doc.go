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

// Package db actually implement how explorer store data and retrieve data from storage
// Supported storage: mongoDB and postgres
package db

/*
Assume each blocks contain around [2000, 20000] txs,
and for each block, mainnet need about [1, 3] seconds for validate,
and we dont want over 2 blocks behind mainnet

Performance requirement:
- InsertBlock: 20% validate time
-
*/
