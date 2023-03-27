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

package flowdata

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/telekom/aml-jens/internal/assets"
	"github.com/telekom/aml-jens/internal/config"
)

type NetEndPoint struct {
	Host string
	Port int
}

func NewNetEndPoint(host string, port int) *NetEndPoint {
	return &NetEndPoint{
		Host: host,
		Port: port,
	}
}
func NewNetEndPointFromString(combined string) *NetEndPoint {
	data := strings.Split(combined, ":")
	port, err := strconv.ParseInt(data[1], 10, 32)
	if err != nil {
		return nil
	}
	return &NetEndPoint{
		Host: data[0],
		Port: int(port),
	}
}
func (self *NetEndPoint) Equals(other *NetEndPoint) bool {
	return other != nil && self.Host == other.Host &&
		self.Port == other.Port
}
func (self *NetEndPoint) Str() string {
	return fmt.Sprintf("%s:%d", self.Host, self.Port)
}

type FlowT struct {
	D      FlowDatapointsT
	Src    NetEndPoint
	Dst    NetEndPoint
	FlowId int32 //Relic of the past. used for id-ing. set by manager!
	Prio   int
}

func (s *FlowT) Color() uint8 {
	color := uint8(s.FlowId)
	if color > 14 {
		color += 1
	}
	return color
}

func (s *FlowT) identifier() string {
	return fmt.Sprintf("%s-%s", s.Src.Str(), s.Dst.Str())
}

func (self *FlowT) hasCombinedLoadGreater1() bool {
	return self.D.HasLoadGreater1
}

func (self *FlowT) ExportToFile(path string) error {

	file_name := fmt.Sprintf(assets.DRSHOW_EXPORT_PATH_NAME, self.identifier(), time.Now().Round(time.Second).Format("2006-01-02_15-04-05"))
	file_name, _ = filepath.Abs(filepath.Join(config.ShowCfg().ExportPathPrefix, file_name))
	INFO.Printf("Trying to export [%s] to %s\n", self.identifier(), file_name)
	var file *os.File
	var err error
	if file, err = os.OpenFile(file_name, os.O_CREATE|os.O_WRONLY, 0666); err != nil {
		return err
	}

	csvWriter := csv.NewWriter(file)
	defer file.Close()
	heading := assets.CONST_HEADING
	// = config.ShowCfg().MagicHeader but I can't use that
	if err := csvWriter.Write(heading); err != nil {
		return err
	}
	line := make([]string, len(assets.CONST_HEADING))
	line[i_netw] = self.identifier()
	/*line[config.I_dstIP] = self.DstIP
	line[config.I_srcIP] = self.SrcIP
	line[config.I_flowid] = fmt.Sprint((self.FlowId))*/
	for i := 0; i < self.D.Length(); i++ {
		line[i_timestamp] = fmt.Sprint(int64((*self.D.TimeStamp)[i]))
		line[i_capacity] = fmt.Sprint(int64((*self.D.Capacity)[i]))
		line[i_drop] = fmt.Sprint(int64((*self.D.Dropp)[i]))
		line[i_ecn] = fmt.Sprint(int64((*self.D.Ecn)[i]))
		line[i_load] = fmt.Sprint(int64((*self.D.Load)[i]))
		line[i_sojourntime] = fmt.Sprint(int64((*self.D.Sojourn)[i]))
		line[i_prio] = fmt.Sprint(self.Prio)
		if err = csvWriter.Write(line); err != nil {
			return err
		}
	}
	csvWriter.Flush()
	INFO.Printf("Done Exporting [%s]\n", self.identifier())
	return nil
}

func (self *FlowT) FmtString() string {
	return fmt.Sprintf(
		"FlowId:\n------------\nSrc:%s\nDst:%s\nSamples:%d",
		self.Src.Str(),
		self.Dst.Str(),
		self.D.Length(),
	)
}

func NewFlow(src string, dst string, prio string) *FlowT {
	a, err := strconv.Atoi(prio)
	if err != nil {
		a = 9
	}

	return &FlowT{
		Src:  *NewNetEndPointFromString(src),
		Dst:  *NewNetEndPointFromString(dst),
		D:    *NewFlowDataPoints(),
		Prio: a}
}

func (self *FlowT) Equals(other *FlowT) bool {
	return other != nil && self.Src.Equals(&other.Src) &&
		self.Dst.Equals(&other.Dst)
}
