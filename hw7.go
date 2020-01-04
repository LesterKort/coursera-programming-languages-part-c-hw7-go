/*
 * MIT License
 * 
 * Copyright 2020 Lester Kortenhoeven
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package main

import (
	"encoding/json"
	"fmt"
	"github.com/LesterKort/coursera-programming-languages-part-c-hw7-go/geometry"
	"io/ioutil"
	"os"
)

func getValue(data interface{}, env map[string]interface{}, c chan<- interface{}) {
	switch dt := data.(type) {
	case map[string]interface{}:
		// eval data
		c <- eval(dt, env)
	case string:
		// lookup variable
		if out := env[dt]; out != nil {
			c <- out
		} else {
			panic(fmt.Sprintf("Unknown Variable %s", dt))
		}
	default:
		// output value
		c <- dt
	}
}

func getMultipleValues(data []interface{}, env map[string]interface{}) []chan interface{} {
	var lsChan []chan interface{}
	for i := range data {
		c := make(chan interface{})
		lsChan = append(lsChan, c)
		go getValue(data[i], env, c)
	}
	return lsChan
}

func eval(prog map[string]interface{}, env map[string]interface{}) interface{} {
	switch len(prog) {
	case 1:
		for cmd, data := range prog {
			switch cmd {
			case "Point":
				if len(data.([]interface{})) == 2 {
					lsChan := getMultipleValues(data.([]interface{}), env)
					return geometry.NewPoint((<-lsChan[0]).(float64), (<-lsChan[1]).(float64))
				} else {
					panic("Wrong Parameters Count")
				}
			case "Line":
				if len(data.([]interface{})) == 2 {
					lsChan := getMultipleValues(data.([]interface{}), env)
					return geometry.NewLine((<-lsChan[0]).(float64), (<-lsChan[1]).(float64))
				} else {
					panic("Wrong Parameters Count")
				}
			case "LineSegment":
				if len(data.([]interface{})) == 4 {
					lsChan := getMultipleValues(data.([]interface{}), env)
					return geometry.NewLineSegment((<-lsChan[0]).(float64), (<-lsChan[1]).(float64), (<-lsChan[2]).(float64), (<-lsChan[3]).(float64))
				} else {
					panic("Wrong Parameters Count")
				}
			case "Shift":
				if len(data.([]interface{})) == 3 {
					lsChan := getMultipleValues(data.([]interface{}), env)
					return geometry.Shift((<-lsChan[0]).(float64), (<-lsChan[1]).(float64), (<-lsChan[2]).(geometry.Value))
				} else {
					panic("Wrong Parameters Count")
				}
			case "Intersect":
				lsChan := getMultipleValues(data.([]interface{}), env)
				var result geometry.Value = geometry.Everywhere
				for i := range data.([]interface{}) {
					result = geometry.Intersect(result, (<-lsChan[i]).(geometry.Value))
				}
				return result
			}
		}
		panic("Unknown Command")
	case 2:
		for cmd, data := range prog {
			switch cmd {
			case "Let":
				if prog["in"] != nil {
					vars := data.(map[string]interface{})
					var lsChan []chan interface{}
					var lsName []string
					for name, exp := range vars {
						lsName = append(lsName, name)
						c := make(chan interface{})
						lsChan = append(lsChan, c)
						go getValue(exp, env, c)
					}
					new_env := make(map[string]interface{})
					for name, value := range env {
						new_env[name] = value
					}
					for i := range lsName {
						new_env[lsName[i]] = <-lsChan[i]
					}
					c := make(chan interface{})
					go getValue(prog["in"], new_env, c)
					return <-c
				} else {
					panic("\"Let\" without \"in\"")
				}
			}
		}
		panic("Unknown Command")
	default:
		panic("Invalid Syntax")
	}
}

func main() {
	prog_raw, _ := ioutil.ReadAll(os.Stdin)
	var prog_data interface{}
	if err := json.Unmarshal(prog_raw, &prog_data); err != nil {
		panic(err)
	}
	env := make(map[string]interface{})
	env["Nowhere"] = geometry.Nowhere
	env["Everywhere"] = geometry.Everywhere
	c := make(chan interface{})
	go getValue(prog_data, env, c)
	fmt.Printf("%#v\n", <-c)
}
