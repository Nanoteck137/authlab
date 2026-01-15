package render

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/nanoteck137/authlab"
)

//go:embed templates
var embedFS embed.FS

type Data struct {
	Icon    string
	AppName string
	Header  template.HTML
	Content template.HTML
}

var templates = template.Must(template.New("index").ParseFS(embedFS, "templates/index.html"))

func RenderCallbackSuccess(w io.Writer) error {
	return templates.ExecuteTemplate(w, "base", Data{
		Icon:    "success",
		AppName: authlab.AppName,
		Header:  "Login Successful!",
		Content: template.HTML(fmt.Sprintf("You have been authenticated to <strong>%s</strong> successfully.<br>You can now close this tab.", authlab.AppName)),
	})
}

func RenderCallbackRequestExpired(w io.Writer) error {
	return templates.ExecuteTemplate(w, "base", Data{
		Icon:    "error",
		AppName: authlab.AppName,
		Header:  "Request Expired!",
		Content: template.HTML("This request is expired. Please retry.<br>You can now close this tab."),
	})
}

func RenderCallbackError(w io.Writer) error {
	return templates.ExecuteTemplate(w, "base", Data{
		Icon:    "error",
		AppName: authlab.AppName,
		Header:  "Error!",
		Content: template.HTML("An unknown error occurred. Please retry<br>You can now close this tab."),
	})
}
