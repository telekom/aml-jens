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

	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/telekom/aml-jens/internal/logging"
)

var DEBUG, INFO, WARN, FATAL = logging.GetLogger()

func NewStaticTextBox(ctx context.Context, t terminalapi.Terminal, txt []StrWithTextOpts) (*text.Text, error) {
	wrapped, err := text.New(text.WrapAtRunes())
	if err != nil {
		return nil, err
	}
	for _, v := range txt {
		if err := wrapped.Write(v.str+"\n", v.Opts...); err != nil {
			WARN.Println(err)
		}
	}

	return wrapped, nil
}

func NewStaticTextBoxQuitMessage(ctx context.Context, t terminalapi.Terminal, additionalText []StrWithTextOpts) (*text.Text, error) {
	return NewStaticTextBox(ctx, t, append([]StrWithTextOpts{NewStrTextWriteOps("Quit: <ESC> || <C-C> || <q>")}, additionalText...))

}
