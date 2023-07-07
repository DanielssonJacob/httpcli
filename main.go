package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"github.com/common-nighthawk/go-figure"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Request struct {
	url     string
	method  string
	headers map[string]string
	body    string
}

var app = tview.NewApplication()

var requestForm = tview.NewForm()
var menuGrid = tview.NewGrid().SetRows(0, 0).SetColumns(20, 0)
var responseView = tview.NewTextView().SetDynamicColors(true)
var errorView = tview.NewTextView().SetDynamicColors(true)
var titleView = tview.NewTextView().SetDynamicColors(true)
var infoView = tview.NewTextView().SetDynamicColors(true)
var menuList = tview.NewList().
	AddItem("Create Request", "", 'r', func() {
		app.SetFocus(requestForm)
	}).AddItem("Quit", "", 'q', func() {
	app.Stop()
})
var response *http.Response

func renderResponse() {
	menuGrid.RemoveItem(errorView)
	responseView.Clear()
	switch response.StatusCode {
	case 200, 201, 202, 203, 204, 205, 206:
		fmt.Fprintf(responseView, "[green]%s[white]\n", response.Status)
	case 300, 301, 302, 303, 304, 305, 306, 307, 308:
		fmt.Fprintf(responseView, "[yellow]%s[white]\n", response.Status)
	case 400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410:
		fmt.Fprintf(responseView, "[red]%s[white]\n", response.Status)
	case 500, 501, 502, 503, 504, 505:
		fmt.Fprintf(responseView, "[red]%s[white]\n", response.Status)
	}
	fmt.Fprint(responseView, "[orange]Headers:[white]\n")
	for key, value := range response.Header {
		fmt.Fprintf(responseView, "[white]%s: %s\n", key, value)
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		renderError(err)
		return
	}
	defer response.Body.Close()
	fmt.Fprint(responseView, "[orange]ResponseBody:[white]\n")
	fmt.Fprintf(responseView, "[white]%s\n", body)
}

func createRequestForm() {
	var request Request
	requestForm.
		AddInputField("URL", "https://", 60, nil, func(url string) {
			request.url = url
		}).
		AddDropDown("Method", []string{"GET", "POST", "PUT", "DELETE", "PATCH"}, 0, func(method string, methodIndex int) {
			request.method = method
		}).
		AddTextArea("Headers", "", 60, 3, 0, func(headers string) {
			request.headers = make(map[string]string)
			for _, header := range strings.Split(headers, "\n") {
				header = strings.TrimSpace(header)
				if header == "" {
					continue
				}
				parts := strings.Split(header, ":")
				if len(parts) != 2 {
					renderError(fmt.Errorf("invalid header: %s", header))
					return
				}
				request.headers[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}).
		AddTextArea("Body", "", 60, 3, 0, func(body string) {
			request.body = body
		}).
		AddButton("Send", func() {
			httpClient := http.Client{}
			req, err := http.NewRequest(request.method, request.url, nil)
			if err != nil {
				renderError(err)
				return
			}
			response, err = httpClient.Do(req)
			if err != nil {
				renderError(err)
				return
			}
			menuGrid.AddItem(responseView, 2, 3, 4, 1, 0, 0, true)
			renderResponse()
			app.SetFocus(responseView)

		})
	requestForm.SetBorder(true).
		SetTitle("Request").
		SetTitleAlign(tview.AlignLeft)
}

func renderError(err error) {
	errorView.Clear()
	fmt.Fprintf(errorView, "[red]%s[white]", err)
	app.SetFocus(errorView)
	menuGrid.AddItem(errorView, 2, 3, 4, 1, 0, 0, true)
}

func init() {
	httpCliTitle := figure.NewColorFigure("HTTPCLI", "", "green", true).String()
	responseView.SetBorder(true).SetTitle("Response").SetTitleAlign(tview.AlignLeft)
	errorView.SetBorder(true).SetTitle("Error").SetTitleAlign(tview.AlignLeft)
	menuList.SetBorder(true).SetTitle("Menu").SetTitleAlign(tview.AlignLeft)
	menuGrid.SetBorder(true).SetTitle("HTTPCLI").SetTitleAlign(tview.AlignCenter)
	createRequestForm()
	fmt.Fprint(titleView, httpCliTitle)
	fmt.Fprint(infoView, "[orange](Esc)[white] to return.\n")
	menuGrid.
		AddItem(titleView, 0, 0, 2, 2, 0, 0, false).
		AddItem(infoView, 0, 2, 2, 2, 0, 0, false).
		AddItem(menuList, 2, 0, 4, 2, 0, 0, true).
		AddItem(requestForm, 2, 2, 4, 2, 0, 0, false)
}

func main() {

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			menuGrid.RemoveItem(errorView)
			menuGrid.RemoveItem(responseView)
			if app.GetFocus() == responseView {
				app.SetFocus(requestForm)
				return event
			} else {
				app.SetFocus(menuList)
			} 

		}
		return event
	})
	if err := app.SetRoot(menuGrid, true).SetFocus(menuGrid).Run(); err != nil {
		panic(err)
	}
}
