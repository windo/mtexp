package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"

	"github.com/golang/freetype/truetype"
	"github.com/gopherjs/gopherjs/js"
	"github.com/llgcode/draw2d/draw2dimg"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"

	"github.com/hajimehoshi/ebiten"
	"github.com/hajimehoshi/ebiten/inpututil"
	"github.com/hajimehoshi/ebiten/text"
)

var (
	page        = Greeting
	lastPage    = Greeting
	experiment  *Experiment
	lastCorrect bool

	experimentID string
	epoch        int64
	idx          int

	resultsFile     = flag.String("results_file", "", "If set, instead of POST, save results to a local file.")
	imagesDirectory = flag.String("images_directory", "", "If set, save each experiment image to a local file in this directory.")

	face font.Face
	fix  *ebiten.Image
)

const (
	screenWidth  = 300
	screenHeight = 300

	postURL = "https://p6drad-teel.net/~windo/mtexp/post.php"
)

type Page int

const (
	UnknownPage Page = iota
	Greeting
	ShowExperiment
	Reading
	ThankYou
)

func (p Page) String() string {
	switch p {
	case Greeting:
		return "Greeting"
	case ShowExperiment:
		return "ShowExperiment"
	case Reading:
		return "Reading"
	case ThankYou:
		return "ThankYou"
	default:
		return fmt.Sprintf("Unknown[%d]", p)
	}
}

type Experiment struct {
	preDelay    time.Duration
	displayTime time.Duration
	postDelay   time.Duration
	image       image.Image
	expect      int
	question    string

	startTime           time.Time
	displayStartTime    time.Time
	measuredDisplayTime time.Duration
}

func (e *Experiment) Update(screen *ebiten.Image) error {
	emptyTime := time.Time{}

	if e.startTime == emptyTime {
		e.startTime = time.Now()
		log.Printf("Pre-delay")
		return nil
	}

	if e.startTime.Add(e.preDelay).After(time.Now()) {
		// Just wait
		return nil
	}

	if e.startTime.Add(e.preDelay + e.displayTime).After(time.Now()) {
		if e.displayStartTime == emptyTime {
			log.Printf("Experiment")
			e.displayStartTime = time.Now()
		}
		img, err := ebiten.NewImageFromImage(e.image, ebiten.FilterDefault)
		if err != nil {
			return err
		}
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(
			float64(screen.Bounds().Max.X-img.Bounds().Max.X)/2,
			float64(screen.Bounds().Max.Y-img.Bounds().Max.Y)/2,
		)
		screen.DrawImage(img, opts)
		return nil
	}

	if e.startTime.Add(e.preDelay + e.displayTime + e.postDelay).After(time.Now()) {
		if e.measuredDisplayTime == 0 {
			e.measuredDisplayTime = time.Now().Sub(e.displayStartTime)
			log.Printf("Post-delay, measured display was: %s", e.measuredDisplayTime)
		}
		// Just wait
		return nil
	}

	page = Reading
	return nil
}

func Greet(screen *ebiten.Image) error {
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		experiment = NewExperiment()
		idx++

		if *imagesDirectory != "" {
			draw2dimg.SaveToPngFile(fmt.Sprintf(
				"%s/impression-%s-%d-%d.png",
				*imagesDirectory, experimentID, epoch, idx,
			), experiment.image)
		}
		page = ShowExperiment
	}

	blankWithFix(screen)
	if experiment != nil {
		if lastCorrect {
			text.Draw(screen, "Correct!", face, 25, 260, color.RGBA{0x40, 0xff, 0x40, 0xff})
		} else {
			text.Draw(screen, "Incorrect!", face, 25, 260, color.RGBA{0xff, 0x40, 0x40, 0xff})
		}
	}
	text.Draw(screen, "Press space for next", face, 25, 280, color.White)
	return nil
}

