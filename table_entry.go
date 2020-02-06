// Copyright 2020 smilofoundation/regression Authors
// Copyright 2019 smilofoundation/regression Authors
// Copyright 2017 AMIS Technologies
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package go_smilo_regression

import (
	"reflect"

	"github.com/onsi/ginkgo"
)

/*
TableEntry represents an entry in a table test.  You generally use the `Entry` constructor.
*/
type tableEntry struct {
	Description string
	Parameters  []interface{}
	Pending     bool
	Focused     bool
}

func (t tableEntry) generate(itBody reflect.Value, entries []tableEntry, pending bool, focused bool) {
	if t.Pending {
		ginkgo.PDescribe(t.Description, func() {
			for _, entry := range entries {
				entry.generate(itBody, entries, pending, focused)
			}
		})
		return
	}

	values := []reflect.Value{}
	for i, param := range t.Parameters {
		var value reflect.Value

		if param == nil {
			inType := itBody.Type().In(i)
			value = reflect.Zero(inType)
		} else {
			value = reflect.ValueOf(param)
		}

		values = append(values, value)
	}

	body := func() {
		itBody.Call(values)
	}

	if t.Focused {
		ginkgo.FDescribe(t.Description, body)
	} else {
		ginkgo.Describe(t.Description, body)
	}
}

/*
Entry constructs a tableEntry.

The first argument is a required description (this becomes the content of the generated Ginkgo `It`).
Subsequent parameters are saved off and sent to the callback passed in to `DescribeTable`.

Each Entry ends up generating an individual Ginkgo It.
*/
func Case(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, false, false}
}

/*
You can focus a particular entry with FEntry.  This is equivalent to FIt.
*/
func FCase(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, false, true}
}

/*
You can mark a particular entry as pending with PEntry.  This is equivalent to PIt.
*/
func PCase(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, true, false}
}

/*
You can mark a particular entry as pending with XEntry.  This is equivalent to XIt.
*/
func XCase(description string, parameters ...interface{}) tableEntry {
	return tableEntry{description, parameters, true, false}
}
