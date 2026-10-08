// Command fitness-import loads exercises (and their photos) from
// free-exercise-db into the fitness data folder.
//
// It never overwrites an exercise that already exists, so it is safe to
// re-run and never clobbers hand edits. Photos already on disk are skipped.
//
//	go run ./cmd/fitness-import -data data/fitness
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/XanderOnGithub/home-tools/internal/users"
	"github.com/XanderOnGithub/home-tools/tools/fitness"
)

const (
	datasetURL = "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/dist/exercises.json"
	imagesURL  = "https://raw.githubusercontent.com/yuhonas/free-exercise-db/main/exercises/"
	workers    = 8 // parallel photo downloads; polite to GitHub, fast enough
)

var client = &http.Client{Timeout: 30 * time.Second}

func main() {
	dataDir := flag.String("data", "data/fitness", "fitness data folder")
	usersDir := flag.String("users", "data/users", "shared profiles folder (fitness checks its data against it)")
	withImages := flag.Bool("images", true, "download exercise photos")
	flag.Parse()

	if err := run(*dataDir, *usersDir, *withImages); err != nil {
		log.Fatal(err)
	}
}

func run(dataDir, usersDir string, withImages bool) error {
	body, err := fetch(datasetURL)
	if err != nil {
		return err
	}
	exercises, err := fitness.ParseFreeExerciseDB(body)
	if err != nil {
		return err
	}

	us, err := users.Open(usersDir)
	if err != nil {
		return err
	}
	store, err := fitness.Open(dataDir, us)
	if err != nil {
		return err
	}

	var added, skipped int
	var images []string
	for _, ex := range exercises {
		if _, exists := store.Exercise(ex.ID); exists {
			skipped++
		} else {
			if err := store.SaveExercise(ex); err != nil {
				return err
			}
			added++
		}
		images = append(images, ex.Images...)
	}
	fmt.Printf("exercises: %d added, %d already present\n", added, skipped)

	if !withImages {
		return nil
	}
	got, err := downloadImages(filepath.Join(dataDir, "images"), images)
	fmt.Printf("photos: %d downloaded, %d already present\n", got, len(images)-got)
	return err
}

// downloadImages fetches each path not already on disk, using a fixed pool
// of workers fed by a channel. It returns how many were downloaded and the
// first error (other downloads still finish).
func downloadImages(dir string, paths []string) (int, error) {
	jobs := make(chan string)
	var (
		wg         sync.WaitGroup
		downloaded atomic.Int64
		firstErr   error
		errOnce    sync.Once
	)

	for range workers {
		wg.Go(func() {
			for rel := range jobs {
				ok, err := downloadImage(dir, rel)
				if err != nil {
					errOnce.Do(func() { firstErr = err })
					continue
				}
				if ok {
					downloaded.Add(1)
				}
			}
		})
	}
	for _, p := range paths {
		jobs <- p
	}
	close(jobs) // workers' range loops end once the channel drains
	wg.Wait()
	return int(downloaded.Load()), firstErr
}

// downloadImage saves imagesURL+rel to dir/rel unless it already exists.
// It writes to a temp file and renames, so a crash never leaves a partial
// photo that later runs would mistake for a complete one.
func downloadImage(dir, rel string) (bool, error) {
	if !filepath.IsLocal(rel) { // dataset paths must stay inside dir
		return false, fmt.Errorf("unsafe image path %q", rel)
	}
	dst := filepath.Join(dir, rel)
	if _, err := os.Stat(dst); err == nil {
		return false, nil
	}

	body, err := fetch(imagesURL + rel)
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return false, err
	}
	tmp := dst + ".tmp"
	if err := os.WriteFile(tmp, body, 0o644); err != nil {
		return false, err
	}
	return true, os.Rename(tmp, dst)
}

func fetch(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}
