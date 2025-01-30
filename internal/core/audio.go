package core

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"

	"bot/pkg/tech/e"
)

type Audio struct {
	URL   string
	Data  []byte
	Title string
}

func (a *Audio) ExtractAudio() error {
	log.Printf("downloading audio...")

	proxyURL := os.Getenv("PROXY_URL")

	fmt.Println(proxyURL)

	title, err := a.videoTitle(proxyURL)
	if err != nil {
		return e.Wrap("can't download audio with yt-dlp", err)
	}

	a.Title = title

	outputFile := fmt.Sprintf("%s.mp3", title)

	err = a.downloadAudioWithYTDLP(outputFile, proxyURL)
	if err != nil {
		return e.Wrap("failed to download audio", err)
	}

	videoData, err := os.ReadFile(outputFile)
	if err != nil {
		return e.Wrap("failed to read downloaded audio file", err)
	}
	defer os.Remove(outputFile)

	if err := checkFileSize(outputFile); err != nil {
		return e.Wrap("file size is too large", err)
	}

	a.Data = videoData

	log.Printf("audio was succesfully downloaded")

	return nil
}

func (a *Audio) downloadAudioWithYTDLP(outputFile, proxyURL string) error {
	_, err := exec.LookPath("yt-dlp")
	if err != nil {
		return e.Wrap("yt-dlp not found", err)
	}

	if err = updateYTDLP(); err != nil {
		return e.Wrap("can't download audio with yt-dlp", err)
	}

	if err := a.downloadAudioToFile(outputFile, proxyURL); err != nil {
		return e.Wrap("can't download audio with yt-dlp", err)
	}

	return nil
}

func (a *Audio) videoTitle(proxyURL string) (string, error) {
	titleCmd := exec.Command(
		"yt-dlp",
		"--print", "title",
		"--proxy", proxyURL,
		"--cookies-from-browser", "chrome",
		a.URL,
	)

	titleOutput, err := titleCmd.Output()
	if err != nil {
		return "", e.Wrap("failed to get video title", err)
	}

	return strings.TrimRight(string(titleOutput), "\n"), nil
}

func (a *Audio) downloadAudioToFile(outputFile, proxyURL string) error {
	cmdArgs := []string{
		"-x",
		"--audio-format", "mp3",
		"--proxy", proxyURL,
		"--cookies-from-browser", "chrome",
		"--no-post-overwrites",
		"--retries", "10",
		"--fragment-retries", "10",
		"--socket-timeout", "30",
		"-o", outputFile,
		a.URL,
	}

	downloadCmd := exec.Command("yt-dlp", cmdArgs...)

	downloadCmd.Stdout = os.Stdout
	downloadCmd.Stderr = os.Stderr

	done := make(chan error)
	go func() {
		done <- downloadCmd.Run()
	}()

	select {
	case err := <-done:
		if err != nil {
			return e.Wrap("yt-dlp command failed", err)
		}
	case <-time.After(60 * time.Second):
		downloadCmd.Process.Kill()
		return e.Wrap("yt-dlp process timed out", e.ErrProcessTimedOut)
	}

	return nil
}

func checkFileSize(outputFile string) error {
	fileInfo, err := os.Stat(outputFile)
	if err != nil {
		return e.Wrap("can't get file info", err)
	}

	const maxSize = 50 * 1024 * 1024
	if fileInfo.Size() > maxSize {
		return e.ErrFileSizeIsTooLarge
	}

	return nil
}

func updateYTDLP() error {
	updateCmd := exec.Command("yt-dlp", "-U")

	if err := updateCmd.Run(); err != nil {
		return e.Wrap("failed to update yt-dlp", err)
	}

	return nil
}
