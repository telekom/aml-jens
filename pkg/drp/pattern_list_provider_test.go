package drp

import (
	"path/filepath"
	"testing"

	"github.com/telekom/aml-jens/internal/assets/paths"
)

func compareDrps(expected []float64, got []float64, t *testing.T) {
	if len(expected) != len(got) {
		t.Fatalf("Expected %d entries in DRP, got: %d", len(expected), len(got))
	}
	for i, v := range got {
		if v != expected[i] {
			t.Fatalf("Got incorrect value in DRP: %f Expected: %f", v, expected[i])
		}
	}
}

var ExpectationSaw = []float64{
	10000,
	20000,
	30000,
	40000,
	50000,
	60000,
	70000,
	80000,
	90000,
	100000,
}

var PathSaw = filepath.Join(paths.TESTDATA_DRP(), "saw.csv")

func createExpectation(base []float64, sca float64, min float64) []float64 {
	expected := make([]float64, len(base))
	copy(expected, base)
	for i := range expected {
		if sca != 0 {
			expected[i] *= sca
		}
		if expected[i] < min {
			//Not using math.Min on purpose.
			expected[i] = min
		}
	}
	return expected
}
func HelperTestNewDrpFileListSingleScaleXMinY(sca float64, min float64, t *testing.T) {
	t.Helper()
	drp_list, err := NewDrpList(
		[]DataRatePatternProvider{
			NewDataRatePatternFileProvider(PathSaw),
		},
		[]struct {
			scale     float64
			minrateKb float64
		}{{scale: sca, minrateKb: min}},
		true,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(drp_list.Drps) != 1 {
		t.Fatalf("Expected one drp to be loaded, got: %d", len(drp_list.Drps))
	}
	expected := createExpectation(ExpectationSaw, sca, min)

	compareDrps(expected, (*drp_list.GetSelected().data), t)
}

func TestNewDrpFileListFolderScale0Min0(t *testing.T) {
	drp_list, err := NewDrpListFromFolder(paths.TESTDATA_DRP())
	if err != nil {
		t.Fatal(err)
	}
	if drp_list.Selected != 0 {
		t.Fatal("drp_list selected was not set to zero, inidcating an empty list")
	}
	read_ok := 0
	for i, v := range drp_list.Drps {
		if v.Name == "" {
			t.Fatalf("Name of drp %d from was not set", i)
		}
		switch v.Name {
		case "saw.csv":
			compareDrps(ExpectationSaw, *v.data, t)
			read_ok++
		case "drp_3valleys.csv":
			Validate3Valleys_Stats(t, v)
			read_ok++
		case "drp_3valleys_generic_comment.csv":
			Validate3Valleys_Stats(t, v)
			read_ok++
		case "drp_3valleys_kv_comment.csv":
			Validate3Valleys_Stats(t, v)
			read_ok++
		default:
			t.Fatalf("Got unexpected pattern: %s", v.Name)
		}
	}
	if read_ok != 4 {
		t.Fatalf("Expted %d patterns read, got: %d", 4, read_ok)
	}
}

func TestNewDrpFileListSingleScale1Min5001(t *testing.T) {
	HelperTestNewDrpFileListSingleScaleXMinY(1, 5001, t)
}
func TestNewDrpFileListSingleScale2Min5001(t *testing.T) {
	HelperTestNewDrpFileListSingleScaleXMinY(2, 5001, t)
}
func TestNewDrpFileListSingleScale1Min1(t *testing.T) {
	HelperTestNewDrpFileListSingleScaleXMinY(1, 1, t)
}
func TestNewDrpFileListSingleScale0Min0(t *testing.T) {
	HelperTestNewDrpFileListSingleScaleXMinY(0, 0, t)
}
