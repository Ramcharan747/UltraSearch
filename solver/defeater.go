package solver

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/chromedp"
)

type Point struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

var trajectoryPool [][]Point

// LoadTrajectories loads the pre-generated JSON pool into memory
func LoadTrajectories(filepath string) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}
	// The JSON is a list of paths. Each path is a list of [x, y] floats.
	var rawPool [][][]float64
	if err := json.Unmarshal(data, &rawPool); err != nil {
		return err
	}
	trajectoryPool = make([][]Point, len(rawPool))
	for i, rawPath := range rawPool {
		path := make([]Point, len(rawPath))
		for j, pt := range rawPath {
			path[j] = Point{X: pt[0], Y: pt[1]}
		}
		trajectoryPool[i] = path
	}
	log.Printf("🤖 [Solver] Loaded %d instant trajectories.", len(trajectoryPool))
	return nil
}

// GetInstantPath adapts a base trajectory to the new start/end points using affine transformation
func GetInstantPath(startX, startY, endX, endY float64) []Point {
	if len(trajectoryPool) == 0 {
		return []Point{{X: startX, Y: startY}, {X: endX, Y: endY}} // Fallback linear
	}
	base := trajectoryPool[rand.Intn(len(trajectoryPool))]
	n := len(base)
	warped := make([]Point, n)
	
	// Base is guaranteed to start at (0,0) and end at (100,100) based on python logic
	adx := base[n-1].X - base[0].X // usually 100
	ady := base[n-1].Y - base[0].Y // usually 100

	for i := 0; i < n; i++ {
		alpha := float64(i) / float64(n-1)
		wx := base[i].X + alpha*((endX-startX)-adx) - base[0].X + startX
		wy := base[i].Y + alpha*((endY-startY)-ady) - base[0].Y + startY
		warped[i] = Point{X: wx, Y: wy}
	}
	return warped
}

func humanDelays(n int, totalMs int) []float64 {
	delays := make([]float64, n)
	sum := 0.0
	for i := 0; i < n; i++ {
		t := float64(i) / float64(n-1)
		speed := 0.3 + 0.7*math.Sin(math.Pi*t)
		delay := 1.0 / (speed + 0.01)
		
		// Add tiny jitter
		jitter := (rand.Float64() - 0.5) * 0.004
		delay += jitter
		
		delays[i] = delay
		sum += delay
	}
	
	totalSec := float64(totalMs) / 1000.0
	for i := 0; i < n; i++ {
		delays[i] = delays[i] / sum * totalSec
		if delays[i] < 0.002 {
			delays[i] = 0.002
		}
		if delays[i] > 0.05 {
			delays[i] = 0.05
		}
	}
	return delays
}

// ExecuteHumanPath executes a generated mouse path on the target context and shows a visual cursor
func ExecuteHumanPath(ctx context.Context, path []Point, totalMs int) error {
	// Inject a visible red dot to act as the fake cursor
	cursorJS := `(() => {
		if (document.getElementById('fake-cursor')) return;
		const cursor = document.createElement('div');
		cursor.id = 'fake-cursor';
		cursor.style.position = 'fixed';
		cursor.style.width = '12px';
		cursor.style.height = '12px';
		cursor.style.backgroundColor = 'red';
		cursor.style.borderRadius = '50%';
		cursor.style.zIndex = '999999';
		cursor.style.pointerEvents = 'none';
		cursor.style.transition = 'none';
		document.body.appendChild(cursor);
	})();`
	chromedp.Run(ctx, chromedp.Evaluate(cursorJS, nil))

	delays := humanDelays(len(path), totalMs)
	for i, pt := range path {
		// Update the visual cursor position
		moveJS := fmt.Sprintf("document.getElementById('fake-cursor').style.left = '%fpx'; document.getElementById('fake-cursor').style.top = '%fpx';", pt.X, pt.Y)
		
		err := chromedp.Run(ctx, 
			chromedp.Evaluate(moveJS, nil),
			chromedp.ActionFunc(func(c context.Context) error {
				return input.DispatchMouseEvent(input.MouseMoved, pt.X, pt.Y).Do(c)
			}),
		)
		if err != nil {
			return err
		}
		time.Sleep(time.Duration(delays[i] * float64(time.Second)))
	}
	return nil
}

type CaptchaCoords struct {
	Type string  `json:"type"`
	CX   float64 `json:"cx"`
	CY   float64 `json:"cy"`
}

