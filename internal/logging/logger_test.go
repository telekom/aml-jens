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

package logging

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
)

func TestGetLoggerPrefix(t *testing.T) {
	DEBUG, INFO, WARN, FATAL := GetLogger()
	if DEBUG.Prefix() != "[DEBUG] " {
		t.Fatalf("DebugPrefix was not set correctly. '%s'", DEBUG.Prefix())
	}
	if INFO.Prefix() != "[INFO] " {
		t.Fatalf("InfoPrefix was not set correctly. '%s'", INFO.Prefix())
	}
	if WARN.Prefix() != "[WARN] " {
		t.Fatalf("InfoPrefix was not set correctly. '%s'", WARN.Prefix())
	}
	if FATAL.Prefix() != "[FATAL] " {
		t.Fatalf("InfoPrefix was not set correctly. '%s'", FATAL.Prefix())
	}

}
func TestGetLoggerSingelton(t *testing.T) {
	_, INFOa, WARNa, FATALa := GetLogger()
	_, INFOb, WARNb, FATALb := GetLogger()
	if INFOa != INFOb {
		t.Fatal("Info Logger did not use singelton")
	}
	if WARNa != WARNb {
		t.Fatal("Warn Logger did not use singelton")
	}
	if FATALa != FATALb {
		t.Fatal("Fatal Logger did not use singelton")
	}

}
func doesFileExists(path string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return true, nil

	} else if errors.Is(err, os.ErrNotExist) {
		return false, nil
	} else {
		return false, err
	}
}
func postTest(t *testing.T, path string) {
	singelton_fatal_logger = nil
	log_path_exists, err := doesFileExists(path)
	if err != nil {
		t.Logf("Error during postTest: %s", err)
	}
	if log_path_exists {
		os.Remove(path)
	}
}
func TestGetLoggerFileCreation(t *testing.T) {
	log_path := filepath.Join(paths.LOG_PATH(), "TESTING.log")
	_, _, _, _ = GetLogger()
	InitLogger("TESTING")
	defer postTest(t, log_path)

	log_file_exists, err := doesFileExists(log_path)
	if err != nil {
		t.Fatal(err)
	}
	if !log_file_exists {
		t.Fatalf("Logging path '%s' was not created", log_path)
	}
}

func TestInitLoggerDEBUG(t *testing.T) {
	log_path := filepath.Join(paths.LOG_PATH(), "TESTING.log")
	os.Setenv("JENS_DEBUG", "1")
	_, _, _, _ = GetLogger()
	InitLogger("TESTING")
	defer postTest(t, log_path)
	DEBUG, _, _, _ := GetLogger()
	w := DEBUG.Writer()
	if w == io.Discard {
		t.Fatal("Debug output was not turned on")
	}
}
func TestInitLoggerDEBUG_0(t *testing.T) {
	log_path := filepath.Join(paths.LOG_PATH(), "TESTING.log")
	os.Setenv("JENS_DEBUG", "0")
	_, _, _, _ = GetLogger()
	InitLogger("TESTING")
	defer postTest(t, log_path)
	DEBUG, _, _, _ := GetLogger()
	w := DEBUG.Writer()
	if w != io.Discard {
		t.Fatal("Debug output was turned on")
	}
}
func TestInitLoggerDEBUG_not(t *testing.T) {
	os.Unsetenv("JENS_DEBUG")
	log_path := filepath.Join(paths.LOG_PATH(), "TESTING.log")
	_, _, _, _ = GetLogger()
	InitLogger("TESTING")
	defer postTest(t, log_path)
	DEBUG, _, _, _ := GetLogger()
	w := DEBUG.Writer()
	if w != io.Discard {
		t.Fatal("Debug output was turned on")
	}
}
