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

package modes

import (
	"fmt"
	"os"

	"github.com/mum4k/termdash/keyboard"
)

func ExitTermApp(cancel func(), msg string, code int) {
	os.Stdout.Write([]byte{0033, 0143})
	os.Stdout.Write([]byte{'\n'})
	cancel()
	fmt.Println(msg)
	os.Exit(code)
}
func IsExitKey(key keyboard.Key) bool {
	return key == keyboard.KeyCtrlC || key == keyboard.KeyEsc || key.String() == "q"
}
