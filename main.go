package main

import (
	"flag"
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
	list.Height = 12
	list.Width = 100
	list.Y = 0
	//list.Overflow = "wrap"

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
	path = strings.TrimRight(path, "\n")
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

func getRepositories(url string) []string {
	doc, _ := goquery.NewDocument(url)

	selectorString := ".repo-list li"
	repositories := []string{}

	doc.Find(selectorString).Each(func(i int, s *goquery.Selection) {
		repoInfo := ""
		s.Find("div h3 a").Each(func(_ int, s *goquery.Selection) {
			repositoryURL, _ := s.Attr("href")
			repositoryURL = strings.Replace(repositoryURL, "/", "", 1)
			repositoryURL = strings.TrimRight(repositoryURL, "\n")
			repositoryURL = strings.TrimRight(repositoryURL, " ")
			repoInfo += repositoryURL
		})
		s.Find(".py-1 p").Each(func(_ int, s *goquery.Selection) {
			repositoryDescription, _ := s.Html()
			repositoryDescription = strings.TrimLeft(repositoryDescription, "\n")
			repositoryDescription = strings.TrimLeft(repositoryDescription, " ")
			repoInfo += " " + repositoryDescription
		})

		repositories = append(repositories, repoInfo)
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

func getURL(language string) string {
	path := ""
	if language != "" {
		path = "/" + language
	} else {
		path = ""
	}
	return "https://github.com/trending" + path
}

func main() {
	var language string

	flag.StringVar(&language, "language", "blank", "language flag")
	flag.StringVar(&language, "l", "blank", "language flag")
	flag.Parse()

	url := getURL(language)
	repositoryArray := getRepositories(url)
	formattedRepositories := formatOptions(repositoryArray)

	var menu Menu
	menu.init(formattedRepositories)
}
