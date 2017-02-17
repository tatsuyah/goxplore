package main

import (
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gizak/termui"
	"github.com/skratchdot/open-golang/open"
)

type Menu struct {
	currentIndex int
	list         *termui.List
	urlList      []string
}

func (m Menu) init(options []string) {
	m.currentIndex = 0
	m.list = createList("GitHub Trending [(choose number to open repo or press q to quit)](fg-blue)", options)

	err := termui.Init()
	if err != nil {
		panic(err)
	}
	defer termui.Close()

	termui.Render(m.list)
	termui.Handle("/sys/kbd/q", func(termui.Event) {
		termui.StopLoop()
	})

	setKeyEvents(options, m.list)

	termui.Loop()
}

func createList(label string, items []string) *termui.List {

	list := termui.NewList()
	list.Items = items
	list.ItemFgColor = termui.ColorYellow
	list.BorderLabel = label
	list.Height = len(items) + 2
	list.Width = 100
	list.Y = 0

	return list
}

func highlightRow(row string) string {
	if strings.HasPrefix(row, "[[") {
		return row
	}

	highlightedRow := "[" + row + "](fg-white,bg-red)"
	return highlightedRow
}

func deHighlightRow(row string) string {
	if strings.HasPrefix(row, "[[") {
		index := strings.Index(row, "](")
		if index != -1 {
			row = row[0:index]
		}
		row = strings.Replace(row, "[", "", 1)
	}
	return row
}

func deHighlightRows(rows []string) []string {
	var deHighlightedRows []string
	for i := 0; i < len(rows); i++ {
		deHighlightedRow := deHighlightRow(rows[i])
		deHighlightedRows = append(deHighlightedRows, deHighlightedRow)
	}
	return deHighlightedRows
}

func setKeyEvents(options []string, ls *termui.List) {
	for i := 0; i < len(options); i++ {
		keyString := "/sys/kbd/" + strconv.Itoa(i)
		indexString := strconv.Itoa(i)
		termui.Handle(keyString, func(termui.Event) {
			options = deHighlightRows(options)
			index, _ := strconv.Atoi(indexString)
			options[index] = highlightRow(options[index])
			ls := createList("GitHub Trending [(choose number to open repo or press q to quit)](fg-blue)", options)
			openRepo(options, index)
			termui.Render(ls)
		})
	}
}

func openRepo(options []string, index int) {
	elms := strings.Split(options[index], " ")
	path := elms[1]
	url := "https://github.com/" + path
	open.Run(url)
}

func formatOptions(options []string) []string {
	var formattedOptions []string
	for i := 0; i < len(options); i++ {
		formattedString := "[" + strconv.Itoa(i) + "] " + options[i]
		formattedOptions = append(formattedOptions, formattedString)
	}
	return formattedOptions
}

type Page struct {
	url     string
	Scraper func() [][]string
}

func (p *Page) Scrape() [][]string {
	return p.Scraper()
}

func ScrapeTranding() [][]string {
	url := getURL("daily")
	doc, _ := goquery.NewDocument(url)
	selectorString := makeSelectorString("#explore-trending")
	repositories := [][]string{}

	doc.Find(selectorString).Each(func(i int, s *goquery.Selection) {
		repo := []string{}
		s.Find(".repo-name").Each(func(_ int, s *goquery.Selection) {
			repositoryURL, _ := s.Attr("href")
			repositoryURL = strings.Replace(repositoryURL, "/", "", 1)
			repo = append(repo, repositoryURL)
		})
		s.Find(".repo-description").Each(func(_ int, s *goquery.Selection) {
			repositoryDescription, _ := s.Html()
			repo = append(repo, repositoryDescription)
		})

		repositories = append(repositories, repo)
	})

	return repositories
}

func flattenArray(twoDArray [][]string) []string {
	flattenedArray := []string{}
	for _, repo := range twoDArray {
		repoElement := repo[0] + " " + "\"" + repo[1] + "\""
		flattenedArray = append(flattenedArray, repoElement)
	}
	return flattenedArray
}

func makeSelectorString(category string) string {
	selectorString := category + " .collection-item"
	return selectorString
}

func getURL(span string) string {
	return "https://github.com/explore?since=" + span
}

func main() {
	page := Page{
		url:     getURL("daily"),
		Scraper: ScrapeTranding,
	}

	repositories := page.Scrape()
	flattenedArray := flattenArray(repositories)
	formattedRepositories := formatOptions(flattenedArray)

	var menu Menu
	menu.init(formattedRepositories)
}
