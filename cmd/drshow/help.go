/*
 * aml-jens
 *
 * (C) 2023 Deutsche Telekom AG
 *
 * Deutsche Telekom AG and all other contributors /
 * copyright owners license this file to you under the Apache
 * License, Version 2.0 (the "License"); you may not use this
 * file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

package main

import (
	"fmt"
	"os"
	"strings"
)

type helpmode string

const (
	mode_str_generic helpmode = ""
	mode_str_pipe    helpmode = "pipe"
	mode_str_static  helpmode = "static"
)

var helpstringMap = map[string]bool{
	"HELP":   true,
	"H":      true,
	"--HELP": true,
	"-HELP":  true,
	"-H":     true,
}
var verstionstringMap = map[string]bool{
	"V":         true,
	"VERSION":   true,
	"--VERSION": true,
	"-VERSION":  true,
	"-V":        true,
}
var ModeStringMap = map[string]helpmode{
	"PIPE":   mode_str_pipe,
	"STATIC": mode_str_static,
	"FILE":   mode_str_static,
	"FOLDER": mode_str_static,
}

func parseArg2ForHelp(arg string) helpmode {
	arg = strings.ToUpper(arg)
	if arg == "PIPE" {
		return mode_str_pipe
	}
	if arg == "STATIC" {
		return mode_str_static
	}
	if arg == "FOLDER" {
		return mode_str_static
	}
	if arg == "FILE" {
		return mode_str_static
	}

	return mode_str_generic
}

func printHelp(part string) {
	switch part {
	case string(mode_str_pipe):
		fmt.Fprintf(os.Stdout, `usage: drshow || in pipeMode
This mode is meant to view the live output of drplay.
When no Data is received within the first few seconds, the application 
will quit automatically.
After a valid input is detected, the Program will NOT exit, even if drplay
exits. (A red message will appear in the top right)


In App behavior:
	The top right textfield will contain a list of flows. These flows represent
	network activity on a specific Src-/DstIp and Port.
	Using ArrowUp and ArrowDown the selected listentry can be moved. 
	The top middle field displays some information about the selected entry.
	The graphs below show the realtime data.
	The top graph shows a combination of Linkcapacity (white) and the current
	Load (any color) of the flow.
	All graphs can be zoomed and selected using the mouse.
	By pressing 'e' the currently selected flow is exported and saved according
	to the path specified in the config file.
	Exit Using Esc, 'q' or Ctrl-C
Config changes:
	Using the config-file the behavior of the graphs can be changed:
		Either keep all Entries or have a moving graph of the last n entries.
NOTE:
Depending on windowsize the size of the graphs might change. 
	If the height of the graphs is to small: Change fontsize/zoom out of terminal.
Depending on the terminal the colors might not be accurate.
Incase of a fatal crash the terminalwindow might be in a not sane state.
	To fix this, usually, the 'reset' command can reset the current terminal.
`)
	case string(mode_str_static):
		fmt.Fprintf(os.Stdout, `usage: drshow || in static-mode
This mode is meant to view static DataRatePatterns (DRPs).
When a directory is supplied, any and all .csv files matching the structure of a
DRP will be put into a list.

In App behavior:
	This list is navigable using ArrowUp and ArrowDown.
	Using the mouse: a portion of the graph can be zoomed in / selected.
	Exit Using Esc, 'q' or Ctrl-C
`)
	default: // Generic
		//           [-p | --paginate | -P | --no-pager] [--no-replace-objects] [--bar
		fmt.Fprintf(os.Stdout, `usage: drshow [--help] [-h] [help]
       File/FolderMode: [-p <path>] <path>
	   LivePreview: {drplay [...] | drshow }
These are common drshow usages:

a) File/FolderMode -> 'static'
	By specifying a directory containing or a single DataRatePattern (*.csv)
	the application will visualize the drp.
	For more help see 'drshow help [static | folder | file]'
b) LivePreview -> 'pipe'
	By not supplying any arguments and piping in the output of 'drplay'
	into stdin of drshow --> 'drplay | drshow'
	For more help see drshow 'help [pipe]'

Additional instructions and infos are located in the man-pages.  
`)
	}
}