func GetReading(screen *ebiten.Image) error {
	blankWithFix(screen)
	text.Draw(screen, fmt.Sprintf("Experiment #%d", idx), face, 25, 240, color.White)
	text.Draw(screen, experiment.question, face, 25, 260, color.White)
	text.Draw(screen, "Press 1 (yes) or 2 (no)", face, 25, 280, color.White)

	reading := 0
	if inpututil.IsKeyJustPressed(ebiten.Key1) {
		reading = 1
	}
	if inpututil.IsKeyJustPressed(ebiten.Key2) {
		reading = 2
	}

	if reading != 0 {
		if *resultsFile != "" {
			f, err := os.OpenFile(*resultsFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0)
			if err != nil {
				log.Fatalf("Failed to open file for writing: %v", err)
			}
			f.WriteString(fmt.Sprintf(
				"%q,%d,%d,%d,%d,%.3f",
				experimentID, epoch, idx, experiment.expect, reading, experiment.measuredDisplayTime.Seconds()))
			f.Close()
			log.Printf("Saved to file")
		} else {
			resp, err := http.PostForm(postURL,
				url.Values{
					"experiment_id": {experimentID},
					"epoch":         {fmt.Sprintf("%d", epoch)},
					"idx":           {fmt.Sprintf("%d", idx)},
					"expect":        {fmt.Sprintf("%d", experiment.expect)},
					"response":      {fmt.Sprintf("%d", reading)},
					"display_time": {fmt.Sprintf(
						"%.3f", experiment.measuredDisplayTime.Seconds())},
				})
			if err != nil {
				log.Fatalf("Failed to post experiment results: %v", err)
			}
			log.Printf("Post response: %q", resp.Status)
		}
		lastCorrect = reading == experiment.expect
		page = Greeting
	}
	return nil
}

func drawFix() (*ebiten.Image, error) {
	// Draw the experiment
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	gc := draw2dimg.NewGraphicContext(img)

	gc.SetLineWidth(1)
	gc.SetStrokeColor(color.White)
	gc.MoveTo(0, 5)
	gc.LineTo(10, 5)
	gc.Stroke()
	gc.MoveTo(5, 0)
	gc.LineTo(5, 10)
	gc.Stroke()
	gc.Close()
	return ebiten.NewImageFromImage(img, ebiten.FilterDefault)
}

func blankWithFix(screen *ebiten.Image) {
	screen.Fill(color.Black)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(
		float64(screen.Bounds().Max.X-fix.Bounds().Max.X)/2,
		float64(screen.Bounds().Max.Y-fix.Bounds().Max.Y)/2,
	)
	screen.DrawImage(fix, opts)
}

func update(screen *ebiten.Image) error {
	if lastPage != page {
		log.Printf("Page changed %s -> %s", lastPage, page)
		lastPage = page
	}

	screen.Fill(color.Black)

	switch page {
	case Greeting:
		return Greet(screen)
	case ShowExperiment:
		return experiment.Update(screen)
	case Reading:
		return GetReading(screen)
	case ThankYou:
	}

	return nil
}

func main() {
	var err error

	if runtime.GOARCH == "js" {
		location := js.Global.Get("location").Get("href").String()
		purl, err := url.Parse(location)
		if err != nil {
			log.Fatalf("Failed to parse location %q: %v", location, err)
		}
		experimentID = purl.Query().Get("experiment_id")
	} else {
		id := flag.String("experiment_id", "", "ID to store with each experiment in this run.")
		flag.Parse()
		experimentID = *id
	}
	epoch := time.Now().Unix()
	log.Printf("Epoch: %d, experiment ID: %q", epoch, experimentID)

	rand.Seed(time.Now().UnixNano())

	font, err := truetype.Parse(goregular.TTF)
	if err != nil {
		log.Fatalf("Failed to load goregular face: %v", err)
	}
	face = truetype.NewFace(font, &truetype.Options{Size: 20})

	fix, err = drawFix()
	if err != nil {
		log.Fatalf("Failed to draw a fix: %v", err)
	}

	if err = ebiten.Run(
		update,
		screenWidth, screenHeight, 1.0,
		"Mechanical Turk Experiment",
	); err != nil {
		log.Fatalf("Failed to run ebiten: %v", err)
	}
}
