package go_vanity_urls

import "html/template"

var indexTmpl = template.Must(template.New("index").Parse(`
<!DOCTYPE html>
<html>
	<body>
		<h1>{{.getHost}}</h1>
		<ul>
			{{range .Handlers}}<li><a href="https://pkg.go.dev/{{.}}">{{.}}</a></li>{{end}}
		</ul>
	</body>
</html>
`))

var vanityTmpl = template.Must(template.New("vanity").Parse(`
<!DOCTYPE html>
<html>
	<head>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
		<meta name="go-import" content="{{.Import}} git {{.Repo}}">
		<meta name="go-source" content="{{.Import}} {{.Display}}">
		<meta http-equiv="refresh" content="0; url=https://pkg.go.dev/{{.Import}}/{{.Subpath}}">
	</head>
	<body>
		Nothing to see here; <a href="https://pkg.go.dev/{{.Import}}/{{.Subpath}}">see the package on pkg.go.dev</a>.
	</body>
</html>
`))
