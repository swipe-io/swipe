# {{.ServiceName}} : A short description of the service. <code>{{ .GIT.LastTag.Name }}</code>
A complete description of the service and what it does.

## Example

<code>
go run ./cmd/service
</code>

## Docs

ToDo.

## Contributing

ToDo.

## Contributors

ToDo.

## Author

ToDo.

## Changelog

ToDo.

## Versions

{{range $index, $tag := .GIT.Tags -}}
   {{if gt $index 0 -}}, {{end -}}
   [{{$tag.Name}}](https://{{$.RootPkgPath}}/tree/{{$tag.Name}})
{{end -}}