var LocateJS = `(() => {
	const iframes = document.querySelectorAll('iframe');
	for (const iframe of iframes) {
		const src = iframe.src || '';
		if (src.includes('api2/anchor') || src.includes('recaptcha/enterprise/anchor')) {
			const r = iframe.getBoundingClientRect();
			if (r.width > 0 && r.height > 0) {
				return { type: 'recaptcha', cx: r.x + 28, cy: r.y + r.height / 2 };
			}
		}
		if (src.includes('challenges.cloudflare.com')) {
			const r = iframe.getBoundingClientRect();
			if (r.width > 0 && r.height > 0) {
				let cx = r.x + 28;
				if (r.width > 320) cx = r.x + (r.width / 2) - 150 + 22;
				return { type: 'turnstile', cx: cx, cy: r.y + r.height / 2 };
			}
		}
		if (src.includes('datadome')) {
			const r = iframe.getBoundingClientRect();
			if (r.width > 0 && r.height > 0) {
				return { type: 'datadome_iframe', cx: r.x + r.width / 2, cy: r.y + r.height / 2 };
			}
		}
	}
	
	// Turnstile in-DOM widget (no iframe)
	const cfWidget = document.querySelector('input[id^="cf-chl-widget-"]');
	if (cfWidget) {
		// The input itself is hidden (0x0), we need its parent container which has the bounding rect
		const container = cfWidget.closest('div').parentElement;
		if (container) {
			const r = container.getBoundingClientRect();
			if (r.width > 0 && r.height > 0) {
				// The widget is left-aligned in its container. 
				// The checkbox is typically ~28-30px from the left edge.
				let cx = r.x + 28;
				return { type: 'turnstile_widget', cx: cx, cy: r.y + 34 }; // 68px height, center is 34
			}
		}
	}

	const submitBtn = document.querySelector('input[type="submit"]');
	if (submitBtn) {
		const r = submitBtn.getBoundingClientRect();
		if (r.width > 0) {
			return { type: 'google_submit', cx: r.x + r.width / 2, cy: r.y + r.height / 2 };
		}
	}
	
	// DataDome or generic challenge fallback
	const bodyText = document.body ? document.body.innerText.toLowerCase() : '';
	if (bodyText.includes('verify you are human') || bodyText.includes('datadome') || bodyText.includes('security check') || bodyText.includes('just a moment')) {
		return { type: 'telemetry_only', cx: window.innerWidth / 2, cy: window.innerHeight / 2 };
	}
	return null;
})()`

// LocateCaptchaCoordinates scans the DOM
func LocateCaptchaCoordinates(ctx context.Context) (*CaptchaCoords, error) {
	var res *CaptchaCoords
	for attempt := 0; attempt < 3; attempt++ {
		err := chromedp.Run(ctx, chromedp.Evaluate(LocateJS, &res))
		if err != nil {
			return nil, err
		}
		if res != nil {
			return res, nil
		}
		time.Sleep(800 * time.Millisecond)
	}
	return nil, fmt.Errorf("CAPTCHA not found in DOM")
}

// DefeatCaptcha orchestrates the entire process in Go
func DefeatCaptcha(ctx context.Context, currentX, currentY float64) (bool, error) {
	log.Println("🔍 [Solver] Scanning DOM for CAPTCHA...")
	coords, err := LocateCaptchaCoordinates(ctx)
	if err != nil || coords == nil {
		log.Println("❌ [Solver] Could not locate CAPTCHA.")
		return false, err
	}
	
	log.Printf("🎯 [Solver] Target: %s at (%.0f, %.0f)", coords.Type, coords.CX, coords.CY)
	path := GetInstantPath(currentX, currentY, coords.CX, coords.CY)
	
	log.Printf("🖱️  [Solver] Strike: (%.0f,%.0f) → (%.0f,%.0f)", currentX, currentY, coords.CX, coords.CY)
	totalMs := rand.Intn(300) + 400 // 400ms - 700ms
	if err := ExecuteHumanPath(ctx, path, totalMs); err != nil {
		return false, err
	}
	
	if coords.Type == "telemetry_only" {
		log.Println("🤖 [Solver] Performing telemetry injection (scroll & hover)...")
		chromedp.Run(ctx, chromedp.Evaluate(`window.scrollBy(0, 300)`, nil))
		time.Sleep(2 * time.Second)
		return true, nil
	}

	time.Sleep(time.Duration(rand.Float64()*170+80) * time.Millisecond) // 80ms - 250ms
	
	log.Println("👆 [Solver] Click!")
	err = chromedp.Run(ctx, chromedp.ActionFunc(func(c context.Context) error {
		p1 := input.DispatchMouseEvent(input.MousePressed, coords.CX, coords.CY).WithButton("left").WithClickCount(1)
		if err := p1.Do(c); err != nil {
			return err
		}
		time.Sleep(time.Duration(rand.Intn(40)+80) * time.Millisecond) // 80-120ms
		p2 := input.DispatchMouseEvent(input.MouseReleased, coords.CX, coords.CY).WithButton("left").WithClickCount(1)
		return p2.Do(c)
	}))
	if err != nil {
		return false, err
	}
	
	// Wait a moment for the challenge to resolve
	time.Sleep(3 * time.Second)
	return true, nil
}
