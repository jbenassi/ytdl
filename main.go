package main

import (
	f "fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/kkdai/youtube/v2"
	"io"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
)

// Helper function to download a file from a URL
func downloadFile(url, filename string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// random name generator 6 characters
func randomNameGenerator() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	b := make([]rune, 6)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func downloadAndConvert(videoURL string, statusLabel *widget.Label, progressBar *widget.ProgressBar) {
	// Create a new YouTube client
	client := youtube.Client{}

	// Get the video information
	video, err := client.GetVideo(videoURL)
	if err != nil {
		statusLabel.SetText("Error getting video info: " + err.Error())
		return
	}

	// Get the highest quality audio format
	format := video.Formats.Type("audio/mp4")[0] // Select the first available audio format

	// Get the download URL
	downloadURL, err := client.GetStreamURL(video, &format)
	if err != nil {
		statusLabel.SetText("Error getting download URL: " + err.Error())
		return
	}

	vFileName := f.Sprintf("%s-audio", randomNameGenerator())
	// Download the video
	err = downloadFile(downloadURL, f.Sprintf("%s.mp4", vFileName))
	if err != nil {
		statusLabel.SetText("Error downloading video: " + err.Error())
		return
	}

	// Update progress bar to indicate download completion
	progressBar.SetValue(0.5)
	statusLabel.SetText("Download completed. Converting...")

	// Convert the downloaded video to uncompressed WAV using ffmpeg
	cmd := exec.Command("ffmpeg", "-i", f.Sprintf("%s.mp4", vFileName), "-acodec", "pcm_s16le", "-ar", "44100", "-ac", "2", f.Sprintf("%s.wav", vFileName))
	err = cmd.Start()
	if err != nil {
		statusLabel.SetText("Error converting audio: " + err.Error())
		return
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			statusLabel.SetText("Error converting audio: " + err.Error())
			return
		}

		// Update progress bar to indicate conversion completion
		progressBar.SetValue(1.0)
		statusLabel.SetText(f.Sprintf("Audio file extracted to %s.wav", vFileName))
	}()
}

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("YouTube Downloader")

	urlEntry := widget.NewEntry()
	urlEntry.SetPlaceHolder("Enter YouTube URL")
	statusLabel := widget.NewLabel("")
	progressBar := widget.NewProgressBar()

	downloadButton := widget.NewButton("Download and Convert", func() {
		statusLabel.SetText("Processing...")
		progressBar.SetValue(0.5)
		go downloadAndConvert(urlEntry.Text, statusLabel, progressBar)
	})

	content := container.NewVBox(
		urlEntry,
		downloadButton,
		statusLabel,
		progressBar,
	)
	myWindow.SetFixedSize(true)
	myWindow.SetContent(content)
	myWindow.Resize(fyne.NewSize(800, 600))
	myWindow.ShowAndRun()
}
