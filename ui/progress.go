package ui

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"os"
	"time"
)

func NewSpinner(desc string) *progressbar.ProgressBar {
	return progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(desc),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetRenderBlankState(true),
	)

}

func RunWithSpinner(enable bool, desc string, fn func()) {
	if !enable {
		fn()
		return
	}
	bar := NewSpinner(desc)
	defer bar.Finish()
	defer bar.Clear()
	fn()
}	

