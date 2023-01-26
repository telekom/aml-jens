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

package shared

import (
	"context"
	"fmt"
	"time"

	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
)

const UPDATE_TEXT_CLEAR = "\"æ\"\r"

func NewTextBox(ctx context.Context, t terminalapi.Terminal, updateText <-chan string) (*text.Text, error) {
	wrapped, err := text.New(text.WrapAtRunes())
	if err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case t := <-updateText:
				if t == UPDATE_TEXT_CLEAR {
					wrapped.Reset()
				}
				wrapped.Write(fmt.Sprintf("%s\n", t))

			case <-ctx.Done():
				return
			}
		}
	}()
	return wrapped, nil
}

type StrWithTextOpts struct {
	str  string
	Opts []text.WriteOption
}

func NewStrTextWriteOps(s string) StrWithTextOpts {
	return StrWithTextOpts{
		str:  s,
		Opts: []text.WriteOption{},
	}
}

func NewTextWithOptsBoxAppendTopLast2(ctx context.Context, t terminalapi.Terminal, updateText <-chan StrWithTextOpts) (*text.Text, error) {
	last := ""
	wrapped, err := text.New(text.WrapAtRunes())
	if err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case t := <-updateText:
				wrapped.Reset()
				if t.str == UPDATE_TEXT_CLEAR {
					last = ""
				}

				wrapped.Write(fmt.Sprintf("%s\n", t.str), t.Opts...)
				if last != "" {
					wrapped.Write(fmt.Sprintf("%s\n", last))
				}
				last = t.str
			case <-ctx.Done():
				return
			}
		}
	}()
	return wrapped, nil
}
