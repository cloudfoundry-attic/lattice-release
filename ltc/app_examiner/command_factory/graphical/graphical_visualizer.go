package graphical

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/cloudfoundry-incubator/lattice/ltc/app_examiner"
	"github.com/gizak/termui"
	"github.com/pivotal-golang/clock"
)

const (
	graphicalRateDelta = 100 * time.Millisecond
)

//go:generate counterfeiter -o fake_graphical_visualizer/fake_graphical_visualizer.go . GraphicalVisualizer
type GraphicalVisualizer interface {
	PrintDistributionChart(rate time.Duration) error
}

type graphicalVisualizer struct {
	appExaminer app_examiner.AppExaminer
}

type cellBar struct {
	index   int
	label   string
	running int
	claimed int
}

type cellBarSliceSortedByLabelNumber []cellBar

func (c cellBarSliceSortedByLabelNumber) Len() int { return len(c) }
func (c cellBarSliceSortedByLabelNumber) Less(i, j int) bool {
	return c[i].index < c[j].index
}
func (c cellBarSliceSortedByLabelNumber) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func NewGraphicalVisualizer(appExaminer app_examiner.AppExaminer) *graphicalVisualizer {
	return &graphicalVisualizer{
		appExaminer: appExaminer,
	}
}

func (gv *graphicalVisualizer) PrintDistributionChart(rate time.Duration) error {

	//Initialize termui
	err := termui.Init()
	if err != nil {
		return errors.New("Unable to initalize terminal graphics mode.")
		//panic(err)
	}
	defer termui.Close()
	if rate <= time.Duration(0) {
		rate = graphicalRateDelta
	}

	termui.UseTheme("helloworld")

	//Initalize some widgets
	p := termui.NewPar("Lattice Visualization")
	if p == nil {
		return errors.New("Error Initializing termui objects NewPar")
	}
	p.Height = 1
	p.Width = 25
	p.TextFgColor = termui.ColorWhite
	p.HasBorder = false

	r := termui.NewPar(fmt.Sprintf("rate:%v", rate))
	if r == nil {
		return errors.New("Error Initializing termui objects NewPar")
	}
	r.Height = 1
	r.Width = 10
	r.TextFgColor = termui.ColorWhite
	r.HasBorder = false

	s := termui.NewPar("hit [+=inc; -=dec; q=quit]")
	if s == nil {
		return errors.New("Error Initializing termui objects NewPar")
	}
	s.Height = 1
	s.Width = 30
	s.TextFgColor = termui.ColorWhite
	s.HasBorder = false

	bg := termui.NewMBarChart()
	if bg == nil {
		return errors.New("Error Initializing termui objects NewMBarChart")
	}
	bg.IsDisplay = false
	bg.Data[0] = []int{0}
	bg.DataLabels = []string{"Missing"}
	bg.Width = termui.TermWidth() - 10
	bg.Height = termui.TermHeight() - 5
	bg.BarColor[0] = termui.ColorGreen
	bg.BarColor[1] = termui.ColorYellow
	bg.NumColor[0] = termui.ColorRed
	bg.NumColor[1] = termui.ColorRed
	bg.TextColor = termui.ColorWhite
	bg.Border.LabelFgColor = termui.ColorWhite
	bg.Border.Label = "[X-Axis: Cells; Y-Axis: Instances]"
	bg.BarWidth = 10
	bg.BarGap = 1
	bg.ShowScale = true

	//12 column grid system
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 5, p)))
	termui.Body.AddRows(termui.NewRow(termui.NewCol(12, 0, bg)))
	termui.Body.AddRows(termui.NewRow(termui.NewCol(6, 0, s), termui.NewCol(6, 5, r)))

	termui.Body.Align()

	termui.Render(termui.Body)

	bg.IsDisplay = true
	clock := clock.NewClock()
	evt := termui.EventCh()
	for {
		select {
		case e := <-evt:
			if e.Type == termui.EventKey {
				switch {
				case (e.Ch == 'q' || e.Ch == 'Q'):
					return nil
				case (e.Ch == '+' || e.Ch == '='):
					rate += graphicalRateDelta
				case (e.Ch == '_' || e.Ch == '-'):
					rate -= graphicalRateDelta
					if rate <= time.Duration(0) {
						rate = graphicalRateDelta
					}
				}
				r.Text = fmt.Sprintf("rate:%v", rate)
				termui.Render(termui.Body)
			}
			if e.Type == termui.EventResize {
				termui.Body.Width = termui.TermWidth()
				termui.Body.Align()
				termui.Render(termui.Body)
			}
		case <-clock.NewTimer(rate).C():
			err := gv.getProgressBars(bg)
			if err != nil {
				return err
			}
			termui.Render(termui.Body)
		}
	}
	return nil
}

func (gv *graphicalVisualizer) getProgressBars(bg *termui.MBarChart) error {
	cells, err := gv.appExaminer.ListCells()
	if err != nil {
		return err
	}

	cellBars := []cellBar{}
	maxTotal := -1
	for _, cell := range cells {
		var bar cellBar
		if cell.Missing {
			bar = cellBar{label: "Missing"}
		} else {
			index := 0
			if strings.HasPrefix(cell.CellID, "cell-") {
				s := cell.CellID[5:len(cell.CellID)]
				if cellIndexInt, err := strconv.Atoi(s); err == nil {
					index = cellIndexInt
				}
			}

			bar = cellBar{
				label:   cell.CellID,
				index:   index,
				running: cell.RunningInstances,
				claimed: cell.ClaimedInstances,
			}

			total := cell.RunningInstances + cell.ClaimedInstances
			if total > maxTotal {
				maxTotal = total
			}
		}
		cellBars = append(cellBars, bar)
	}

	sort.Sort(cellBarSliceSortedByLabelNumber(cellBars))

	bg.Data[0] = getRunningFromBars(cellBars)
	bg.Data[1] = getClaimedFromBars(cellBars)
	bg.DataLabels = getLabelFromBars(cellBars)

	if maxTotal < 10 {
		bg.SetMax(10)
	} else {
		bg.SetMax(maxTotal)
	}

	return nil
}

func getRunningFromBars(cellBars []cellBar) []int {
	ints := []int{}
	for _, bar := range cellBars {
		ints = append(ints, bar.running)
	}
	return ints
}

func getClaimedFromBars(cellBars []cellBar) []int {
	ints := []int{}
	for _, bar := range cellBars {
		ints = append(ints, bar.claimed)
	}
	return ints
}

func getLabelFromBars(cellBars []cellBar) []string {
	labels := []string{}
	for _, bar := range cellBars {
		labels = append(labels, bar.label)
	}
	return labels
}
