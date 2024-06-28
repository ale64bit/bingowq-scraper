package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

var playerRegexp = regexp.MustCompile(`(.*)\((.*)\)(.*)岁`)

var regions = map[int]string{
	1:   "China",         // "中国冠军",
	2:   "Japan",         // "日本冠军",
	3:   "Korea",         // "韩国冠军",
	98:  "Taiwan",        // "中国台湾冠军",
	99:  "International", // "世界冠军",
	100: "Other",         // "其他冠军",
}

func main() {
	collectTournaments()
}

func collectTournaments() {
	if err := os.MkdirAll("tournament", 0755); err != nil {
		log.Fatalf("creating tournament directory: %v", err)
	}

	f, err := os.Create("tournaments.md")
	if err != nil {
		log.Fatalf("creating tournament index file: %v", err)
	}
	defer f.Close()

	fmt.Fprintln(f, "# Tournaments")
	fmt.Fprintln(f)

	for _, id := range []int{1, 2, 3, 98, 99, 100} {
		regionName := regions[id]

		fmt.Fprintf(f, "## %s\n", regionName)
		fmt.Fprintln(f)

		url := fmt.Sprintf("http://bingoweiqi.com/bingoweiqi2012/inquiry_match_change_scale.php?scale_id=%d", id)
		resp, err := http.Post(url, "", strings.NewReader(""))
		if err != nil {
			log.Fatalf("reading: %v", err)
		}
		defer resp.Body.Close()

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		if err != nil {
			log.Fatalf("parsing: %v", err)
		}

		doc.Find("option").Each(func(i int, s *goquery.Selection) {
			val, ok := s.Attr("value")
			if !ok {
				return
			}
			if val == "0" {
				return
			}
			id := strings.TrimSpace(val)
			name := strings.TrimSpace(s.Text())
			collectTournamentInfo(id, name)
			fmt.Fprintf(f, "- [%s](%s)\n", name, fmt.Sprintf("tournament/%s.md", id))
			log.Printf("%s: %s: %s", regionName, id, name)
		})

		fmt.Fprintln(f)
	}
}

func collectTournamentInfo(id, name string) {
	url := fmt.Sprintf("http://bingoweiqi.com/bingoweiqi2012/match_sub.php?sub_id=1&match_idx=%s&ep_id=0", id)
	resp, err := http.Post(url, "", strings.NewReader(""))
	if err != nil {
		log.Fatalf("reading: %v", err)
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatalf("parsing: %v", err)
	}

	props := map[string]string{}
	doc.Find("div table").First().Find("tr").Each(func(i int, s *goquery.Selection) {
		key := s.Find(`td[width="20%"]`).Text()
		val := s.Find(`td[width="80%"]`).Text()
		if len(val) > 0 {
			props[key] = val
		}
	})

	type prevResult struct {
		Num      string `json:"num"`
		Year     string `json:"year"`
		Winner   player `json:"winner"`
		RunnerUp player `json:"runner_up"`
		Score    string `json:"score"`
	}

	var results []prevResult
	doc.Find("div div table tr").Each(func(i int, s *goquery.Selection) {
		var res prevResult
		s.Find("td").Each(func(i int, s *goquery.Selection) {
			switch i {
			case 0:
				res.Num = strings.Trim(s.Find("a").Text(), "<>")
			case 1:
				res.Year = strings.TrimSuffix(s.Text(), "年")
			case 2:
				res.Winner = parsePlayer(s)
			case 3:
				res.Score = s.Text()
			case 4:
				res.RunnerUp = parsePlayer(s)
			}
		})
		results = append(results, res)
	})

	f, err := os.Create(filepath.Join("tournament", fmt.Sprintf("%s.md", id)))
	if err != nil {
		log.Fatalf("creating tournament file: %v", err)
	}
	defer f.Close()

	fmt.Fprintln(f, "# "+name)
	fmt.Fprintln(f)

	keys := mapKeys(props)
	for i := range keys {
		keys[i] = tr(keys[i])
	}
	sort.Strings(keys)
	fmt.Fprintln(f, "## Properties")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "| Key | Value |")
	fmt.Fprintln(f, "| --- | ----- |")
	for _, k := range keys {
		fmt.Fprintf(f, "| %s | %s |\n", k, props[k])
	}
	fmt.Fprintln(f)

	fmt.Fprintln(f, "## Previous")
	fmt.Fprintln(f)
	fmt.Fprintln(f, "| # | Year | Winner | Score | Runner Up |")
	fmt.Fprintln(f, "| --- | --- | --- | --- | --- |")
	for _, res := range results {
		fmt.Fprintf(f, "| %s | %s | %s | %s | %s |\n", res.Num, res.Year, res.Winner, res.Score, res.RunnerUp)
	}
	fmt.Fprintln(f)
}

func parsePlayer(s *goquery.Selection) player {
	a := s.Find("a")
	txt := strings.TrimSpace(a.Text())

	id := ""
	onClick, exists := a.Attr("onclick")
	if exists {
		const prefix = "parent.go_to('player', 'player.php?sub_id=1&player_idx="
		const suffix = "');"
		if strings.HasPrefix(onClick, prefix) && strings.HasSuffix(onClick, suffix) {
			id = strings.TrimSuffix(strings.TrimPrefix(onClick, prefix), suffix)
		}
	}

	gs := playerRegexp.FindStringSubmatch(txt)
	if len(gs) == 4 {
		return player{
			ID:   id,
			Name: strings.TrimSpace(gs[1]),
			Rank: strings.TrimSpace(gs[2]),
			Age:  strings.TrimSpace(gs[3]),
		}
	}

	return player{ID: id, Name: txt}
}

func mapKeys[K comparable, V any](m map[K]V) []K {
	var ks []K
	for k := range m {
		ks = append(ks, k)
	}
	return ks
}

func tr(s string) string {
	return s
}
